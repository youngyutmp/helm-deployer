package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "unset"

var versionCmd = cobra.Command{
	Run:   showVersion,
	Use:   "version",
	Short: "Show version",
	Long:  "Show current application version",
}

func showVersion(cmd *cobra.Command, args []string) {
	fmt.Println(version)
}
