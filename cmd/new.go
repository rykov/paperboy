package cmd

import (
	"github.com/bep/inflect"
	"github.com/rykov/paperboy/config"
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

var (
	newProjectDirs = []string{"content", "layouts", "lists", "themes"}
)

const (
	contentTemplate = `---
subject: "{{ .Subject }}"
date: {{ .Date }}
---

Email content goes here (Markdown formatted)
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

1. Add campaign content with "{{ .Cmd }} new <FILENAME>.md"
2. Add a recipient list with "{{ .Cmd }} new list <FILENAME>.yaml"
3. Configure your SMTP server in config.toml
3. Send that campaign "{{ .Cmd }} send <CONTENT> <LIST>"

Visit https://www.paperboy.email/ to learn more.
`
)

// "init" alias for "new project"
func initCmd() *cobra.Command {
	cmd := newProjectCmd()
	cmd.Use = "init [path]"
	return cmd
}

// "new" parent command for creation
func newCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "new [path]",
		Short:   "Create new content for a campaign",
		Example: "paperboy new the-announcement.md",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}

			path := cfg.AppFs.ContentPath(args[0])
			return writeTemplate(cfg.AppFs, path, contentTemplate, map[string]string{
				"Date":    time.Now().Format(time.RFC3339),
				"Subject": pathToName(path),
			}, false)
		},
	}

	// Subcommand "project" to start a new project
	cmd.AddCommand(newProjectCmd())

	// Subcommand "list" to start a new recipient list
	cmd.AddCommand(&cobra.Command{
		Use:     "list [path]",
		Short:   "Create a new recipient list",
		Example: "paperboy new list in-the-know",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}

			path := cfg.AppFs.ListPath(args[0])
			return writeTemplate(cfg.AppFs, path, listTemplate, nil, false)
		},
	})

	return cmd
}

func newProjectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "project [path]",
		Short: "Create new project directory",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			path, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			// Initalize new configuration
			cfg := config.NewConfig(afero.NewOsFs())

			// Check for config to see if a project exists
			configPath := filepath.Join(path, "config.toml")
			if ok, _ := afero.Exists(cfg.AppFs, configPath); ok {
				return newUserError("%s already contains a project", path)
			}

			// Create project directories
			for _, dir := range newProjectDirs {
				if err := os.MkdirAll(filepath.Join(path, dir), 0755); err != nil {
					return err
				}
			}

			// Write basic configuration
			if err := writeTemplate(cfg.AppFs, configPath, configTemplate, nil, true); err != nil {
				return err
			}

			// Success message
			vars := map[string]string{"Path": path, "Cmd": filepath.Base(os.Args[0])}
			out, _ := renderTemplate(newProjectBanner, vars)
			fmt.Print(out)
			return nil
		},
	}
}

func pathToName(path string) string {
	name, ext := filepath.Base(path), filepath.Ext(path)
	return inflect.Humanize(strings.TrimSuffix(name, ext))
}

func writeTemplate(fs *config.Fs, path, content string, data interface{}, quiet bool) error {
	if ex, err := afero.Exists(fs, path); ex {
		return newUserError("%s already exists", path)
	} else if err != nil {
		return err
	}

	out, err := renderTemplate(content, data)
	if err != nil {
		return err
	}

	err = afero.WriteFile(fs, path, out.Bytes(), 0644)
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
