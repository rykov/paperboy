package cmd

import (
	"github.com/rykov/paperboy/mail"
	"github.com/spf13/cobra"

	"fmt"
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send campaign to recipients",
	Long:  `A longer description...`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := loadConfig(); err != nil {
			return err
		}

		if len(args) != 2 {
			return fmt.Errorf("Invalid arguments")
		}

		return mail.SendCampaign(args[0], args[1])
	},
}
