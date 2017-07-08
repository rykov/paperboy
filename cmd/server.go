package cmd

import (
	"github.com/rykov/paperboy/mail"
	"github.com/spf13/cobra"

	"bytes"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Launch a preview server for emails",
	Long:  `A longer description...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := loadConfig(); err != nil {
			return err
		}

		fmt.Println("Preview server listening at :8080 ... ")
		http.HandleFunc("/", serverPreview)
		return http.ListenAndServe(":8080", nil)
	},
}

func serverPreview(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<html><body>")
	defer fmt.Fprintf(w, "</body></html>")

	parts := strings.Split(r.URL.Path[1:], "/")
	if len(parts) != 2 {
		fmt.Fprintf(w, "Invalid URL")
		return
	}

	template, who := parts[0], parts[1]
	fmt.Fprintf(w, "<p>TEMPLATE: %s, RECIPIENTS: %s</p>", template, who)

	c, err := mail.LoadCampaign(template, who)
	if err != nil || len(c.Recipients) == 0 {
		fmt.Fprintf(w, "Invalid campaign: %s", err)
		return
	}

	fmt.Fprintf(w, "<pre>")
	msg, err := c.MessageFor(0)
	if err != nil {
		fmt.Fprintf(w, "Campaign error: %s", err)
		return
	}

	var buf bytes.Buffer
	msg.WriteTo(&buf)
	io.WriteString(w, html.EscapeString(buf.String()))
	fmt.Fprintf(w, "</pre>")
}
