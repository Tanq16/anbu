package interactionsCmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tanq16/anbu/internal/interactions/gdrive"
	u "github.com/tanq16/anbu/utils"
)

var gdriveFlags struct {
	credentialsFile string
}

var gdriveSyncFlags struct {
	concurrency int
	ignore      string
}

var GDriveCmd = &cobra.Command{
	Use:     "gdrive",
	Aliases: []string{"gd"},
	Short:   "Interact with Google Drive to list, upload, download, sync, and index & search files and folders",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if gdriveFlags.credentialsFile == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get user home directory: %w", err)
			}
			gdriveFlags.credentialsFile = filepath.Join(homeDir, ".anbu", "gdrive-credentials.json")
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
	Short:   "List files and folders in Google Drive (defaults to root 'My Drive' without a path)",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "root"
		if len(args) > 0 {
			path, _ = gdrive.ResolvePath(args[0])
		}

		srv, err := gdrive.GetDriveService(gdriveFlags.credentialsFile)
		if err != nil {
			u.PrintFatal("Failed to get Google Drive service", err)
		}

		folders, files, err := gdrive.ListDriveContents(srv, path)
		if err != nil {
			u.PrintFatal(fmt.Sprintf("Failed to list contents for '%s'", path), err)
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
			u.PrintGeneric(fmt.Sprintf("No items found in '%s'", u.FDebug(path)))
			return
		}

		table.PrintTable(false)
	},
}

var gdriveUploadCmd = &cobra.Command{
	Use:     "upload <local-path> [drive-folder]",
	Aliases: []string{"up"},
	Short:   "Upload a local file or folder to Google Drive (defaults to root 'My Drive' without a path)",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		localPath := args[0]
		driveFolder := "root"
		if len(args) > 1 {
			driveFolder, _ = gdrive.ResolvePath(args[1])
		}

		srv, err := gdrive.GetDriveService(gdriveFlags.credentialsFile)
		if err != nil {
			u.PrintFatal("Failed to get Google Drive service", err)
		}
		u.PrintGeneric(fmt.Sprintf("Starting upload of %s to %s", u.FDebug(localPath), u.FDebug(driveFolder)))

		if err := gdrive.UploadDriveItem(srv, localPath, driveFolder); err != nil {
			u.PrintFatal("Failed to upload", err)
		}
		u.PrintSuccess(fmt.Sprintf("%s Upload completed successfully", u.StyleSymbols["pass"]))
	},
}

var gdriveDownloadCmd = &cobra.Command{
	Use:     "download <drive-path> [local-path]",
	Aliases: []string{"dl"},
	Short:   "Download a file or folder from Google Drive (defaults to current directory without a path)",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		drivePath, _ := gdrive.ResolvePath(args[0])
		localPath := ""
		if len(args) > 1 {
			localPath = args[1]
		}
		srv, err := gdrive.GetDriveService(gdriveFlags.credentialsFile)
		if err != nil {
			u.PrintFatal("Failed to get Google Drive service", err)
		}
		u.PrintGeneric(fmt.Sprintf("Starting download of %s", u.FDebug(drivePath)))

		downloadedPath, err := gdrive.DownloadDriveItem(srv, drivePath, localPath)
		if err != nil {
			u.PrintFatal("Failed to download", err)
		}
		if downloadedPath != "" {
			u.PrintGeneric(fmt.Sprintf("Successfully downloaded %s %s %s", u.FDebug(drivePath), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(downloadedPath)))
		}
		u.PrintSuccess(fmt.Sprintf("%s Download completed successfully", u.StyleSymbols["pass"]))
	},
}

var gdriveSyncCmd = &cobra.Command{
	Use:   "sync <local-dir> <remote-dir>",
	Short: "Sync local directory with Google Drive remote directory",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		localDir := args[0]
		remotePath, _ := gdrive.ResolvePath(args[1])
		ignore := parseIgnoreArg(gdriveSyncFlags.ignore)
		srv, err := gdrive.GetDriveService(gdriveFlags.credentialsFile)
		if err != nil {
			u.PrintFatal("Failed to get Google Drive service", err)
		}
		u.PrintGeneric(fmt.Sprintf("Starting sync of %s to %s", u.FDebug(localDir), u.FDebug(remotePath)))
		if err := gdrive.SyncDriveDirectory(srv, localDir, remotePath, gdriveSyncFlags.concurrency, ignore); err != nil {
			u.PrintFatal("Failed to sync", err)
		}
		u.PrintSuccess(fmt.Sprintf("%s Sync completed successfully", u.StyleSymbols["pass"]))
	},
}

var gdriveIndexCmd = &cobra.Command{
	Use:   "index [path]",
	Short: "Index file metadata for fast searching (defaults to root 'My Drive' without a path)",
	Run: func(cmd *cobra.Command, args []string) {
		path := "root"
		if len(args) > 0 {
			path, _ = gdrive.ResolvePath(args[0])
		}
		srv, err := gdrive.GetDriveService(gdriveFlags.credentialsFile)
		if err != nil {
			u.PrintFatal("Failed to get Google Drive service", err)
		}
		if err := gdrive.GenerateIndex(srv, path); err != nil {
			u.PrintFatal("Indexing failed", err)
		}
		u.PrintSuccess("Google Drive indexing completed")
	},
}

var gdriveSearchFlags struct {
	excludeDirs  bool
	excludeFiles bool
}

var gdriveSearchCmd = &cobra.Command{
	Use:   "search <regex> [path]",
	Short: "Search indexed files using regex (defaults to full index without a path)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regex := args[0]
		path := "root"
		if len(args) > 1 {
			path, _ = gdrive.ResolvePath(args[1])
		}
		u.PrintInfo(fmt.Sprintf("Searching Google Drive index in '%s'", path))
		items, err := gdrive.SearchIndex(regex, path, gdriveSearchFlags.excludeDirs, gdriveSearchFlags.excludeFiles)
		if err != nil {
			u.PrintFatal("Search failed", err)
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
			u.PrintWarning("No matches found", nil)
			return
		}
		table.PrintTable(false)
	},
}

func init() {
	GDriveCmd.PersistentFlags().StringVarP(&gdriveFlags.credentialsFile, "credentials", "c", "", "Path to Google Drive credentials.json file (default ~/.anbu/gdrive-credentials.json)")

	GDriveCmd.AddCommand(gdriveListCmd)
	GDriveCmd.AddCommand(gdriveUploadCmd)
	GDriveCmd.AddCommand(gdriveDownloadCmd)
	GDriveCmd.AddCommand(gdriveSyncCmd)
	GDriveCmd.AddCommand(gdriveIndexCmd)
	GDriveCmd.AddCommand(gdriveSearchCmd)

	gdriveSyncCmd.Flags().IntVarP(&gdriveSyncFlags.concurrency, "concurrency", "t", 8, "Number of items to process concurrently")
	gdriveSyncCmd.Flags().StringVarP(&gdriveSyncFlags.ignore, "ignore", "i", "", "Comma-separated list of file or folder names to ignore")
	gdriveSearchCmd.Flags().BoolVar(&gdriveSearchFlags.excludeDirs, "exclude-dirs", false, "Exclude directories from results")
	gdriveSearchCmd.Flags().BoolVar(&gdriveSearchFlags.excludeFiles, "exclude-files", false, "Exclude files from results")
}

// helper function to parse ignore arg
// also used in Box sync command
func parseIgnoreArg(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}
