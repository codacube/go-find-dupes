package cmd

import (
	"fmt"

	"github.com/codacube/go-find-dupes/internal/drivers/local"
	"github.com/codacube/go-find-dupes/internal/processor"
	"github.com/codacube/go-find-dupes/internal/tui"

	"github.com/spf13/cobra"
)

var localCmd = &cobra.Command{
	Use:   "local [folder]",
	Short: "Scan local directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetFolder := args[0]

		fmt.Println("\n-=> Finding duplicates in local folder <=-")

		var autoSetDeletePath string
		cmd.Flags().StringVar(&autoSetDeletePath, "delete-path", "", "Path to a folder where duplicates will be set to delete by default")

		drive, err := local.NewClient(targetFolder)
		if err != nil {
			return fmt.Errorf("init failed: %w", err)
		}

		pipe := processor.NewDefaultPipeline(drive)

		if err != nil {
			return err
		}

		// Everything is valid, start the TUI
		tui.Start(drive, pipe, targetFolder)

		return nil
	},
}
