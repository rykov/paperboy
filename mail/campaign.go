package mail

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/rykov/paperboy/parser"
	"github.com/chris-ramon/douceur/inliner"
	"github.com/ghodss/yaml"
	"github.com/go-gomail/gomail"
	"github.com/russross/blackfriday"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// Like "User-Agent"
const xMailer = "paperboy/0.1.0 (https://paperboy.email)"

// Sender configuration
// TODO: Move this into a global space
var Config *viper.Viper

// Context for email template
type tmplContext struct {
	User     map[string]interface{}
	Campaign map[string]interface{}

	// For layout rendering
	CssContent string
	Content    string
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
	var content bytes.Buffer
	w := c.Recipients[i]

	// Render template body with text/template
	ctx := &tmplContext{User: w, Campaign: c.EmailMeta}
	if err := c.tText.Execute(&content, ctx); err != nil {
		return err
	}

	// Until we support file <style/> tags, load CSS into a variable
	if cssFile := AppFs.layoutPath("_default.css"); AppFs.isFile(cssFile) {
		cssBytes, _ := afero.ReadFile(AppFs, cssFile)
		ctx.CssContent = string(cssBytes)
	}

	// Render plain content into a layout (no Markdown)
	tLayoutFile := AppFs.layoutPath("_default.text")
	plainBody, err := renderPlain(content.Bytes(), tLayoutFile, ctx)
	if err != nil {
		return err
	}

	// Render content through Markdown and into a layout
	hLayoutFile := AppFs.layoutPath("_default.html")
	htmlBody, err := renderHTML(content.Bytes(), hLayoutFile, ctx)
	if err != nil {
		return err
	}

	toEmail := cast.ToString(w["email"])
	toName := cast.ToString(w["username"])

	m.Reset() // Return to NewMessage state
	m.SetAddressHeader("To", toEmail, toName)
	m.SetHeader("From", cast.ToString(c.EmailMeta["from"]))
	m.SetHeader("Subject", cast.ToString(c.EmailMeta["subject"]))
	m.SetHeader("X-Mailer", xMailer)
	m.SetBody("text/plain", plainBody)
	m.AddAlternative("text/html", htmlBody)
	return nil
}

func LoadCampaign(tmplID, listID string) (*Campaign, error) {
	// Translate IDs to files
	tmplFile := AppFs.contentPath(tmplID + ".md")
	listFile := AppFs.listPath(listID + ".yml")

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
	raw, err := afero.ReadFile(AppFs, path)
	if err != nil {
		return nil, err
	}

	var out []map[string]interface{}
	return out, yaml.Unmarshal(raw, &out)
}

func parseTemplate(path string) (parser.Email, error) {
	fmt.Println("Loading template: ", path)
	file, err := AppFs.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	return parser.ReadFrom(file)
}

func renderPlain(body []byte, layoutPath string, ctx *tmplContext) (string, error) {
	return renderIntoLayout(body, layoutPath, []byte{}, ctx)
}

// TODO: Uses a common text/template renderer, should use html/template instead
func renderHTML(body []byte, layoutPath string, ctx *tmplContext) (string, error) {
	bodyMD := blackfriday.MarkdownCommon(body)
	defaultLayout := []byte("<html><body>{{ .Content }}</body></html>")
	html, err := renderIntoLayout(bodyMD, layoutPath, defaultLayout, ctx)
	if err != nil {
		return "", err
	}
	return inliner.Inline(html)
}

func renderIntoLayout(body []byte, layoutPath string, defaultLayout []byte, ctx *tmplContext) (string, error) {
	layout := defaultLayout
	var err error

	if AppFs.isFile(layoutPath) {
		layout, err = afero.ReadFile(AppFs, layoutPath)
		if err != nil {
			return "", err
		}
	} else if len(layout) == 0 {
		return string(body), nil
	}

	tmpl, err := template.New("layout").Parse(string(layout))
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	var layoutCtx tmplContext = *ctx
	layoutCtx.Content = string(body)
	err = tmpl.Execute(&out, layoutCtx)
	return out.String(), err
}
