package upstream

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"voltium/config"
)

func newFileUpstream(t *testing.T) (*fileUpstream, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("0123456789"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	u, err := newFile(config.Service{Upstream: "file://" + dir})
	if err != nil {
		t.Fatal(err)
	}
	return u, dir
}

func TestFileServeAndRange(t *testing.T) {
	u, _ := newFileUpstream(t)

	// Tam dosya
	req := httptest.NewRequest("GET", "/index.html", nil)
	rec := httptest.NewRecorder()
	u.Forward(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status=%d, want 200", rec.Code)
	}
	if rec.Body.String() != "0123456789" {
		t.Errorf("body=%q", rec.Body.String())
	}
	if rec.Header().Get("ETag") == "" {
		t.Errorf("ETag set edilmeliydi")
	}

	// Range isteği → 206 + kısmi içerik
	req = httptest.NewRequest("GET", "/index.html", nil)
	req.Header.Set("Range", "bytes=0-4")
	rec = httptest.NewRecorder()
	u.Forward(rec, req)
	if rec.Code != http.StatusPartialContent {
		t.Fatalf("range status=%d, want 206", rec.Code)
	}
	if rec.Body.String() != "01234" {
		t.Errorf("range body=%q, want 01234", rec.Body.String())
	}
}

func TestFileDirectoryListingKapali(t *testing.T) {
	u, _ := newFileUpstream(t)
	req := httptest.NewRequest("GET", "/sub", nil)
	rec := httptest.NewRecorder()
	u.Forward(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("dizin isteği status=%d, want 403", rec.Code)
	}
}

func TestFileNotFound(t *testing.T) {
	u, _ := newFileUpstream(t)
	req := httptest.NewRequest("GET", "/yok.html", nil)
	rec := httptest.NewRecorder()
	u.Forward(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status=%d, want 404", rec.Code)
	}
}

func TestFileTraversalReddedilir(t *testing.T) {
	u, _ := newFileUpstream(t)
	// Kök dışına çıkma denemesi Clean sonrası köke sabitlenir → dosya yok → 404 (dışarı sızmaz).
	req := httptest.NewRequest("GET", "/../../../etc/passwd", nil)
	rec := httptest.NewRecorder()
	u.Forward(rec, req)
	if rec.Code == 200 {
		t.Errorf("traversal 200 döndürmemeli, status=%d", rec.Code)
	}
}
