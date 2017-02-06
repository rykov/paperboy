package main

import (
	"bytes"
	"fmt"
	"github.com/rykov/paperboy/parser"
	"github.com/ghodss/yaml"
	"github.com/go-gomail/gomail"
	"github.com/toorop/go-dkim"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

// Context for email template
type tmplContext struct {
	User     map[string]interface{}
	Campaign map[string]interface{}
}

func main() {
	if len(os.Args) != 3 {
		printUsageError(fmt.Errorf("Invalid arguments"))
		return
	}

	// Load up template with frontmatter
	email, err := parseTemplate(os.Args[1])
	if err != nil {
		printUsageError(err)
		return
	}

	// Read and cast frontmatter
	var fMeta map[string]interface{}
	if meta, err := email.Metadata(); err == nil && meta != nil {
		fMeta, _ = meta.(map[string]interface{})
	}

	// Parse email template for processing
	tmpl, err := template.New("email").Parse(string(email.Content()))
	if err != nil {
		printUsageError(err)
		return
	}

	// Load up recipient metadata
	who, err := parseRecipients(os.Args[2])
	if err != nil {
		printUsageError(err)
		return
	}

	// Dial up the sender
	sender, err := dialSMTPURL(os.Getenv("SMTP_URL"))
	if err != nil {
		printUsageError(err)
		return
	}

	// Load private key
	keyBytes, _ := ioutil.ReadFile("dkim.private")

	// Configure with DKIM signing
	defer sender.Close()
	dOpts := dkim.NewSigOptions()
	dOpts.PrivateKey = keyBytes
	dOpts.Domain = "gemfury.com"
	dOpts.Selector = "rails"
	dOpts.SignatureExpireIn = 3600
	dOpts.AddSignatureTimestamp = true
	dOpts.Canonicalization = "relaxed/relaxed"
	dOpts.Headers = []string{
		"Mime-Version", "To", "From", "Subject", "Reply-To",
		"Sender", "Content-Transfer-Encoding", "Content-Type",
	}

	// DKIM-signing sender
	sender = &dkimSendCloser{Options: dOpts, sc: sender}

	// Send emails
	m := gomail.NewMessage()
	for _, w := range who {
		var body bytes.Buffer
		ctx := &tmplContext{User: w, Campaign: fMeta}
		if err := tmpl.Execute(&body, ctx); err != nil {
			printUsageError(err)
			return
		}

		toEmail := w["email"].(string)
		toName, _ := w["username"].(string)
		m.SetAddressHeader("To", toEmail, toName)
		m.SetHeader("From", fMeta["from"].(string))
		m.SetHeader("Subject", fMeta["subject"].(string))
		m.SetBody("text/plain", body.String())

		fmt.Println("Sending email to ", m.GetHeader("To"))
		if os.Getenv("DRY_RUN") != "" {
			fmt.Println("---------")
			m.WriteTo(os.Stdout)
			fmt.Println("\n---------")
		} else if err := gomail.Send(sender, m); err != nil {
			fmt.Println("  Could not send email: ", err)
		}

		// Throttle to account for quotas, etc
		time.Sleep(200 * time.Millisecond)
	}
}

func dialSMTPURL(smtpURL string) (gomail.SendCloser, error) {
	// Dial to SMTP server (with SSL)
	surl, err := url.Parse(smtpURL)
	if err != nil {
		return nil, err
	}

	// Authentication
	user, pass := os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS")
	if auth := surl.User; auth != nil {
		pass, _ = auth.Password()
		user = auth.Username()
	}

	// TODO: Split & parse port from url.Host
	host, port := surl.Host, 465

	// Dial SMTP server
	d := gomail.NewDialer(host, port, user, pass)
	d.SSL = true // Force SSL (TODO: use schema)
	return d.Dial()
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

func printUsageError(err error) {
	base := filepath.Base(os.Args[0])
	fmt.Printf("USAGE: %s [template] [recipients]\n", base)
	fmt.Println("Error: ", err)
}

// ======= DKIM ========
type dkimSendCloser struct {
	Options dkim.SigOptions
	sc      gomail.SendCloser
}

func (d *dkimSendCloser) Send(from string, to []string, msg io.WriterTo) error {
	return d.sc.Send(from, to, dkimMessage{d.Options, msg})
}

func (d *dkimSendCloser) Close() error {
	return d.sc.Close()
}

type dkimMessage struct {
	options dkim.SigOptions
	msg     io.WriterTo
}

func (dm dkimMessage) WriteTo(w io.Writer) (n int64, err error) {
	var b bytes.Buffer
	if _, err := dm.msg.WriteTo(&b); err != nil {
		return 0, err
	}

	email := b.Bytes()
	if err := dkim.Sign(&email, dm.options); err != nil {
		return 0, err
	}

	return bytes.NewBuffer(email).WriteTo(w)
}
