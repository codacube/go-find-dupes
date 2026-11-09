package cmd

import (
	"fmt"

	"github.com/codacube/go-find-dupes/internal/drivers/gdrive"
	"github.com/codacube/go-find-dupes/internal/processor"
	"github.com/codacube/go-find-dupes/internal/tui"

	"github.com/spf13/cobra"
)

var gdriveCmd = &cobra.Command{
	Use:   "gdrive [folderId]",
	Short: "Scan Google Drive folder",
	//Args:  cobra.ExactArgs(1), // TODO Fix the args for this - look at the old code
	RunE: func(cmd *cobra.Command, args []string) error {

		fmt.Printf("\n-=> Finding duplicates in Google Drive <=-")
		fmt.Printf("\nAuthenticating and initializing Google Drive client...")

		credentialsFile, err := cmd.Flags().GetString("creds")
		if err != nil {
			return err
		}

		folderID, err := cmd.Flags().GetString("folder-id")
		if err != nil {
			return err
		}

		drive, err := gdrive.NewClient(credentialsFile)
		if err != nil {
			fmt.Printf("Error initializing GDrive: %v\n", err)
			return err
		}
		pipe := processor.NewDefaultPipeline(drive)

		if err != nil {
			return err
		}

		tui.Start(drive, pipe, folderID)

		return nil
	},
}

// var gdriveGetTokenCmd = &cobra.Command{
// 	Use:   "get-token",
// 	Short: "Get Google Drive API token",
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		fmt.Printf("\n-=> Getting Google Drive API token <=-")

// 		credentialsFile, err := cmd.Flags().GetString("creds")
// 		if err != nil {
// 			return err
// 		}

// 		_, err = gdrive.GetToken(credentialsFile)
// 		if err != nil {
// 			fmt.Printf("Error getting token: %v\n", err)
// 			return err
// 		}

// 		fmt.Printf("Token saved to token.json\n")
// 		return nil
// 	},
// }

func init() {
	// TODO What is persistent flags vs Flags?
	gdriveCmd.PersistentFlags().StringP("creds", "c", "credentials.json", "Path to Google credentials file")
	gdriveCmd.Flags().StringP("folder-id", "f", "", "Google Drive Folder ID to scan for duplicates")
	// gdriveCmd.AddCommand(gdriveGetTokenCmd)
}
