package cmd

import (
	"github.com/bep/inflect"
	"github.com/rykov/paperboy/mail"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

func init() {
	newCmd.AddCommand(newListCmd)
}

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
		subj := filepath.Base(path)
		subj = inflect.Humanize(strings.TrimSuffix(subj, filepath.Ext(subj)))

		return writeTemplate(path, contentTemplate, map[string]string{
			"Subject": subj, "Date": time.Now().Format(time.RFC3339),
		})
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
		return writeTemplate(path, listTemplate, nil)
	},
}

func writeTemplate(path, content string, data interface{}) error {
	if ex, err := afero.Exists(mail.AppFs, path); ex {
		return newUserError("%s already exists", path)
	} else if err != nil {
		return err
	}

	var out bytes.Buffer
	t, _ := template.New("template").Parse(content)
	if err := t.Execute(&out, data); err != nil {
		return err
	}

	err := afero.WriteFile(mail.AppFs, path, out.Bytes(), 0644)
	if err == nil {
		fmt.Println(path, "created")
	}

	return err
}
