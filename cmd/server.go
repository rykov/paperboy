package cmd

import (
	"github.com/rykov/paperboy/mail"
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
		if err := mail.LoadConfig(); err != nil {
			return err
		}
		return startAPIServer()
	},
}

func startAPIServer() error {
	fmt.Println("API server listening at :8080 ... ")
	http.Handle("/graphql", server.GraphQLHandler())
	return http.ListenAndServe(":8080", nil)
}
