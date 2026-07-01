package config

import (
	"fmt"
	"strings"
)

var validRedirectCodes = map[int]bool{301: true, 302: true, 307: true, 308: true}

// Validate, config'in Faz-1 için tutarlı olduğunu doğrular.
func Validate(cfg *AppConfig) error {
	if len(cfg.Proxy.Ports) == 0 {
		return fmt.Errorf("proxy.ports boş olamaz")
	}
	for _, p := range cfg.Proxy.Ports {
		if p < 1 || p > 65535 {
			return fmt.Errorf("geçersiz port: %d", p)
		}
	}

	seenHosts := map[string]string{} // host -> service id
	for _, proj := range cfg.Projects {
		for _, svc := range proj.Services {
			if err := validateService(proj.ID, svc); err != nil {
				return err
			}
			if prev, ok := seenHosts[svc.Host]; ok {
				return fmt.Errorf("host %q birden fazla serviste (%s, %s)", svc.Host, prev, svc.ID)
			}
			seenHosts[svc.Host] = svc.ID
		}
	}
	return nil
}

func validateService(projID string, svc Service) error {
	if svc.Host == "" {
		return fmt.Errorf("proje %s: servis %s host boş", projID, svc.ID)
	}
	hasUpstream := svc.Upstream != ""
	hasRedirect := svc.Redirect != ""
	switch {
	case hasUpstream && hasRedirect:
		return fmt.Errorf("servis %s: hem upstream hem redirect tanımlı", svc.ID)
	case !hasUpstream && !hasRedirect:
		return fmt.Errorf("servis %s: upstream veya redirect gerekli", svc.ID)
	}
	if hasRedirect {
		code := svc.RedirectCode
		if code == 0 {
			code = 301
		}
		if !validRedirectCodes[code] {
			return fmt.Errorf("servis %s: geçersiz redirectCode %d (301|302|307|308)", svc.ID, code)
		}
		return nil
	}
	// upstream şeması http/https/file olmalı
	u := svc.Upstream
	if !(strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "file://")) {
		return fmt.Errorf("servis %s: bilinmeyen upstream şeması: %s", svc.ID, u)
	}
	return nil
}
