package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-find-dupes",
	Short: "A tool to find and remove duplicate files from various sources.",
	Long:  `go-find-dupes is a command-line tool that scans a source for duplicate files based on size and hash, allowing users to review and delete duplicates.`,

	// Don't show usage help if there's an error in RunE etc
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(gdriveCmd)
	rootCmd.AddCommand(localCmd)
}
