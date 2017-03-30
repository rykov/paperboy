package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/rykov/paperboy/mail"
	"github.com/rykov/paperboy/parser"
	"github.com/ghodss/yaml"
	"github.com/go-gomail/gomail"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
	var cfgFile string
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./config.yaml)")
	cobra.OnInitialize(func() {
		initConfig(cfgFile)
	})
}

// Global viper config
var Config *viper.Viper

// initConfig reads in config file and ENV variables if set.
func initConfig(cfgFile string) {
	v := viper.New()
	Config = v

	// From --config
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	}

	// Tie configuration to ENV
	v.SetEnvPrefix("fugo")
	v.AutomaticEnv()

	// Load project's config.*
	v.SetConfigName("config")
	v.AddConfigPath(".")

	// Find and read the config file
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Config file error: %s \n", err))
	}

	// Defaults
	v.SetDefault("smtpURL", "")
	v.SetDefault("smtpUser", "")
	v.SetDefault("smtpPass", "")
	v.SetDefault("dryRun", false)
}

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
	sender, err := dialSMTPURL(Config.GetString("smtpURL"))
	if err != nil {
		printUsageError(err)
		return
	}
	defer sender.Close()

	// DKIM-signing sender, if configuration is present
	if cfg := Config.GetStringMap("dkim"); len(cfg) > 0 {
		sender, err = mail.SendCloserWithDKIM(sender, cfg)
		if err != nil {
			printUsageError(err)
			return
		}
	}

	// Send emails
	m := gomail.NewMessage()
	for _, w := range who {
		var body bytes.Buffer
		ctx := &tmplContext{User: w, Campaign: fMeta}
		if err := tmpl.Execute(&body, ctx); err != nil {
			printUsageError(err)
			return
		}

		toEmail := cast.ToString(w["email"])
		toName := cast.ToString(w["username"])
		m.SetAddressHeader("To", toEmail, toName)
		m.SetHeader("From", cast.ToString(fMeta["from"]))
		m.SetHeader("Subject", cast.ToString(fMeta["subject"]))
		m.SetBody("text/plain", body.String())

		fmt.Println("Sending email to ", m.GetHeader("To"))
		if Config.GetBool("dryRun") {
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
	user, pass := Config.GetString("smtpUser"), Config.GetString("smtpPass")
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
