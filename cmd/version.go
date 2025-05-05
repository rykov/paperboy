package cmd

import (
	"github.com/rykov/paperboy/config"
	"github.com/spf13/cobra"

	"fmt"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Paperboy",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, _ := config.LoadConfig()
		fmt.Printf("Paperboy Email Engine %s\n", cfg.Build)
	},
}
