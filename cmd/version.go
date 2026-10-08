package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Versiyon bilgisini gösterir",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Voltium v%s\n", Version)
	},
}
