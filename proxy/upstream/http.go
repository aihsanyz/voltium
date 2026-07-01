package upstream

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"voltium/config"
)

// httpUpstream, httputil.ReverseProxy sarmalayıcısıdır. Streaming ve hop-by-hop
// header temizliği ReverseProxy'den gelir; X-Forwarded-For otomatik eklenir.
type httpUpstream struct {
	proxy *httputil.ReverseProxy
	addr  string
}

func newHTTP(svc config.Service) (*httpUpstream, error) {
	target, err := url.Parse(svc.Upstream)
	if err != nil {
		return nil, err
	}
	timeout := time.Duration(svc.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	rp := httputil.NewSingleHostReverseProxy(target)
	orig := rp.Director
	rp.Director = func(req *http.Request) {
		originalHost := req.Host
		orig(req) // scheme/host/path'i hedefe göre ayarlar
		req.Header.Set("X-Forwarded-Host", originalHost)
		if ip := clientHost(req.RemoteAddr); ip != "" {
			req.Header.Set("X-Real-IP", ip)
		}
	}
	rp.Transport = &http.Transport{
		DialContext:           (&net.Dialer{Timeout: timeout}).DialContext,
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: time.Second,
	}
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		if isTimeout(err) {
			w.WriteHeader(http.StatusGatewayTimeout) // 504
			return
		}
		w.WriteHeader(http.StatusBadGateway) // 502
	}

	return &httpUpstream{proxy: rp, addr: target.Host}, nil
}

func (h *httpUpstream) Forward(w http.ResponseWriter, r *http.Request) {
	h.proxy.ServeHTTP(w, r)
}

func (h *httpUpstream) Addr() string { return h.addr }
func (h *httpUpstream) Type() string { return "http" }

// isTimeout, hatanın bir timeout (504) mı yoksa başka bir transport hatası (502) mı
// olduğunu ayırır.
func isTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var ne net.Error
	if errors.As(err, &ne) {
		return ne.Timeout()
	}
	return false
}

func clientHost(remoteAddr string) string {
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return remoteAddr
}
