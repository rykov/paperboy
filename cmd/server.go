package cmd

import (
	"github.com/rykov/paperboy/config"
	"github.com/rykov/paperboy/server"
	"github.com/rykov/paperboy/ui"
	"github.com/spf13/cobra"

	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"time"
)

const (
	// Local server configuration
	serverGraphQLPath = "/graphql"
)

func serverCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Launch a preview server for emails",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(cmd.Context())
			if err != nil {
				return err
			}
			return startAPIServer(cfg, nil)
		},
	}
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

	// The rest is handled by UI
	mux.Handle("/", uiHandler())

	// Initialize server with standard middleware
	s := &http.Server{Handler: server.WithMiddleware(mux, cfg)}
	s.Addr = fmt.Sprintf(":%d", cfg.ServerPort)

	// Request base context to gracefully kill connections
	s.BaseContext = func(_ net.Listener) context.Context {
		return cfg.Context
	}

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

	// Start server in goroutine with error handling
	errChan := make(chan error, 1)
	go func() {
		fmt.Printf("API server listening at %s ... \n", s.Addr)
		if err := s.Serve(l); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	// Wait until shutdown
	select {
	case <-cfg.Context.Done():
		// Parent-triggered shutdown
	case err := <-errChan:
		if err != nil {
			fmt.Printf("Server failed to start: %v\n", err)
			return fmt.Errorf("server failed to start: %w", err)
		}
	}

	// Graceful shutdown with timeout
	doneCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("\nShutting down server...")
	return s.Shutdown(doneCtx)
}

// Handle paths for Browser UI
func uiHandler() http.Handler {
	httpFS := http.FS(ui.FS)
	handler := http.FileServer(httpFS)

	// All paths unrecognized by FS are rewritten to /index.html
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := httpFS.Open(r.URL.Path); errors.Is(err, fs.ErrNotExist) {
			r.URL.Path = "/" // Let the UI sort out the rest
		}
		handler.ServeHTTP(w, r)
	})
}
