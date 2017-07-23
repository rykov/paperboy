package cmd

import (
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
		if err := loadConfig(); err != nil {
			return err
		}

		server.AddGraphQLRoutes()
		fmt.Println("API server listening at :8080 ... ")
		return http.ListenAndServe(":8080", nil)
	},
}
