// Package upstream, bir servisin hedefini (http/https veya file://) temsil eder
// ve gelen isteği bu hedefe iletir.
package upstream

import (
	"fmt"
	"net/http"
	"strings"

	"voltium/config"
)

// Upstream, bir isteği hedefe ileten bileşendir.
type Upstream interface {
	// Forward, isteği hedefe iletir ve yanıtı w'ye yazar.
	Forward(w http.ResponseWriter, r *http.Request)
	// Addr, access log $upstream_addr alanı için hedefi döner (ör. "127.0.0.1:3000", "file").
	Addr() string
	// Type, upstream türünü döner ("http" | "file").
	Type() string
}

// New, servis tanımından uygun Upstream'i üretir (PRD §19.4).
func New(svc config.Service) (Upstream, error) {
	u := svc.Upstream
	switch {
	case strings.HasPrefix(u, "http://"), strings.HasPrefix(u, "https://"):
		return newHTTP(svc)
	case strings.HasPrefix(u, "file://"):
		return newFile(svc)
	default:
		return nil, fmt.Errorf("bilinmeyen upstream şeması: %s", u)
	}
}
