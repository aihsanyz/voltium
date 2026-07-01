package proxy

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"voltium/config"
)

func newHandler(t *testing.T, svcs ...config.Service) *Handler {
	t.Helper()
	cfg := config.DefaultConfig([]int{80})
	cfg.Projects = []config.Project{{ID: "p", Services: svcs}}
	r, err := BuildRouter(cfg)
	if err != nil {
		t.Fatalf("BuildRouter: %v", err)
	}
	return NewHandler(r, nil, nil)
}

func TestHandlerHTTPProxy(t *testing.T) {
	var gotXFH, gotXRealIP, gotXFF string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotXFH = r.Header.Get("X-Forwarded-Host")
		gotXRealIP = r.Header.Get("X-Real-IP")
		gotXFF = r.Header.Get("X-Forwarded-For")
		w.Header().Set("X-Upstream", "yes")
		w.WriteHeader(200)
		w.Write([]byte("hello"))
	}))
	defer upstream.Close()

	h := newHandler(t, config.Service{ID: "s", Host: "api.example.com", Upstream: upstream.URL})

	req := httptest.NewRequest("GET", "http://api.example.com/path", nil)
	req.Host = "api.example.com"
	req.RemoteAddr = "1.2.3.4:5555"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status=%d, want 200", rec.Code)
	}
	if rec.Body.String() != "hello" {
		t.Errorf("body=%q, want hello", rec.Body.String())
	}
	if rec.Header().Get("X-Upstream") != "yes" {
		t.Errorf("upstream header iletilmedi")
	}
	if gotXFH != "api.example.com" {
		t.Errorf("X-Forwarded-Host=%q, want api.example.com", gotXFH)
	}
	if gotXRealIP != "1.2.3.4" {
		t.Errorf("X-Real-IP=%q, want 1.2.3.4", gotXRealIP)
	}
	if gotXFF == "" {
		t.Errorf("X-Forwarded-For eklenmeliydi")
	}
}

func TestHandlerRedirect(t *testing.T) {
	h := newHandler(t, config.Service{ID: "s", Host: "www.example.com", Redirect: "https://example.com", RedirectCode: 301})

	req := httptest.NewRequest("GET", "http://www.example.com/", nil)
	req.Host = "www.example.com"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 301 {
		t.Fatalf("status=%d, want 301", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "https://example.com" {
		t.Errorf("Location=%q, want https://example.com", loc)
	}
}

func TestHandlerNotFound(t *testing.T) {
	h := newHandler(t, config.Service{ID: "s", Host: "api.example.com", Upstream: "http://localhost:3000"})

	req := httptest.NewRequest("GET", "http://other.com/", nil)
	req.Host = "other.com"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("status=%d, want 404", rec.Code)
	}
}

func TestHandlerUpstreamDown502(t *testing.T) {
	// Kapatılmış bir listener'ın adresi → connection refused → 502.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()

	h := newHandler(t, config.Service{ID: "s", Host: "api.example.com", Upstream: "http://" + addr, Timeout: 2})

	req := httptest.NewRequest("GET", "http://api.example.com/", nil)
	req.Host = "api.example.com"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("status=%d, want 502", rec.Code)
	}
}

func TestHandlerUpstreamTimeout504(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1500 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer slow.Close()

	h := newHandler(t, config.Service{ID: "s", Host: "api.example.com", Upstream: slow.URL, Timeout: 1})

	req := httptest.NewRequest("GET", "http://api.example.com/", nil)
	req.Host = "api.example.com"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Errorf("status=%d, want 504", rec.Code)
	}
}
