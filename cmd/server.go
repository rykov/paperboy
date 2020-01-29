package cmd

import (
	"github.com/rykov/paperboy/config"
	"github.com/rykov/paperboy/server"
	"github.com/spf13/cobra"

	"fmt"
	"net"
	"net/http"
)

const (
	// Local server configuration
	serverGraphQLPath = "/graphql"
	serverLocalPort   = 8080
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
		return startAPIServer(cfg, nil)
	},
}

// Function is called before booting the server to configure
// additional routes for mux, and to provide "ready" hooks
type configFunc func(*http.ServeMux, chan bool) error

func startAPIServer(cfg *config.AConfig, configFn configFunc) error {
	// Simple router, for now
	mux := http.NewServeMux()

	// GraphQL API is handled via API
	mux.Handle(serverGraphQLPath, server.GraphQLHandler(cfg))

	// Append additional routes (e.g. preview)
	var ready chan bool = nil
	if configFn != nil {
		ready = make(chan bool)
		if err := configFn(mux, ready); err != nil {
			return err
		}
	}

	// Initialize server
	s := &http.Server{Handler: mux}
	s.Addr = fmt.Sprintf(":%d", serverLocalPort)

	// Open port for listening
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	// Notify listeners if ready (see preview command)
	if ready != nil {
		ready <- true
		close(ready)
	}

	// Serve server API
	fmt.Printf("API server listening at %s ... \n", s.Addr)
	return s.Serve(l)
}
