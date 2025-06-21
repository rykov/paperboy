package cmd

import (
	"github.com/rykov/paperboy/config"
	"github.com/spf13/cobra"

	"fmt"
)

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func New(build config.BuildInfo) *cobra.Command {
	config.Build = build

	rootCmd := &cobra.Command{Use: "paperboy"}
	rootCmd.AddCommand(newCmd())
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(sendCmd())
	rootCmd.AddCommand(serverCmd())
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(previewCmd())

	var cfgFile string
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./config.yaml)")
	cobra.OnInitialize(func() {
		config.ViperConfigFile = cfgFile
	})

	return rootCmd
}

// Error helpers
func newUserError(msg string, a ...interface{}) error {
	return fmt.Errorf(msg, a...)
}
