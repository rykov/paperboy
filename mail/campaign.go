package mail

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
	"time"

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

func SendCampaign(tmplFile, recipientFile string) error {
	// Load up template with frontmatter
	email, err := parseTemplate(tmplFile)
	if err != nil {
		return err
	}

	// Read and cast frontmatter
	var fMeta map[string]interface{}
	if meta, err := email.Metadata(); err == nil && meta != nil {
		fMeta, _ = meta.(map[string]interface{})
	}

	// Parse email template for processing
	tmpl, err := template.New("email").Parse(string(email.Content()))
	if err != nil {
		return err
	}

	// Load up recipient metadata
	who, err := parseRecipients(recipientFile)
	if err != nil {
		return err
	}

	// Dial up the sender
	sender, err := configureSender()
	if err != nil {
		return err
	}
	defer sender.Close()

	// Send emails
	m := gomail.NewMessage()
	for _, w := range who {
		var body bytes.Buffer
		ctx := &tmplContext{User: w, Campaign: fMeta}
		if err := tmpl.Execute(&body, ctx); err != nil {
			return err
		}

		toEmail := cast.ToString(w["email"])
		toName := cast.ToString(w["username"])
		m.SetAddressHeader("To", toEmail, toName)
		m.SetHeader("From", cast.ToString(fMeta["from"]))
		m.SetHeader("Subject", cast.ToString(fMeta["subject"]))
		m.SetBody("text/plain", body.String())

		fmt.Println("Sending email to ", m.GetHeader("To"))
		if err := gomail.Send(sender, m); err != nil {
			fmt.Println("  Could not send email: ", err)
		}

		// Throttle to account for quotas, etc
		time.Sleep(200 * time.Millisecond)
	}

	return nil
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
