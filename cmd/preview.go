package cmd

import (
	"github.com/rykov/paperboy/config"
	"github.com/spf13/cobra"

	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
)

const (
	// Default configuration location for Preview UI
	previewDefaultConfigFile = "https://www.paperboy.email/ui/server_config.json"

	// Request headers when requesting server info
	cliVersionHeader = "X-Paperboy-Version"
	cliUserAgent     = "Paperboy (%s)"
)

var previewCmd = &cobra.Command{
	Use:   "preview [content] [list]",
	Short: "Preview campaign in browser",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		// Start server, notifies channel when listening
		return startAPIServer(cfg, func(mux *http.ServeMux, serverReady chan bool) error {

			// Wait for server and open preview
			go func() {
				if r, _ := <-serverReady; r {
					openPreview(args[0], args[1])
				}
			}()

			return nil
		})
	},
}

// Configuration file (local or remote) and viper
var previewConfigFlag string

// Relevant for both "server" and "preview" commands
func init() {
	desc := "Path or URL of preview server config"
	previewCmd.PersistentFlags().StringVar(&previewConfigFlag, "previewConfig", "", desc)
}

func openPreview(content, list string) {
	// Root URL for preview and GraphQL server
	previewRoot := fmt.Sprintf("http://localhost:%d", serverLocalPort)
	previewPath := fmt.Sprintf("/preview/%s/%s", url.PathEscape(content), url.PathEscape(list))

	// Open preview URL on various platform
	url := previewRoot + previewPath
	var err error
	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		err = fmt.Errorf("Unsupported platform")
	}

	if err != nil {
		fmt.Printf("\nPlease open the browser to the following URL:\n%s\n\n", url)
	}
}
