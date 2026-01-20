package interactionsCmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tanq16/anbu/internal/interactions/box"
	u "github.com/tanq16/anbu/utils"
)

var boxFlags struct {
	credentialsFile string
}

var boxSyncFlags struct {
	concurrency int
	ignore      string
}

var BoxCmd = &cobra.Command{
	Use:   "box",
	Short: "Interact with Box.com to list, upload, download, sync, and index & search files and folders",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if boxFlags.credentialsFile == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get user home directory: %w", err)
			}
			boxFlags.credentialsFile = filepath.Join(homeDir, ".anbu", "box-credentials.json")
		}
		if _, err := os.Stat(boxFlags.credentialsFile); os.IsNotExist(err) {
			return fmt.Errorf("credentials file not found at %s. Please provide one using the --credentials flag or place it at the default location", boxFlags.credentialsFile)
		}
		return nil
	},
}

var boxListCmd = &cobra.Command{
	Use:     "list [path]",
	Aliases: []string{"ls"},
	Short:   "List files and folders in Box (defaults to root folder without a path)",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := ""
		if len(args) > 0 {
			path, _ = box.ResolvePath(args[0])
		}
		client, err := box.GetBoxClient(boxFlags.credentialsFile)
		if err != nil {
			u.PrintFatal("Failed to get Box client", err)
		}
		folders, files, err := box.ListBoxContents(client, path)
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
				box.HumanReadableBoxSize(f.Size),
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

var boxUploadCmd = &cobra.Command{
	Use:     "upload <local-path> [box-folder-path]",
	Aliases: []string{"up"},
	Short:   "Upload a local file or folder to Box (defaults to root folder without a path)",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		localPath := args[0]
		boxFolderPath := ""
		if len(args) > 1 {
			boxFolderPath, _ = box.ResolvePath(args[1])
		}
		client, err := box.GetBoxClient(boxFlags.credentialsFile)
		if err != nil {
			u.PrintFatal("Failed to get Box client", err)
		}
		fmt.Printf("Starting upload of %s to %s...\n", u.FDebug(localPath), u.FDebug(boxFolderPath))
		if err := box.UploadBoxItem(client, localPath, boxFolderPath); err != nil {
			u.PrintFatal("Failed to upload", err)
		}
		u.PrintSuccess(fmt.Sprintf("%s Upload completed successfully", u.StyleSymbols["pass"]))
	},
}

var boxDownloadCmd = &cobra.Command{
	Use:     "download <box-path> [local-path]",
	Aliases: []string{"dl"},
	Short:   "Download a file or folder from Box (defaults to current directory without a path)",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		boxPath, _ := box.ResolvePath(args[0])
		localPath := ""
		if len(args) > 1 {
			localPath = args[1]
		}
		client, err := box.GetBoxClient(boxFlags.credentialsFile)
		if err != nil {
			u.PrintFatal("Failed to get Box client", err)
		}
		u.PrintGeneric(fmt.Sprintf("Starting download of %s", u.FDebug(boxPath)))
		downloadedPath, err := box.DownloadBoxItem(client, boxPath, localPath)
		if err != nil {
			u.PrintFatal("Failed to download", err)
		}
		if downloadedPath != "" {
			u.PrintGeneric(fmt.Sprintf("Successfully downloaded %s %s %s", u.FDebug(boxPath), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(downloadedPath)))
		}
		u.PrintSuccess(fmt.Sprintf("%s Download completed successfully", u.StyleSymbols["pass"]))
	},
}

var boxSyncCmd = &cobra.Command{
	Use:   "sync <local-dir> <remote-dir>",
	Short: "Sync local directory with Box remote directory with multi-threaded operations and ignore patterns",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		localDir := args[0]
		remotePath, _ := box.ResolvePath(args[1])
		ignore := parseIgnoreArg(boxSyncFlags.ignore)
		client, err := box.GetBoxClient(boxFlags.credentialsFile)
		if err != nil {
			u.PrintFatal("Failed to get Box client", err)
		}
		u.PrintGeneric(fmt.Sprintf("Starting sync of %s to %s", u.FDebug(localDir), u.FDebug(remotePath)))
		if err := box.SyncBoxDirectory(client, localDir, remotePath, boxSyncFlags.concurrency, ignore); err != nil {
			u.PrintFatal("Failed to sync", err)
		}
		u.PrintSuccess(fmt.Sprintf("%s Sync completed successfully", u.StyleSymbols["pass"]))
	},
}

var boxIndexCmd = &cobra.Command{
	Use:   "index [path]",
	Short: "Index file metadata for fast searching (defaults to root folder without a path)",
	Run: func(cmd *cobra.Command, args []string) {
		path := ""
		if len(args) > 0 {
			path, _ = box.ResolvePath(args[0])
		}
		client, err := box.GetBoxClient(boxFlags.credentialsFile)
		if err != nil {
			u.PrintFatal("Failed to get Box client", err)
		}
		if err := box.GenerateIndex(client, path); err != nil {
			u.PrintFatal("Indexing failed", err)
		}
		u.PrintSuccess("Box indexing completed")
	},
}

var boxSearchFlags struct {
	excludeDirs  bool
	excludeFiles bool
}

var boxSearchCmd = &cobra.Command{
	Use:   "search <regex> [path]",
	Short: "Search indexed files using regex (defaults to full index without a path)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regex := args[0]
		path := ""
		if len(args) > 1 {
			path, _ = box.ResolvePath(args[1])
		}
		displayPath := path
		if displayPath == "" {
			displayPath = "/"
		}
		u.PrintInfo(fmt.Sprintf("Searching Box index in '%s'", displayPath))
		items, err := box.SearchIndex(regex, path, boxSearchFlags.excludeDirs, boxSearchFlags.excludeFiles)
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
	BoxCmd.PersistentFlags().StringVarP(&boxFlags.credentialsFile, "credentials", "c", "", "Path to Box credentials.json file (default ~/.anbu/box-credentials.json)")

	BoxCmd.AddCommand(boxListCmd)
	BoxCmd.AddCommand(boxUploadCmd)
	BoxCmd.AddCommand(boxDownloadCmd)
	BoxCmd.AddCommand(boxSyncCmd)
	BoxCmd.AddCommand(boxIndexCmd)
	BoxCmd.AddCommand(boxSearchCmd)

	boxSyncCmd.Flags().IntVarP(&boxSyncFlags.concurrency, "concurrency", "t", 8, "Number of items to process concurrently")
	boxSyncCmd.Flags().StringVarP(&boxSyncFlags.ignore, "ignore", "i", "", "Comma-separated list of file or folder names to ignore")
	boxSearchCmd.Flags().BoolVar(&boxSearchFlags.excludeDirs, "exclude-dirs", false, "Exclude directories from results")
	boxSearchCmd.Flags().BoolVar(&boxSearchFlags.excludeFiles, "exclude-files", false, "Exclude files from results")
}
