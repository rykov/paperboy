package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Paperboy",
	Long:  `A longer description...`,
}

func SetVersion(ver, date string) {
	versionStr := fmt.Sprintf("v%s %s/%s (%s)", ver, runtime.GOOS, runtime.GOARCH, date)
	versionCmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Println("Paperboy Email Engine " + versionStr)
	}
}
