package cmd

import (
	"github.com/rykov/paperboy/client"
	"github.com/rykov/paperboy/config"
	"github.com/rykov/paperboy/mail"
	"github.com/spf13/cobra"

	"context"
	"fmt"
	"os"
	"os/signal"
)

func sendCmd() *cobra.Command {
	var serverURL string
	var recipientsFilter string

	cmd := &cobra.Command{
		Use:     "send [content] [list]",
		Short:   "Send campaign to recipients",
		Example: "paperboy send the-announcement in-the-know",
		Args:    cobra.ExactArgs(2),
		// Uncomment the following line if your bare application
		// has an action associated with it:
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig() // Validate project
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			if u := serverURL; u == "" {
				ctx = withSignalTrap(ctx)
				return mail.LoadAndSendCampaign(ctx, cfg, args[0], args[1], recipientsFilter)
			} else {
				return client.New(ctx, u).Send(client.SendArgs{
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
	cmd.Flags().StringVar(&recipientsFilter, "filter", "", "Recipients filter")

	return cmd
}

func withSignalTrap(cmdCtx context.Context) context.Context {
	ctx, cancel := context.WithCancel(cmdCtx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Printf("Stopping on %s\n", sig)
			cancel()
			return
		}
	}()

	return ctx
}
