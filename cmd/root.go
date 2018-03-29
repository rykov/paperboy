package cmd

import (
	"fmt"
	"os"

	"github.com/rykov/paperboy/mail"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
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
func Execute(build mail.BuildInfo) {
	mail.SetFs(afero.NewOsFs())
	mail.Config.Build = build
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
		mail.InitConfig(cfgFile)
	})
}

// Error helpers
func newUserError(msg string, a ...interface{}) error {
	return fmt.Errorf(msg, a...)
}
