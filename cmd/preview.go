package cmd

import (
	"github.com/rykov/paperboy/mail"
	"github.com/spf13/cobra"

	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"time"
)

const previewURL = "http://www.paperboy.email/preview/connect"

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Preview campaign in browser",
	Long:  `A longer description...`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := mail.LoadConfig(); err != nil {
			return err
		}

		if len(args) != 2 {
			return newUserError("Invalid arguments")
		}

		go func() {
			runtime.Gosched()
			time.Sleep(100 * time.Millisecond)
			openPreview(args[0], args[1])
		}()

		return startAPIServer()
	},
}

func openPreview(content, list string) {
	q := url.Values{}
	q.Set("uri", "http://localhost:8080/graphql")
	q.Set("content", content)
	q.Set("list", list)

	url := previewURL + "?" + q.Encode()
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
