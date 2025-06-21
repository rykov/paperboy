package cmd

import (
	"github.com/rykov/paperboy/config"
	"github.com/rykov/paperboy/mail"
	"github.com/spf13/cobra"

	"context"
	"fmt"
	"os"
	"os/signal"
)

func sendCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "send [content] [list]",
		Short:   "Send campaign to recipients",
		Example: "paperboy send the-announcement in-the-know",
		Args:    cobra.ExactArgs(2),
		// Uncomment the following line if your bare application
		// has an action associated with it:
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}

			ctx := withSignalTrap(cmd.Context())
			return mail.LoadAndSendCampaign(ctx, cfg, args[0], args[1])
		},
	}
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
