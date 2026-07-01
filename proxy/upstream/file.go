package upstream

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"voltium/config"
)

// fileUpstream, bir kök dizinden statik dosya servis eder (PRD §8.5).
// http.ServeContent Range/If-Range/MIME/Last-Modified'ı sağlar; ETag ModTime+Size'dan üretilir.
type fileUpstream struct {
	root string
}

func newFile(svc config.Service) (*fileUpstream, error) {
	p := strings.TrimPrefix(svc.Upstream, "file://")
	if p == "" {
		return nil, fmt.Errorf("file upstream yolu boş: %s", svc.Upstream)
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return nil, err
	}
	return &fileUpstream{root: abs}, nil
}

func (f *fileUpstream) Forward(w http.ResponseWriter, r *http.Request) {
	// Path traversal koruması: baştaki "/" ile mutlaklaştırıp Clean ile ".." temizle,
	// ardından kök prefix'ini doğrula.
	clean := filepath.Clean("/" + r.URL.Path)
	full := filepath.Join(f.root, filepath.FromSlash(clean))
	if full != f.root && !strings.HasPrefix(full, f.root+string(os.PathSeparator)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	info, err := os.Stat(full)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if info.IsDir() {
		// Directory listing varsayılan kapalı.
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	file, err := os.Open(full)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	w.Header().Set("ETag", fmt.Sprintf(`"%x-%x"`, info.ModTime().Unix(), info.Size()))
	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}

func (f *fileUpstream) Addr() string { return "file" }
func (f *fileUpstream) Type() string { return "file" }
