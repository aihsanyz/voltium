// Package cmd, Voltium'un CLI komutlarını (cobra) tanımlar.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version, binary sürümüdür.
const Version = "1.0.0"

var (
	flagPorts    []int
	flagDataPath string
	flagLogLevel string
)

var rootCmd = &cobra.Command{
	Use:           "voltium",
	Short:         "Voltium — Lua plugin destekli akıllı reverse proxy",
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	// Alt komut verilmezse proxy'yi başlatır (PRD §16).
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServe()
	},
}

func init() {
	// Flag'ler yalnız ilk başlatmada geçerlidir (PRD §5); sonraki başlatmalarda config.yaml kullanılır.
	rootCmd.PersistentFlags().IntSliceVar(&flagPorts, "port", []int{80}, "Proxy portu (birden fazla kullanılabilir)")
	rootCmd.PersistentFlags().StringVar(&flagDataPath, "data-path", "./voltium-data", "Config ve log dizini")
	rootCmd.PersistentFlags().StringVar(&flagLogLevel, "log-level", "info", "debug|info|warn|error")

	rootCmd.AddCommand(versionCmd, resetCmd)
}

// Execute, kök komutu çalıştırır ve hata olursa çıkış kodu 1 ile sonlanır.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "hata:", err)
		os.Exit(1)
	}
}
