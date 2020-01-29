package cmd

import (
	"github.com/rykov/paperboy/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
)

const (
	// Default configuration location for Preview UI
	previewDefaultConfigFile = "https://www.paperboy.email/ui/server_config.json"

	// Request headers when requesting server info
	cliVersionHeader = "X-Paperboy-Version"
	cliUserAgent     = "Paperboy (%s)"
)

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Preview campaign in browser",
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

		// Start server, notifies channel when listening
		return startAPIServer(cfg, func(mux *http.ServeMux, serverReady chan bool) error {
			// Preview Viper configuration
			pViper := viper.New()

			// Wait for server and open preview
			go func() {
				if r, _ := <-serverReady; r {
					openPreview(pViper, args[0], args[1])
				}
			}()

			// Add preview routes
			return addPreviewProxyToMux(pViper, mux, cfg)
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

func openPreview(previewViper *viper.Viper, content, list string) {
	// Root URL for preview and GraphQL server
	previewRoot := fmt.Sprintf("http://localhost:%d", serverLocalPort)
	previewPath := previewViper.GetString("paths.previewConnect")

	// Add server configuration params
	q := url.Values{}
	q.Set("uri", previewRoot+serverGraphQLPath)
	q.Set("content", content)
	q.Set("list", list)

	// Open preview URL on various platform
	url := previewRoot + previewPath + "?" + q.Encode()
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

func addPreviewProxyToMux(previewViper *viper.Viper, mux *http.ServeMux, cfg *config.AConfig) error {
	// Use default config location, if not specified
	previewConfig := previewConfigFlag
	if previewConfig == "" {
		previewConfig = previewDefaultConfigFile
	}

	// Fetch config locally or remotely
	rawConfig, err := fetchPreviewConfig(previewConfig, cfg)
	if err != nil {
		return err
	}

	// Parse config with Viper
	previewViper.SetConfigType("JSON")
	if err := previewViper.ReadConfig(bytes.NewBuffer(rawConfig)); err != nil {
		return err
	}

	// Parse URL for remote UI
	u, err := url.Parse(previewViper.GetString("uiURL"))
	if err != nil {
		return err
	}

	// The rest is proxied to preview server
	proxy := httputil.NewSingleHostReverseProxy(u)

	// Fix proxy to properly populate request
	director := proxy.Director
	proxy.Director = func(r *http.Request) {
		r.Host = u.Host
		director(r)
	}

	// Add preview proxy to mux
	mux.Handle("/", proxy)
	return nil
}

func fetchPreviewConfig(loc string, cfg *config.AConfig) ([]byte, error) {
	isRemote := strings.HasPrefix(loc, "https://") || strings.HasPrefix(loc, "http://")

	// Load from local path
	if !isRemote {
		return ioutil.ReadFile(loc)
	}

	// Fetch from remote location URL
	req, err := http.NewRequest("GET", loc, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", fmt.Sprintf(cliUserAgent, cfg.Build.Version))
	req.Header.Set(cliVersionHeader, cfg.Build.Version)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Configuration not found at %q", loc)
	}

	return ioutil.ReadAll(resp.Body)
}
