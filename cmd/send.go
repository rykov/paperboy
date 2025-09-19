package cmd

import (
	"github.com/rykov/paperboy/client"
	"github.com/rykov/paperboy/config"
	"github.com/rykov/paperboy/mail"
	"github.com/spf13/cobra"
)

func sendCmd() *cobra.Command {
	var serverURL string

	cmd := &cobra.Command{
		Use:     "send [content] [list]",
		Short:   "Send campaign to recipients",
		Example: "paperboy send the-announcement in-the-know",
		Args:    cobra.ExactArgs(2),
		// Uncomment the following line if your bare application
		// has an action associated with it:
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(cmd.Context()) // Validate project
			if err != nil {
				return err
			}

			if u := serverURL; u == "" {
				return mail.LoadAndSendCampaign(cfg, args[0], args[1])
			} else {
				return client.New(cmd.Context(), u).Send(client.SendArgs{
					ProjectPath:    ".", // TODO: configurable
					ProjectIgnores: cfg.ClientIgnores,
					Campaign:       args[0],
					List:           args[1],
				})
			}
		},
	}

	// Server to specify remote server
	cmd.Flags().StringVar(&serverURL, "server", "", "URL of server")

	return cmd
}
