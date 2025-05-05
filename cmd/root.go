package cmd

import (
	"fmt"
	"os"

	"github.com/rykov/paperboy/config"
	"github.com/spf13/cobra"
)

// Global configuration

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use: "paperboy",
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(build config.BuildInfo) {
	config.Build = build
	RootCmd.AddCommand(newCmd)
	RootCmd.AddCommand(sendCmd)
	RootCmd.AddCommand(serverCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(previewCmd)

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	var cfgFile string
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./config.yaml)")
	cobra.OnInitialize(func() {
		config.ViperConfigFile = cfgFile
	})
}

// Error helpers
func newUserError(msg string, a ...interface{}) error {
	return fmt.Errorf(msg, a...)
}
