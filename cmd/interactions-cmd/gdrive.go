package interactionsCmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tanq16/anbu/internal/interactions"
	u "github.com/tanq16/anbu/utils"
)

var gdriveFlags struct {
	credentialsFile string
}

var GDriveCmd = &cobra.Command{
	Use:     "gdrive",
	Aliases: []string{"gd"},
	Short:   "Interact with Google Drive (list, upload, download)",
	Long: `Provides commands to interact with Google Drive.
Requires a credentials.json file, which can be specified via a flag
or placed at ~/.anbu-gdrive-credentials.json.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if gdriveFlags.credentialsFile == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get user home directory: %w", err)
			}
			gdriveFlags.credentialsFile = filepath.Join(homeDir, ".anbu-gdrive-credentials.json")
		}
		if _, err := os.Stat(gdriveFlags.credentialsFile); os.IsNotExist(err) {
			return fmt.Errorf("credentials file not found at %s. Please provide one using the --credentials flag or place it at the default location", gdriveFlags.credentialsFile)
		}
		return nil
	},
}

var gdriveListCmd = &cobra.Command{
	Use:     "list [folder-name]",
	Aliases: []string{"ls"},
	Short:   "List files and folders in Google Drive",
	Long:    `Lists files and folders. If [folder-name] is provided, lists content of that folder. Otherwise, lists the root 'My Drive'.`,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		folderName := "root"
		if len(args) > 0 {
			folderName = args[0]
		}

		srv, err := interactions.GetDriveService(gdriveFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Google Drive service")
		}

		folders, files, err := interactions.ListDriveContents(srv, folderName)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to list contents for '%s'", folderName)
		}

		table := u.NewTable([]string{"Type", "Name", "Size", "Modified"})
		for _, f := range folders {
			table.Rows = append(table.Rows, []string{
				u.FInfo("D"),
				f.Name,
				"--",
				f.ModifiedTime,
			})
		}
		for _, f := range files {
			table.Rows = append(table.Rows, []string{
				u.FDebug("F"),
				f.Name,
				interactions.HumanReadableSize(f.Size),
				f.ModifiedTime,
			})
		}

		if len(table.Rows) == 0 {
			u.PrintInfo(fmt.Sprintf("No items found in '%s'", folderName))
			return
		}

		table.PrintTable(false)
	},
}

var gdriveUploadCmd = &cobra.Command{
	Use:     "upload <local-file> [drive-folder]",
	Aliases: []string{"up"},
	Short:   "Upload a local file to Google Drive",
	Long:    `Uploads a single local file. If [drive-folder] is provided, uploads to that folder. Otherwise, uploads to the root 'My Drive'.`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		localPath := args[0]
		driveFolder := "root"
		if len(args) > 1 {
			driveFolder = args[1]
		} else if len(args) > 2 {
			log.Fatal().Msg("Too many arguments. Please provide only the local file and optionally the drive folder.")
		}

		srv, err := interactions.GetDriveService(gdriveFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Google Drive service")
		}
		u.PrintInfo(fmt.Sprintf("Starting upload of %s to %s...", u.FDebug(localPath), u.FDebug(driveFolder)))

		driveFile, err := interactions.UploadFile(srv, localPath, driveFolder)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to upload file")
		}
		fmt.Printf("Successfully uploaded %s %s %s (ID: %s)\n", u.FDebug(localPath), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(driveFile.Name), u.FDebug(driveFile.Id))
	},
}

var gdriveDownloadCmd = &cobra.Command{
	Use:     "download <drive-path>",
	Aliases: []string{"dl"},
	Short:   "Download a file from Google Drive",
	Long:    `Downloads a file from Google Drive to the current directory. <drive-path> is the full path to the file (e.g., 'MyFolder/MyFile.txt').`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		drivePath := args[0]
		srv, err := interactions.GetDriveService(gdriveFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Google Drive service")
		}
		u.PrintInfo(fmt.Sprintf("Starting download of %s...", u.FDebug(drivePath)))

		downloadedPath, err := interactions.DownloadFile(srv, drivePath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to download file")
		}
		fmt.Printf("Successfully downloaded %s %s %s\n", u.FDebug(drivePath), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(downloadedPath))
	},
}

var gdriveUploadFolderCmd = &cobra.Command{
	Use:     "upload-folder <local-folder> [drive-folder]",
	Aliases: []string{"up-f"},
	Short:   "Upload a local folder recursively to Google Drive",
	Long:    `Uploads a local folder recursively. If [drive-folder] is provided, uploads into that folder. Otherwise, creates the new folder in the root 'My Drive'.`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		localPath := args[0]
		driveFolder := "root"
		if len(args) > 1 {
			driveFolder = args[1]
		} else if len(args) > 2 {
			log.Fatal().Msg("Too many arguments. Please provide only the local folder and optionally the drive folder.")
		}

		srv, err := interactions.GetDriveService(gdriveFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Google Drive service")
		}
		u.PrintInfo(fmt.Sprintf("Starting folder upload of %s to %s...", u.FDebug(localPath), u.FDebug(driveFolder)))

		if err := interactions.UploadFolder(srv, localPath, driveFolder); err != nil {
			log.Fatal().Err(err).Msg("Failed to upload folder")
		}
		u.PrintSuccess(fmt.Sprintf("Successfully uploaded folder %s", localPath))
	},
}

var gdriveDownloadFolderCmd = &cobra.Command{
	Use:     "download-folder <drive-path>",
	Aliases: []string{"dl-f"},
	Short:   "Download a folder recursively from Google Drive",
	Long:    `Downloads a folder recursively from Google Drive to the current directory. <drive-path> is the full path to the folder (e.g., 'MyFolder/MySubFolder').`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		drivePath := args[0]
		srv, err := interactions.GetDriveService(gdriveFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Google Drive service")
		}
		u.PrintInfo(fmt.Sprintf("Starting folder download of %s...", u.FDebug(drivePath)))

		if err := interactions.DownloadFolder(srv, drivePath); err != nil {
			log.Fatal().Err(err).Msg("Failed to download folder")
		}
		u.PrintSuccess(fmt.Sprintf("Successfully downloaded folder %s", drivePath))
	},
}

func init() {
	GDriveCmd.PersistentFlags().StringVarP(&gdriveFlags.credentialsFile, "credentials", "c", "", "Path to Google Drive credentials.json file (default ~/.anbu-gdrive-credentials.json)")

	GDriveCmd.AddCommand(gdriveListCmd)
	GDriveCmd.AddCommand(gdriveUploadCmd)
	GDriveCmd.AddCommand(gdriveDownloadCmd)
	GDriveCmd.AddCommand(gdriveUploadFolderCmd)
	GDriveCmd.AddCommand(gdriveDownloadFolderCmd)
}
