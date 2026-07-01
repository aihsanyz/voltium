package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load, verilen yoldan config.yaml okur ve parse eder.
func Load(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config parse: %w", err)
	}
	return &cfg, nil
}

// Save, config'i YAML olarak atomik biçimde diske yazar (temp dosya + rename).
func Save(path string, cfg *AppConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("config marshal: %w", err)
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".config-*.yaml.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // rename başarılıysa no-op

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// Exists, config dosyasının var olup olmadığını söyler.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
