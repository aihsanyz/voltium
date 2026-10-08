package upstream

import (
	"testing"

	"voltium/config"
)

func TestNewUpstreamFactory(t *testing.T) {
	tests := []struct {
		name     string
		upstream string
		wantType string
		wantErr  bool
	}{
		{"http", "http://localhost:3000", "http", false},
		{"https", "https://example.com", "http", false},
		{"file", "file:///var/www", "file", false},
		{"bilinmeyen", "ftp://x", "", true},
		{"boş", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := New(config.Service{ID: "s", Host: "x", Upstream: tt.upstream})
			if (err != nil) != tt.wantErr {
				t.Fatalf("New() err=%v, wantErr=%v", err, tt.wantErr)
			}
			if err == nil && u.Type() != tt.wantType {
				t.Errorf("Type()=%q, want %q", u.Type(), tt.wantType)
			}
		})
	}
}
