package proxy

import (
	"net"
	"strings"
	"sync"

	"voltium/config"
	"voltium/proxy/upstream"
)

// ServiceEntry, router'da bir Host için çözülmüş kuralı taşır: upstream veya redirect.
type ServiceEntry struct {
	ID           string
	Host         string
	Redirect     string
	RedirectCode int
	Upstream     upstream.Upstream
}

// IsRedirect, girdinin bir redirect kuralı olup olmadığını söyler.
func (e *ServiceEntry) IsRedirect() bool { return e.Redirect != "" }

// Router, Host başlığına göre servis çözer: tam eşleşme → wildcard (tek seviye) → yok.
type Router struct {
	mu       sync.RWMutex
	services map[string]*ServiceEntry
}

// BuildRouter, config'ten upstream'leri bir kez kurarak router oluşturur.
func BuildRouter(cfg *config.AppConfig) (*Router, error) {
	services := make(map[string]*ServiceEntry)
	for _, proj := range cfg.Projects {
		for _, svc := range proj.Services {
			entry := &ServiceEntry{ID: svc.ID, Host: svc.Host}
			if svc.IsRedirect() {
				entry.Redirect = svc.Redirect
				entry.RedirectCode = svc.RedirectCode
				if entry.RedirectCode == 0 {
					entry.RedirectCode = 301
				}
			} else {
				up, err := upstream.New(svc)
				if err != nil {
					return nil, err
				}
				entry.Upstream = up
			}
			services[svc.Host] = entry
		}
	}
	return &Router{services: services}, nil
}

// Lookup, Host için servis çözer. Tam eşleşme wildcard'a göre önceliklidir (PRD §8.1).
// Wildcard yalnız tek seviyedir: a.example.com → *.example.com (a.b.example.com → *.b.example.com).
func (r *Router) Lookup(host string) (*ServiceEntry, bool) {
	host = stripPort(host)

	r.mu.RLock()
	defer r.mu.RUnlock()

	if e, ok := r.services[host]; ok {
		return e, true
	}
	if i := strings.IndexByte(host, '.'); i >= 0 {
		wildcard := "*." + host[i+1:]
		if e, ok := r.services[wildcard]; ok {
			return e, true
		}
	}
	return nil, false
}

// stripPort, "example.com:8080" → "example.com". Port yoksa olduğu gibi döner.
func stripPort(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}
