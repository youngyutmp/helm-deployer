package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "unset"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Long:  "Show current application version",
	Run:   showVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func showVersion(_ *cobra.Command, _ []string) {
	fmt.Printf("Version: %s\n", version)
}
