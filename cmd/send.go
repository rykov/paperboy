package cmd

import (
	"github.com/rykov/paperboy/mail"
	"github.com/spf13/cobra"

	"fmt"
	"os"
	"path/filepath"
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send campaign to recipients",
	Long:  `A longer description...`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			printUsageError(fmt.Errorf("Invalid arguments"))
			return
		}

		err := mail.SendCampaign(args[0], args[1])
		if err != nil {
			printUsageError(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(sendCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// {{.cmdName}}Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// {{.cmdName}}Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func printUsageError(err error) {
	base := filepath.Base(os.Args[0])
	fmt.Printf("USAGE: %s send [template] [recipients]\n", base)
	fmt.Println("Error: ", err)
}
