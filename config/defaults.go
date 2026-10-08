package config

// DefaultConfig, ilk çalıştırmada oluşturulan varsayılan config'i döner (PRD §5).
// ports boşsa 80 kullanılır. API varsayılan kapalıdır.
func DefaultConfig(ports []int) *AppConfig {
	if len(ports) == 0 {
		ports = []int{80}
	}
	return &AppConfig{
		API: APIConfig{
			Enabled: false,
			Port:    8080,
		},
		Proxy: ProxyConfig{
			Ports: ports,
		},
		Projects: []Project{},
	}
}
