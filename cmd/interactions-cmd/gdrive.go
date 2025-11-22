package interactionsCmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tanq16/anbu/internal/interactions/gdrive"
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
	Use:     "list [path]",
	Aliases: []string{"ls"},
	Short:   "List files and folders in Google Drive",
	Long:    `Lists files and folders. If [path] is provided and is a folder, lists its contents. If [path] is a file, shows file info. Otherwise, lists the root 'My Drive'.`,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "root"
		if len(args) > 0 {
			path, _ = gdrive.ResolvePath(args[0])
		}

		srv, err := gdrive.GetDriveService(gdriveFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Google Drive service")
		}

		folders, files, err := gdrive.ListDriveContents(srv, path)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to list contents for '%s'", path)
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
				gdrive.HumanReadableSize(f.Size),
				f.ModifiedTime,
			})
		}

		if len(table.Rows) == 0 {
			u.PrintInfo(fmt.Sprintf("No items found in '%s'", path))
			return
		}

		table.PrintTable(false)
	},
}

var gdriveUploadCmd = &cobra.Command{
	Use:     "upload <local-path> [drive-folder]",
	Aliases: []string{"up"},
	Short:   "Upload a local file or folder to Google Drive",
	Long:    `Uploads a local file or folder to Google Drive. If [drive-folder] is provided, uploads to that folder. Otherwise, uploads to the root 'My Drive'. Folders are uploaded recursively.`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		localPath := args[0]
		driveFolder := "root"
		if len(args) > 1 {
			driveFolder, _ = gdrive.ResolvePath(args[1])
		}

		srv, err := gdrive.GetDriveService(gdriveFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Google Drive service")
		}
		u.PrintInfo(fmt.Sprintf("Starting upload of %s to %s...", u.FDebug(localPath), u.FDebug(driveFolder)))

		if err := gdrive.UploadDriveItem(srv, localPath, driveFolder); err != nil {
			log.Fatal().Err(err).Msg("Failed to upload")
		}
	},
}

var gdriveDownloadCmd = &cobra.Command{
	Use:     "download <drive-path> [local-path]",
	Aliases: []string{"dl"},
	Short:   "Download a file or folder from Google Drive",
	Long:    `Downloads a file or folder from Google Drive. If [local-path] is provided for a file, saves to that path. For folders, downloads to current directory with the folder name.`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		drivePath, _ := gdrive.ResolvePath(args[0])
		localPath := ""
		if len(args) > 1 {
			localPath = args[1]
		}
		srv, err := gdrive.GetDriveService(gdriveFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Google Drive service")
		}
		u.PrintInfo(fmt.Sprintf("Starting download of %s...", u.FDebug(drivePath)))

		downloadedPath, err := gdrive.DownloadDriveItem(srv, drivePath, localPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to download")
		}
		if downloadedPath != "" {
			fmt.Printf("Successfully downloaded %s %s %s\n", u.FDebug(drivePath), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(downloadedPath))
		}
	},
}

var gdriveSyncCmd = &cobra.Command{
	Use:   "sync <local-dir> <remote-dir>",
	Short: "Sync local directory with Google Drive remote directory",
	Long:  `Synchronizes a local directory with a remote Google Drive directory. Uploads missing files, deletes remote-only files, and updates changed files.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		localDir := args[0]
		remotePath, _ := gdrive.ResolvePath(args[1])
		srv, err := gdrive.GetDriveService(gdriveFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Google Drive service")
		}
		u.PrintInfo(fmt.Sprintf("Starting sync of %s to %s...", u.FDebug(localDir), u.FDebug(remotePath)))
		if err := gdrive.SyncDriveDirectory(srv, localDir, remotePath); err != nil {
			log.Fatal().Err(err).Msg("Failed to sync")
		}
		u.PrintSuccess("Sync completed successfully")
	},
}

func init() {
	GDriveCmd.PersistentFlags().StringVarP(&gdriveFlags.credentialsFile, "credentials", "c", "", "Path to Google Drive credentials.json file (default ~/.anbu-gdrive-credentials.json)")

	GDriveCmd.AddCommand(gdriveListCmd)
	GDriveCmd.AddCommand(gdriveUploadCmd)
	GDriveCmd.AddCommand(gdriveDownloadCmd)
	GDriveCmd.AddCommand(gdriveSyncCmd)
}
