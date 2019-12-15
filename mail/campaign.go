package mail

import (
	"bytes"
	"fmt"
	html "html/template"
	"path/filepath"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/go-gomail/gomail"
	"github.com/jtacoma/uritemplates"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rykov/paperboy/parser"
	"github.com/spf13/afero"
	"github.com/spf13/cast"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	gmparser "github.com/yuin/goldmark/parser"
)

// Like "User-Agent"
const xMailer = "paperboy/0.1.0 (https://paperboy.email)"

// Context for template rendering
type tmplContext struct {
	Content html.HTML
	context
}

type Campaign struct {
	Recipients []*ctxRecipient
	EmailMeta  *ctxCampaign
	Email      parser.Email

	// For logging, etc
	ID string

	// Internal templates
	bodyTemplate           *template.Template
	unsubscribeURLTemplate *uritemplates.UriTemplate
}

func (c *Campaign) MessageFor(i int) (*gomail.Message, error) {
	m := gomail.NewMessage()
	return m, c.renderMessage(m, i)
}

func (c *Campaign) renderMessage(m *gomail.Message, i int) error {
	var content bytes.Buffer

	// Get template context
	ctx, err := c.templateContextFor(i)
	if err != nil {
		return err
	}

	// Render template body with text/template
	if err := c.bodyTemplate.Execute(&content, ctx); err != nil {
		return err
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

	toEmail := cast.ToString(ctx.Recipient.Email)
	toName := cast.ToString(ctx.Recipient.Name)

	m.Reset() // Return to NewMessage state
	m.SetAddressHeader("To", toEmail, toName)
	m.SetHeader("Subject", cast.ToString(ctx.Campaign.Subject))
	m.SetHeader("From", cast.ToString(ctx.Campaign.From))
	m.SetHeader("X-Mailer", xMailer)
	m.SetBody("text/plain", plainBody)
	m.AddAlternative("text/html", htmlBody)
	return nil
}

// Create template context for messages and layouts
func (c *Campaign) templateContextFor(i int) (*tmplContext, error) {
	ctx := context{
		Recipient: *c.Recipients[i],
		Campaign:  *c.EmailMeta,
		Address:   Config.Address,
	}

	// Populate UnsubscribeURL using uritemplates
	if t := c.unsubscribeURLTemplate; t != nil {
		uu, err := t.Expand(ctx.toFlatMap())
		if err != nil {
			return nil, err
		}
		ctx.UnsubscribeURL = uu
	}

	// Render template body with text/template
	return &tmplContext{context: ctx}, nil
}

func LoadCampaign(tmplID, listID string) (*Campaign, error) {
	// Translate IDs to files
	tmplFile := AppFs.findContentPath(tmplID)
	listFile := AppFs.findListPath(listID)

	// Load up template with frontmatter
	email, err := parseTemplate(tmplFile)
	if err != nil {
		return nil, err
	}

	// Read and cast frontmatter
	var fMeta ctxCampaign
	if meta, err := email.Metadata(); err == nil && meta != nil {
		metadata, _ := meta.(map[string]interface{})
		fMeta = newCampaign(metadata)
	}

	// Parse email template for processing
	tmpl, err := template.New(tmplID).Parse(string(email.Content()))
	if err != nil {
		return nil, err
	}

	// Load up recipient metadata
	who, err := parseRecipients(listFile)
	if err != nil {
		return nil, err
	}

	// Campaign ID
	id := filepath.Base(tmplID)
	if ext := filepath.Ext(id); ext != "" {
		id = id[0 : len(id)-len(ext)]
	}

	// Prepare URI template for UnsubscribeURL
	var unsubscribe *uritemplates.UriTemplate
	if uu := Config.UnsubscribeURL; uu != "" {
		unsubscribe, err = uritemplates.Parse(uu)
		if err != nil {
			return nil, err
		}
	}

	return &Campaign{
		Recipients: who,
		EmailMeta:  &fMeta,
		Email:      email,
		ID:         id,

		unsubscribeURLTemplate: unsubscribe,
		bodyTemplate:           tmpl,
	}, nil
}

func parseRecipients(path string) ([]*ctxRecipient, error) {
	fmt.Println("Loading recipients", path)
	raw, err := afero.ReadFile(AppFs, path)
	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	out := make([]*ctxRecipient, len(data))
	for i, rData := range data {
		r := newRecipient(rData)
		out[i] = &r
	}

	return out, nil
}

func parseTemplate(path string) (parser.Email, error) {
	fmt.Println("Loading template", path)
	file, err := AppFs.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	return parser.ReadFrom(file)
}

func renderPlain(body []byte, layoutPath string, ctx *tmplContext) (string, error) {
	layout, err := loadTemplate(layoutPath, "{{ .Content }}")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(filepath.Base(layoutPath)).Parse(layout)
	if err != nil {
		return "", err
	}

	// Strip all HTML from campaign
	body = bluemonday.StrictPolicy().SanitizeBytes(body)

	var out bytes.Buffer
	var layoutCtx tmplContext = *ctx
	layoutCtx.Content = html.HTML(body)
	err = tmpl.Execute(&out, layoutCtx)
	return out.String(), err
}

// TODO: Uses a common text/template renderer, should use html/template instead
func renderHTML(body []byte, layoutPath string, ctx *tmplContext) (string, error) {
	layout, err := loadTemplate(layoutPath, "<html><body>{{ .Content }}</body></html>")
	if err != nil {
		return "", err
	}

	tmpl, err := html.New(filepath.Base(layoutPath)).Parse(layout)
	if err != nil {
		return "", err
	}

	// Configure Goldmark to match GoHugo and GFM
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			extension.Typographer,
			extension.DefinitionList,
		),
		goldmark.WithParserOptions(
			gmparser.WithAttribute(),
			gmparser.WithAutoHeadingID(),
		),
	)

	// Render Markdown
	var buf bytes.Buffer
	if err := md.Convert(body, &buf); err != nil {
		return "", err
	}

	// Sanitize HTML, in case "unsafe" Markdown is enabled
	bodyHTML := bluemonday.UGCPolicy().SanitizeBytes(buf.Bytes())

	var out bytes.Buffer
	var layoutCtx tmplContext = *ctx
	layoutCtx.Content = html.HTML(bodyHTML)
	if err := tmpl.Execute(&out, layoutCtx); err != nil {
		return "", err
	}

	return inlineStylesheets(layoutPath, out.String())
}

func loadTemplate(path string, defaultTemplate string) (string, error) {
	if path == "" || !AppFs.isFile(path) {
		return defaultTemplate, nil
	}
	raw, err := afero.ReadFile(AppFs, path)
	return string(raw), err
}
