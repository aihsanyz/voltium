package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"voltium/config"
	"voltium/logger"
	"voltium/proxy"
)

// runServe, config'i yükler/oluşturur, log'ları açar, router+handler+server kurar
// ve sinyaller gelene kadar dinler. SIGINT/SIGTERM → graceful shutdown; SIGHUP → log reopen.
func runServe() error {
	dataPath := flagDataPath
	cfgPath := filepath.Join(dataPath, "config.yaml")
	logsDir := filepath.Join(dataPath, "logs")

	firstRun := !config.Exists(cfgPath)

	// Dizin yapısı (PRD §6)
	for _, dir := range []string{logsDir, filepath.Join(dataPath, "plugins")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	// Çift-çalıştırma koruması
	release, err := acquireLock(filepath.Join(dataPath, "voltium.lock"))
	if err != nil {
		return err
	}
	defer release()

	// Config: ilk çalıştırmada oluştur, sonra yükle (flag'ler yalnız ilk başlatmada)
	var cfg *config.AppConfig
	if firstRun {
		cfg = config.DefaultConfig(flagPorts)
		if err := config.Save(cfgPath, cfg); err != nil {
			return err
		}
	} else {
		cfg, err = config.Load(cfgPath)
		if err != nil {
			return err
		}
	}
	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("config geçersiz: %w", err)
	}

	// Log'lar (nginx formatı)
	access, err := logger.NewAccessLogger(filepath.Join(logsDir, "access.log"))
	if err != nil {
		return err
	}
	defer access.Close()
	errlog, err := logger.NewErrorLogger(filepath.Join(logsDir, "error.log"))
	if err != nil {
		return err
	}
	defer errlog.Close()

	// Router + handler + server
	router, err := proxy.BuildRouter(cfg)
	if err != nil {
		return err
	}
	server := proxy.NewServer(proxy.NewHandler(router, access, errlog))
	if err := server.Start(cfg.Proxy.Ports); err != nil {
		errlog.Errorf("başlatma hatası: %v", err)
		return err
	}

	printBanner(cfg, firstRun)

	// Sinyal döngüsü
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for sig := range sigCh {
		if sig == syscall.SIGHUP {
			_ = access.Reopen()
			_ = errlog.Reopen()
			errlog.Info("SIGHUP: log dosyaları yeniden açıldı")
			continue
		}
		fmt.Println("\nKapatılıyor...")
		server.Shutdown()
		return nil
	}
	return nil
}

func printBanner(cfg *config.AppConfig, firstRun bool) {
	fmt.Printf("Voltium v%s\n", Version)
	fmt.Printf("Proxy  → %s\n", portsList(cfg.Proxy.Ports))
	if firstRun {
		fmt.Println("API    → kapalı (voltium api enable ile aktif edebilirsiniz)")
	} else {
		fmt.Println("API    → kapalı")
	}
	fmt.Println()
	fmt.Println("Docs → https://voltium.dev/docs")
}

func portsList(ports []int) string {
	parts := make([]string, len(ports))
	for i, p := range ports {
		parts[i] = fmt.Sprintf(":%d", p)
	}
	return strings.Join(parts, ", ")
}
