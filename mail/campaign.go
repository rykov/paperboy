package mail

import (
	"bytes"
	"fmt"
	html "html/template"
	"io"
	"path/filepath"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/go-gomail/gomail"
	"github.com/jtacoma/uritemplates"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rykov/paperboy/config"
	"github.com/rykov/paperboy/parser"
	"github.com/spf13/afero"
	"github.com/spf13/cast"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	gmparser "github.com/yuin/goldmark/parser"
)

// Shared empty parameters
var emptyParams = map[string]interface{}{}

// Like "User-Agent"
const xMailer = "paperboy/0.1.0 (https://paperboy.email)"

// Context for template rendering
type tmplContext struct {
	Content html.HTML
	Subject string
	renderContext
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

	// Configuration for everything else
	Config *config.AConfig
}

func (c *Campaign) MessageFor(i int) (*gomail.Message, error) {
	m := gomail.NewMessage()
	return m, c.renderMessage(m, i)
}

func (c *Campaign) renderMessage(m *gomail.Message, i int) error {
	var content bytes.Buffer
	appFs := c.Config.AppFs

	// Get template context
	ctx, err := c.templateContextFor(i)
	if err != nil {
		return err
	}

	// Render subject first so it's available to templates
	subject := cast.ToString(ctx.Campaign.subject)
	ctx.Subject, err = renderSubject(subject, ctx)
	if err != nil {
		return err
	}

	// Render template body with text/template
	if err := c.bodyTemplate.Execute(&content, ctx); err != nil {
		return err
	}

	// Render plain content into a layout (no Markdown)
	tLayoutFile := appFs.LayoutPath("_default.text")
	plainBody, err := c.renderPlain(content.Bytes(), tLayoutFile, ctx)
	if err != nil {
		return err
	}

	// Render content through Markdown and into a layout
	hLayoutFile := appFs.LayoutPath("_default.html")
	htmlBody, err := c.renderHTML(content.Bytes(), hLayoutFile, ctx)
	if err != nil {
		return err
	}

	toEmail := cast.ToString(ctx.Recipient.Email)
	to := cast.ToString(ctx.Campaign.To)
	toName, err := renderRecName(to, ctx)
	if err != nil {
		return err
	}

	m.Reset() // Return to NewMessage state
	m.SetAddressHeader("To", toEmail, toName)
	m.SetHeader("Subject", cast.ToString(ctx.Subject))
	m.SetHeader("From", cast.ToString(ctx.Campaign.From))
	m.SetHeader("X-Mailer", xMailer)
	m.SetBody("text/plain", plainBody)
	m.AddAlternative("text/html", htmlBody)
	return nil
}

// Create template context for messages and layouts
func (c *Campaign) templateContextFor(i int) (*tmplContext, error) {
	ctx := renderContext{
		Recipient: *c.Recipients[i],
		Campaign:  *c.EmailMeta,
		Address:   c.Config.Address,
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
	return &tmplContext{renderContext: ctx}, nil
}

// Populate campaign and a receipient list into a Campaign object
func LoadCampaign(cfg *config.AConfig, tmplID, listID string) (*Campaign, error) {
	campaign, err := LoadContent(cfg, tmplID)
	if err != nil {
		return nil, err
	}

	// Load up recipient metadata
	listFile := cfg.AppFs.FindListPath(listID)
	who, err := parseRecipients(cfg.AppFs, listFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load campain's recipients: %w", err)
	}

	// Populate recipients, and fire!
	campaign.Recipients = who
	return campaign, nil
}

// Populate campaign content and metadata from templateID into Campaign object
func LoadContent(cfg *config.AConfig, tmplID string) (*Campaign, error) {
	// Load up template with frontmatter
	tmplFile := cfg.AppFs.FindContentPath(tmplID)
	email, err := parseTemplate(cfg.AppFs, tmplFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load campain's content: %w", err)
	}

	// Read and cast frontmatter
	var fMeta ctxCampaign
	if meta, err := email.Metadata(); err == nil && meta != nil {
		metadata, _ := meta.(map[string]interface{})
		fMeta = newCampaign(cfg, metadata)
	} else { // Just defaults
		fMeta = newCampaign(cfg, emptyParams)
	}

	// Parse email template for processing
	tmpl, err := template.New(tmplID).Parse(string(email.Content()))
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
	if uu := cfg.UnsubscribeURL; uu != "" {
		unsubscribe, err = uritemplates.Parse(uu)
		if err != nil {
			return nil, err
		}
	}

	return &Campaign{
		EmailMeta: &fMeta,
		Email:     email,

		ID:     id,
		Config: cfg,

		unsubscribeURLTemplate: unsubscribe,
		bodyTemplate:           tmpl,
	}, nil
}

func parseRecipients(appFs *config.Fs, path string) ([]*ctxRecipient, error) {
	fmt.Println("Loading recipients", path)
	raw, err := afero.ReadFile(appFs, path)
	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	return MapsToRecipients(data)
}

func MapsToRecipients(data []map[string]interface{}) ([]*ctxRecipient, error) {
	out := make([]*ctxRecipient, len(data))

	for i, rData := range data {
		r := newRecipient(rData)
		out[i] = &r
	}

	return out, nil
}

func parseTemplate(appFs *config.Fs, path string) (parser.Email, error) {
	fmt.Println("Loading template", path)
	file, err := appFs.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	return parser.ReadFrom(file)
}

func (c *Campaign) renderPlain(body []byte, layoutPath string, ctx *tmplContext) (string, error) {
	layout, err := loadTemplate(c.Config.AppFs, layoutPath, "{{ .Content }}")
	if err != nil {
		return "", err
	}

	// Parse template first to bail on errors, if broken
	tmpl, err := template.New(filepath.Base(layoutPath)).Parse(layout)
	if err != nil {
		return "", err
	}

	// For markdown, we use notty style with tweaks
	style := *styles.DefaultStyles["notty"]
	style.Link.BlockPrefix = "("
	style.Link.BlockSuffix = ")"

	// Render markdown to plain text
	r, err := glamour.NewTermRenderer(glamour.WithStyles(style), glamour.WithWordWrap(-1))
	if err != nil {
		return "", err
	}

	body, err = r.RenderBytes(body)
	if err != nil {
		return "", err
	}

	// Strip all HTML from campaign
	body = bluemonday.StrictPolicy().SanitizeBytes(body)

	// Apply text/template
	return executeTemplate(body, tmpl, ctx)
}

// TODO: Uses a common text/template renderer, should use html/template instead
func (c *Campaign) renderHTML(body []byte, layoutPath string, ctx *tmplContext) (string, error) {
	layout, err := loadTemplate(c.Config.AppFs, layoutPath, "<html><body>{{ .Content }}</body></html>")
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

	// Render inner template
	tmplOut, err := executeTemplate(bodyHTML, tmpl, ctx)
	if err != nil {
		return "", err
	}

	// Inline CSS into elements "style" attribute
	return c.inlineStylesheets(layoutPath, tmplOut)
}

func renderSubject(subject string, ctx *tmplContext) (string, error) {
	tmpl, err := template.New("subject").Parse(subject)
	if err != nil {
		return "", err
	}
	return executeTemplate(nil, tmpl, ctx)
}

func renderRecName(name string, ctx *tmplContext) (string, error) {
	tmpl, err := template.New("name").Parse(name)
	if err != nil {
		return "", err
	}
	return executeTemplate(nil, tmpl, ctx)
}

func loadTemplate(appFs *config.Fs, path string, defaultTemplate string) (string, error) {
	if path == "" || !appFs.IsFile(path) {
		return defaultTemplate, nil
	}
	raw, err := afero.ReadFile(appFs, path)
	return string(raw), err
}

type Template interface {
	Execute(wr io.Writer, data interface{}) error
}

// Execute template with common template context (works for text or HTML)
// Context is reused (serially) for all the templates, so please clean up
func executeTemplate(body []byte, tmpl Template, ctx *tmplContext) (string, error) {
	if body != nil {
		ctx.Content = html.HTML(body)
	}

	var out bytes.Buffer
	err := tmpl.Execute(&out, ctx)
	ctx.Content = html.HTML("")
	return out.String(), err
}
