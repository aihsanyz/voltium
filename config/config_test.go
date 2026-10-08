package config

import (
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	orig := DefaultConfig([]int{80, 443})
	orig.Projects = []Project{{
		ID:   "proj-1",
		Name: "E-ticaret",
		Services: []Service{
			{ID: "svc-1", Host: "api.example.com", Upstream: "http://localhost:3000", Timeout: 30},
			{ID: "svc-2", Host: "*.example.com", Upstream: "http://localhost:4000", Timeout: 30},
			{ID: "svc-3", Host: "www.example.com", Redirect: "https://example.com", RedirectCode: 301},
			{ID: "svc-4", Host: "static.example.com", Upstream: "file:///var/www/html", Timeout: 30},
		},
	}}

	if err := Save(path, orig); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(got.Proxy.Ports) != 2 || got.Proxy.Ports[0] != 80 || got.Proxy.Ports[1] != 443 {
		t.Errorf("ports roundtrip hatalı: %v", got.Proxy.Ports)
	}
	if len(got.Projects) != 1 || len(got.Projects[0].Services) != 4 {
		t.Fatalf("projects/services roundtrip hatalı: %+v", got.Projects)
	}
	svc := got.Projects[0].Services[2]
	if svc.Redirect != "https://example.com" || svc.RedirectCode != 301 {
		t.Errorf("redirect roundtrip hatalı: %+v", svc)
	}
	if got.API.Enabled {
		t.Errorf("API varsayılan kapalı olmalı")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *AppConfig
		wantErr bool
	}{
		{
			name: "geçerli",
			cfg: withServices(Service{ID: "s", Host: "a.com", Upstream: "http://localhost:3000"}),
		},
		{
			name:    "port yok",
			cfg:     &AppConfig{},
			wantErr: true,
		},
		{
			name:    "upstream ve redirect birlikte",
			cfg:     withServices(Service{ID: "s", Host: "a.com", Upstream: "http://x", Redirect: "http://y"}),
			wantErr: true,
		},
		{
			name:    "ikisi de yok",
			cfg:     withServices(Service{ID: "s", Host: "a.com"}),
			wantErr: true,
		},
		{
			name:    "host yok",
			cfg:     withServices(Service{ID: "s", Upstream: "http://x"}),
			wantErr: true,
		},
		{
			name:    "geçersiz redirect code",
			cfg:     withServices(Service{ID: "s", Host: "a.com", Redirect: "http://y", RedirectCode: 500}),
			wantErr: true,
		},
		{
			name:    "bilinmeyen şema",
			cfg:     withServices(Service{ID: "s", Host: "a.com", Upstream: "ftp://x"}),
			wantErr: true,
		},
		{
			name: "duplicate host",
			cfg: withServices(
				Service{ID: "s1", Host: "a.com", Upstream: "http://x"},
				Service{ID: "s2", Host: "a.com", Upstream: "http://y"},
			),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func withServices(svcs ...Service) *AppConfig {
	cfg := DefaultConfig([]int{80})
	cfg.Projects = []Project{{ID: "p", Services: svcs}}
	return cfg
}
