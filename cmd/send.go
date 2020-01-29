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

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send campaign to recipients",
	Long:  `A longer description...`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		if len(args) != 2 {
			return newUserError("Invalid arguments")
		}

		ctx := withSignalTrap(context.Background())
		return mail.LoadAndSendCampaign(ctx, cfg, args[0], args[1])
	},
}

func withSignalTrap(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

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
