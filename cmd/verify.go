package cmd

import (
	"github.com/rykov/paperboy/config"
	"github.com/rykov/paperboy/mail"
	"github.com/spf13/cobra"

	"fmt"
)

func verifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "verify [content] [list]",
		Short:   "Verify DKIM signatures in rendered emails",
		Example: "paperboy verify the-announcement customers",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(cmd.Context())
			if err != nil {
				return err
			}

			// Render and verify campaign
			err = mail.VerifyCampaign(cfg, args[0], args[1])
			if err != nil {
				return err
			}

			// No problems found during validation
			fmt.Fprintf(cmd.OutOrStdout(), "Success! No problems found.\n")
			return nil
		},
	}
}
