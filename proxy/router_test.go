package proxy

import (
	"testing"

	"voltium/config"
)

func testConfig() *config.AppConfig {
	cfg := config.DefaultConfig([]int{80})
	cfg.Projects = []config.Project{{
		ID: "p",
		Services: []config.Service{
			{ID: "exact", Host: "api.example.com", Upstream: "http://localhost:3000"},
			{ID: "wild", Host: "*.example.com", Upstream: "http://localhost:4000"},
			{ID: "redir", Host: "www.example.com", Redirect: "https://example.com", RedirectCode: 301},
		},
	}}
	return cfg
}

func TestRouterLookup(t *testing.T) {
	r, err := BuildRouter(testConfig())
	if err != nil {
		t.Fatalf("BuildRouter: %v", err)
	}

	tests := []struct {
		name     string
		host     string
		wantID   string
		wantOK   bool
		redirect bool
	}{
		{"tam eşleşme", "api.example.com", "exact", true, false},
		{"tam eşleşme wildcard'a önceliklidir", "api.example.com", "exact", true, false},
		{"wildcard", "blog.example.com", "wild", true, false},
		{"redirect", "www.example.com", "redir", true, true},
		{"port strip", "api.example.com:8080", "exact", true, false},
		{"eşleşme yok", "other.com", "", false, false},
		{"çok seviyeli wildcard tek seviyeye düşmez", "a.b.example.com", "", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, ok := r.Lookup(tt.host)
			if ok != tt.wantOK {
				t.Fatalf("Lookup(%q) ok=%v, want %v", tt.host, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if e.ID != tt.wantID {
				t.Errorf("Lookup(%q) id=%q, want %q", tt.host, e.ID, tt.wantID)
			}
			if e.IsRedirect() != tt.redirect {
				t.Errorf("Lookup(%q) redirect=%v, want %v", tt.host, e.IsRedirect(), tt.redirect)
			}
		})
	}
}

func TestBuildRouterUnknownScheme(t *testing.T) {
	cfg := config.DefaultConfig([]int{80})
	cfg.Projects = []config.Project{{ID: "p", Services: []config.Service{
		{ID: "bad", Host: "x.com", Upstream: "ftp://x"},
	}}}
	if _, err := BuildRouter(cfg); err == nil {
		t.Error("bilinmeyen şema için hata bekleniyordu")
	}
}
