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

var gdriveSyncFlags struct {
	concurrency int
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
			fmt.Printf("No items found in '%s'\n", u.FDebug(path))
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
		fmt.Printf("Starting upload of %s to %s...\n", u.FDebug(localPath), u.FDebug(driveFolder))

		if err := gdrive.UploadDriveItem(srv, localPath, driveFolder); err != nil {
			log.Fatal().Err(err).Msg("Failed to upload")
		}
		fmt.Printf("%s Upload completed successfully\n", u.FSuccess("✓"))
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
		fmt.Printf("Starting download of %s...\n", u.FDebug(drivePath))

		downloadedPath, err := gdrive.DownloadDriveItem(srv, drivePath, localPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to download")
		}
		if downloadedPath != "" {
			fmt.Printf("Successfully downloaded %s %s %s\n", u.FDebug(drivePath), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(downloadedPath))
		}
		fmt.Printf("%s Download completed successfully\n", u.FSuccess("✓"))
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
		fmt.Printf("Starting sync of %s to %s...\n", u.FDebug(localDir), u.FDebug(remotePath))
		if err := gdrive.SyncDriveDirectory(srv, localDir, remotePath, gdriveSyncFlags.concurrency); err != nil {
			log.Fatal().Err(err).Msg("Failed to sync")
		}
		fmt.Printf("%s Sync completed successfully\n", u.FSuccess("✓"))
	},
}

var gdriveIndexCmd = &cobra.Command{
	Use:   "index [path]",
	Short: "Index file metadata for fast searching",
	Run: func(cmd *cobra.Command, args []string) {
		path := "root"
		if len(args) > 0 {
			path, _ = gdrive.ResolvePath(args[0])
		}
		srv, err := gdrive.GetDriveService(gdriveFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Google Drive service")
		}
		if err := gdrive.GenerateIndex(srv, path); err != nil {
			log.Fatal().Err(err).Msg("Indexing failed")
		}
		u.PrintSuccess("Google Drive indexing completed.")
	},
}

var gdriveSearchFlags struct {
	excludeDirs  bool
	excludeFiles bool
}

var gdriveSearchCmd = &cobra.Command{
	Use:   "search <regex> [path]",
	Short: "Search indexed files using regex",
	Long: `Search across your indexed Google Drive files using regex.
Requires running 'anbu gdrive index' first.
If [path] is omitted, it defaults to the root of the index.
Shortcuts (e.g., %project%) are supported in the path.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regex := args[0]
		path := "root"
		if len(args) > 1 {
			path, _ = gdrive.ResolvePath(args[1])
		}
		u.PrintInfo(fmt.Sprintf("Searching Google Drive index in '%s'...", path))
		items, err := gdrive.SearchIndex(regex, path, gdriveSearchFlags.excludeDirs, gdriveSearchFlags.excludeFiles)
		if err != nil {
			log.Fatal().Err(err).Msg("Search failed")
		}
		table := u.NewTable([]string{"Type", "Path", "Size", "Modified"})
		for _, item := range items {
			typeStr := "F"
			if item.Type == "folder" {
				typeStr = "D"
			}
			sizeStr := "--"
			if item.Type == "file" {
				sizeStr = fmt.Sprintf("%d B", item.Size)
			}
			table.Rows = append(table.Rows, []string{
				typeStr,
				item.Path,
				sizeStr,
				item.ModifiedTime,
			})
		}
		if len(table.Rows) == 0 {
			u.PrintWarning("No matches found.")
			return
		}
		table.PrintTable(false)
	},
}

func init() {
	GDriveCmd.PersistentFlags().StringVarP(&gdriveFlags.credentialsFile, "credentials", "c", "", "Path to Google Drive credentials.json file (default ~/.anbu-gdrive-credentials.json)")

	GDriveCmd.AddCommand(gdriveListCmd)
	GDriveCmd.AddCommand(gdriveUploadCmd)
	GDriveCmd.AddCommand(gdriveDownloadCmd)
	GDriveCmd.AddCommand(gdriveSyncCmd)
	GDriveCmd.AddCommand(gdriveIndexCmd)
	GDriveCmd.AddCommand(gdriveSearchCmd)

	gdriveSyncCmd.Flags().IntVarP(&gdriveSyncFlags.concurrency, "concurrency", "t", 8, "Number of items to process concurrently")
	gdriveSearchCmd.Flags().BoolVar(&gdriveSearchFlags.excludeDirs, "exclude-dirs", false, "Exclude directories from results")
	gdriveSearchCmd.Flags().BoolVar(&gdriveSearchFlags.excludeFiles, "exclude-files", false, "Exclude files from results")
}
