package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Config'i sıfırlar ve proxy'yi yeniden başlatır",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := filepath.Join(flagDataPath, "config.yaml")
		if err := os.Remove(cfgPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		fmt.Println("Config sıfırlandı.")
		return runServe()
	},
}
