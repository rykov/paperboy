package cmd

import (
	"github.com/rykov/paperboy/config"
	"github.com/rykov/paperboy/mail"
	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send campaign to recipients",
	Long:  `A longer description...`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.LoadConfig(); err != nil {
			return err
		}

		if len(args) != 2 {
			return newUserError("Invalid arguments")
		}

		return mail.SendCampaign(config.Config, args[0], args[1])
	},
}
