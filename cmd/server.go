package cmd

import (
	"github.com/rykov/paperboy/config"
	"github.com/rykov/paperboy/server"
	"github.com/spf13/cobra"

	"fmt"
	"net/http"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Launch a preview server for emails",
	Long:  `A longer description...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		return startAPIServer(cfg)
	},
}

func startAPIServer(cfg *config.AConfig) error {
	fmt.Println("API server listening at :8080 ... ")
	http.Handle("/graphql", server.GraphQLHandler(cfg))
	return http.ListenAndServe(":8080", nil)
}
