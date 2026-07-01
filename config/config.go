// Package config, Voltium'un tek config.yaml dosyasını okur/yazar ve thread-safe
// erişim için ConfigManager sağlar. Faz-1'de yalnız proxy/routing alanları kullanılır;
// api ve plugins alanları sonraki fazlarla YAML round-trip için struct'ta tutulur.
package config

import "sync"

// AppConfig, config.yaml'ın tamamını temsil eder (PRD §7).
type AppConfig struct {
	API      APIConfig     `yaml:"api"`
	Proxy    ProxyConfig   `yaml:"proxy"`
	Plugins  PluginsConfig `yaml:"plugins,omitempty"`
	Projects []Project     `yaml:"projects"`
}

// APIConfig — Faz-1'de kullanılmaz; yalnız round-trip için tutulur (PRD §11).
type APIConfig struct {
	Enabled        bool     `yaml:"enabled"`
	Port           int      `yaml:"port"`
	TokenHash      string   `yaml:"tokenHash,omitempty"`
	TokenCreatedAt string   `yaml:"tokenCreatedAt,omitempty"`
	AllowedIPs     []string `yaml:"allowedIPs,omitempty"`
}

// ProxyConfig — dinlenecek portlar.
type ProxyConfig struct {
	Ports []int `yaml:"ports"`
}

// PluginsConfig — Faz-2'ye kadar kullanılmaz.
type PluginsConfig struct {
	Global []PluginConfig `yaml:"global,omitempty"`
}

// PluginConfig — Faz-2'ye kadar kullanılmaz.
type PluginConfig struct {
	Name    string         `yaml:"name"`
	Path    string         `yaml:"path"`
	Order   int            `yaml:"order"`
	Enabled bool           `yaml:"enabled"`
	Timeout int            `yaml:"timeout"`
	Config  map[string]any `yaml:"config,omitempty"`
}

// Project, bir dizi servisi ve (Faz-2) plugin'i gruplar.
type Project struct {
	ID          string         `yaml:"id"`
	Name        string         `yaml:"name"`
	PluginChain PluginChain    `yaml:"pluginChain,omitempty"`
	Plugins     []PluginConfig `yaml:"plugins,omitempty"`
	Services    []Service      `yaml:"services"`
}

// PluginChain — Faz-2'ye kadar kullanılmaz.
type PluginChain struct {
	Mode  string `yaml:"mode,omitempty"`
	Order []any  `yaml:"order,omitempty"`
}

// Service, bir Host için routing kuralıdır: upstream (http/file) veya redirect.
type Service struct {
	ID           string `yaml:"id"`
	Host         string `yaml:"host"`
	Upstream     string `yaml:"upstream,omitempty"`
	Redirect     string `yaml:"redirect,omitempty"`
	RedirectCode int    `yaml:"redirectCode,omitempty"`
	Timeout      int    `yaml:"timeout,omitempty"`
}

// IsRedirect, servisin bir redirect kuralı olup olmadığını söyler.
func (s Service) IsRedirect() bool { return s.Redirect != "" }

// ConfigManager, AppConfig'e thread-safe erişim ve atomik kaydetme sağlar (PRD §19.1).
type ConfigManager struct {
	mu   sync.RWMutex
	cfg  *AppConfig
	path string
}

// NewManager, verilen config ve dosya yolu için bir yönetici döner.
func NewManager(cfg *AppConfig, path string) *ConfigManager {
	return &ConfigManager{cfg: cfg, path: path}
}

// Get, mevcut config'i döner. Dönen değer salt-okunur kabul edilmelidir;
// değişiklik için Update kullanın.
func (cm *ConfigManager) Get() *AppConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.cfg
}

// Path, config dosyasının yolunu döner.
func (cm *ConfigManager) Path() string { return cm.path }

// Update, config'i kilit altında değiştirir ve diske atomik yazar.
func (cm *ConfigManager) Update(fn func(*AppConfig)) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	fn(cm.cfg)
	return Save(cm.path, cm.cfg)
}
