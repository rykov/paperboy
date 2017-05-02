package cmd

import (
	"github.com/rykov/paperboy/mail"
	"github.com/spf13/cobra"

	"fmt"
	"net/http"
	"strings"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Launch a preview server for emails",
	Long:  `A longer description...`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Preview server listening at :8080 ... ")
		http.HandleFunc("/", serverPreview)
		http.ListenAndServe(":8080", nil)
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// {{.cmdName}}Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// {{.cmdName}}Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

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

	c, err := mail.LoadCampaign("_email/"+template+".md", "_email/"+who+".yml")
	if err != nil || len(c.Recipients) == 0 {
		fmt.Fprintf(w, "Invalid campaign: ", err)
		return
	}

	fmt.Fprintf(w, "<pre>")
	msg, err := c.MessageFor(0)
	if err != nil {
		fmt.Fprintf(w, "Campaign error: ", err)
		return
	}

	msg.WriteTo(w)
	fmt.Fprintf(w, "</pre>")
}
