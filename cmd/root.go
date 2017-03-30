package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/rykov/paperboy/parser"
	"github.com/ghodss/yaml"
	"github.com/go-gomail/gomail"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/toorop/go-dkim"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "fury-mail",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		sendCampaign(args)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.fury-mail.yaml)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".fury-mail") // name of config file (without extension)
	viper.AddConfigPath("$HOME")      // adding home directory as first search path
	viper.AutomaticEnv()              // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// Global viper config
var Config *viper.Viper

// Context for email template
type tmplContext struct {
	User     map[string]interface{}
	Campaign map[string]interface{}
}

func sendCampaign(args []string) {
	if len(args) != 2 {
		printUsageError(fmt.Errorf("Invalid arguments"))
		return
	}

	// Initialize configuration
	v := viper.New()
	Config = v

	// Tie configuration to ENV
	v.BindEnv("smtp_url", "SMTP_URL")
	v.BindEnv("smtp_user", "SMTP_USER")
	v.BindEnv("smtp_pass", "SMTP_PASS")
	v.BindEnv("dry_run", "DRY_RUN")

	// Override via config file
	//if configYAML != "" {
	//	v.SetConfigType("yaml")
	//	v.ReadConfig(strings.NewReader(configYAML))
	//}

	// Load up template with frontmatter
	email, err := parseTemplate(args[0])
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
	who, err := parseRecipients(args[1])
	if err != nil {
		printUsageError(err)
		return
	}

	// Dial up the sender
	sender, err := dialSMTPURL(Config.GetString("smtp_url"))
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
		if Config.GetBool("dry_run") {
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
	user, pass := Config.GetString("smtp_user"), Config.GetString("smtp_pass")
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
