package cmd

import (
	"github.com/bep/inflect"
	"github.com/rykov/paperboy/mail"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

func init() {
	// "init" alias for "new project"
	initCmd := *newProjectCmd
	initCmd.Use = "init [path]"
	RootCmd.AddCommand(&initCmd)

	// Subcommands for "new"
	newCmd.AddCommand(newProjectCmd)
	newCmd.AddCommand(newListCmd)
}

var (
	newProjectDirs = []string{"content", "layouts", "lists", "themes"}
)

const (
	contentTemplate = `---
subject: "{{ .Subject }}"
date: {{ .Date }}
---
`

	listTemplate = `---
- email: example@example.com
  name: Nick Example
`

	configTemplate = `# config.toml
# See https://www.paperboy.email/docs/configuration/
from = "Example <example@example.org>"
address = "Paperboy Inc, 123 Main St. New York, NY 10010, USA"
unsubscribeURL = "https://example.org/unsubscribe/{Recipient.Email}"

# SMTP Server
[smtp]
  url = "smtp://smtp.example.org"
`

	newProjectBanner = `Congratulations! Your new project is ready in {{ .Path }}

Your first campaign is only a few steps away:

1. Add campaign content with "paperboy new <FILENAME>.md"
2. Add a recipient list with "paperboy new list <FILENAME>.yaml"
3. Configure your SMTP server in config.toml
3. Send that campaign "paperboy send <CONTENT> <LIST>"

Visit https://www.paperboy.email/ to learn more.
`
)

var newCmd = &cobra.Command{
	Use:   "new [path]",
	Short: "Create new content for a campaign",
	Long:  `A longer description...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := loadConfig(); err != nil {
			return err
		}

		if len(args) < 1 {
			return newUserError("please provide a path")
		}

		path := mail.AppFs.ContentPath(args[0])
		return writeTemplate(path, contentTemplate, map[string]string{
			"Date":    time.Now().Format(time.RFC3339),
			"Subject": pathToName(path),
		}, false)
	},
}

var newListCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "Create a new recipient list",
	Long:  `A longer description...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := loadConfig(); err != nil {
			return err
		}

		if len(args) < 1 {
			return newUserError("please provide a path")
		}

		path := mail.AppFs.ListPath(args[0])
		return writeTemplate(path, listTemplate, nil, false)
	},
}

var newProjectCmd = &cobra.Command{
	Use:   "project [path]",
	Short: "Create new project directory",
	Long:  `A longer description...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		// Check for config to see if a project exists
		configPath := filepath.Join(path, "config.toml")
		if ok, _ := afero.Exists(mail.AppFs, configPath); ok {
			return newUserError("%s already contains a project", path)
		}

		// Create project directories
		for _, dir := range newProjectDirs {
			if err := os.MkdirAll(filepath.Join(path, dir), 0755); err != nil {
				return err
			}
		}

		// Write basic configuration
		if err := writeTemplate(configPath, configTemplate, nil, true); err != nil {
			return err
		}

		// Success message
		out, _ := renderTemplate(newProjectBanner, map[string]string{"Path": path})
		fmt.Print(out)
		return nil
	},
}

func pathToName(path string) string {
	name, ext := filepath.Base(path), filepath.Ext(path)
	return inflect.Humanize(strings.TrimSuffix(name, ext))
}

func writeTemplate(path, content string, data interface{}, quiet bool) error {
	if ex, err := afero.Exists(mail.AppFs, path); ex {
		return newUserError("%s already exists", path)
	} else if err != nil {
		return err
	}

	out, err := renderTemplate(content, data)
	if err != nil {
		return err
	}

	err = afero.WriteFile(mail.AppFs, path, out.Bytes(), 0644)
	if err == nil && !quiet {
		fmt.Println(path, "created")
	}

	return err
}

func renderTemplate(content string, data interface{}) (*bytes.Buffer, error) {
	var out bytes.Buffer
	t, _ := template.New("template").Parse(content)
	return &out, t.Execute(&out, data)
}
