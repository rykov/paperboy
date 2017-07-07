package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rykov/paperboy/mail"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "paperboy",
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
	RootCmd.AddCommand(serverCmd)
	RootCmd.AddCommand(sendCmd)
	RootCmd.AddCommand(newCmd)

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

// Configuration configuration :)
var viperConfig *viper.Viper

// initConfig will initialize the configuration
func initConfig(cfgFile string) {
	viperConfig = viper.New()
	v := viperConfig

	// From --config
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	}

	// Tie configuration to ENV
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("paperboy")
	v.AutomaticEnv()

	// Defaults (General)
	v.SetDefault("smtp.url", "")
	v.SetDefault("smtp.user", "")
	v.SetDefault("smtp.pass", "")
	v.SetDefault("dryRun", false)

	// Defaults (Dirs)
	v.SetDefault("contentDir", "content")
	v.SetDefault("layoutDir", "layouts")
	v.SetDefault("themeDir", "themes")
	v.SetDefault("listDir", "lists")

	// Prepare for project's config.*
	v.SetConfigName("config")
	v.AddConfigPath(".")

	// Wire up default afero.Fs
	mail.SetFs(afero.NewOsFs())
}

// Loading config separately allows us to switch up
// the underlying afero.Fs by calling mail.SetFs
func loadConfig() error {
	if err := viperConfig.ReadInConfig(); err != nil {
		return err
	}
	return viperConfig.Unmarshal(&mail.Config)
}

// Error helpers
func newUserError(msg string, a ...interface{}) error {
	return fmt.Errorf(msg, a...)
}
