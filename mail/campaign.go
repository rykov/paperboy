package mail

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/rykov/paperboy/parser"
	"github.com/ghodss/yaml"
	"github.com/go-gomail/gomail"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// Sender configurationTODO: Move this into a global space
var Config *viper.Viper

// Context for email template
type tmplContext struct {
	User     map[string]interface{}
	Campaign map[string]interface{}
}

type Campaign struct {
	Recipients []map[string]interface{}
	EmailMeta  map[string]interface{}
	Email      parser.Email

	// Internal templates
	tText *template.Template
}

func (c *Campaign) MessageFor(i int) (*gomail.Message, error) {
	m := gomail.NewMessage()
	return m, c.renderMessage(m, i)
}

func (c *Campaign) renderMessage(m *gomail.Message, i int) error {
	var body bytes.Buffer
	w := c.Recipients[i]

	ctx := &tmplContext{User: w, Campaign: c.EmailMeta}
	if err := c.tText.Execute(&body, ctx); err != nil {
		return err
	}

	toEmail := cast.ToString(w["email"])
	toName := cast.ToString(w["username"])

	m.Reset() // Return to NewMessage state
	m.SetAddressHeader("To", toEmail, toName)
	m.SetHeader("From", cast.ToString(c.EmailMeta["from"]))
	m.SetHeader("Subject", cast.ToString(c.EmailMeta["subject"]))
	m.SetBody("text/plain", body.String())
	return nil
}

func LoadCampaign(tmplID, listID string) (*Campaign, error) {
	// Translate IDs to files
	tmplFile := filepath.Join(Config.GetString("contentDir"), tmplID+".md")
	listFile := filepath.Join(Config.GetString("listDir"), listID+".yml")

	// Load up template with frontmatter
	email, err := parseTemplate(tmplFile)
	if err != nil {
		return nil, err
	}

	// Read and cast frontmatter
	var fMeta map[string]interface{}
	if meta, err := email.Metadata(); err == nil && meta != nil {
		fMeta, _ = meta.(map[string]interface{})
	}

	// Parse email template for processing
	tmpl, err := template.New("email").Parse(string(email.Content()))
	if err != nil {
		return nil, err
	}

	// Load up recipient metadata
	who, err := parseRecipients(listFile)
	if err != nil {
		return nil, err
	}

	return &Campaign{
		Recipients: who,
		EmailMeta:  fMeta,
		Email:      email,
		tText:      tmpl,
	}, nil
}

func parseRecipients(path string) ([]map[string]interface{}, error) {
	fmt.Println("Loading recipients: ", path)
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var out []map[string]interface{}
	return out, yaml.Unmarshal(raw, &out)
}

func parseTemplate(path string) (parser.Email, error) {
	fmt.Println("Loading template: ", path)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	return parser.ReadFrom(file)
}
