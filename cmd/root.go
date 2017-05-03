package cmd

import (
	"fmt"
	"os"

	"github.com/rykov/paperboy/mail"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "fury-mail",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	var cfgFile string
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./config.yaml)")
	cobra.OnInitialize(func() {
		initConfig(cfgFile)
	})
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cfgFile string) {
	v := viper.New()

	// From --config
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	}

	// Tie configuration to ENV
	v.SetEnvPrefix("fugo")
	v.AutomaticEnv()

	// Load project's config.*
	v.SetConfigName("config")
	v.AddConfigPath(".")

	// Find and read the config file
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Config file error: %s \n", err))
	}

	// Defaults (General)
	v.SetDefault("smtpURL", "")
	v.SetDefault("smtpUser", "")
	v.SetDefault("smtpPass", "")
	v.SetDefault("dryRun", false)

	// Defaults (Dirs)
	v.SetDefault("contentDir", "content")
	v.SetDefault("layoutDir", "layouts")
	v.SetDefault("listDir", "lists")

	// Wire everything up...
	mail.AppFs = afero.NewOsFs()
	mail.Config = v
}
