package interactionsCmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tanq16/anbu/internal/interactions/box"
	u "github.com/tanq16/anbu/utils"
)

var boxFlags struct {
	credentialsFile string
}

var BoxCmd = &cobra.Command{
	Use:   "box",
	Short: "Interact with Box.com (list, upload, download)",
	Long: `Provides commands to interact with Box.com.
Requires a credentials.json file with client_id and client_secret,
which can be specified via a flag or placed at ~/.anbu-box-credentials.json.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if boxFlags.credentialsFile == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get user home directory: %w", err)
			}
			boxFlags.credentialsFile = filepath.Join(homeDir, ".anbu-box-credentials.json")
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
	Short:   "List files and folders in Box",
	Long:    `Lists files and folders. If [path] is provided and is a folder, lists its contents. If [path] is a file, shows file info. Otherwise, lists the root folder.`,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := ""
		if len(args) > 0 {
			path, _ = box.ResolvePath(args[0])
		}
		client, err := box.GetBoxClient(boxFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Box client")
		}
		folders, files, err := box.ListBoxContents(client, path)
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
				box.HumanReadableBoxSize(f.Size),
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

var boxUploadCmd = &cobra.Command{
	Use:     "upload <local-path> [box-folder-path]",
	Aliases: []string{"up"},
	Short:   "Upload a local file or folder to Box",
	Long:    `Uploads a local file or folder to Box. If [box-folder-path] is provided, uploads to that folder. Otherwise, uploads to the root folder. Folders are uploaded recursively.`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		localPath := args[0]
		boxFolderPath := ""
		if len(args) > 1 {
			boxFolderPath, _ = box.ResolvePath(args[1])
		}
		client, err := box.GetBoxClient(boxFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Box client")
		}
		fmt.Printf("Starting upload of %s to %s...\n", u.FDebug(localPath), u.FDebug(boxFolderPath))
		if err := box.UploadBoxItem(client, localPath, boxFolderPath); err != nil {
			log.Fatal().Err(err).Msg("Failed to upload")
		}
		fmt.Printf("%s Upload completed successfully\n", u.FSuccess("✓"))
	},
}

var boxDownloadCmd = &cobra.Command{
	Use:     "download <box-path> [local-path]",
	Aliases: []string{"dl"},
	Short:   "Download a file or folder from Box",
	Long:    `Downloads a file or folder from Box. If [local-path] is provided for a file, saves to that path. For folders, downloads to current directory with the folder name.`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		boxPath, _ := box.ResolvePath(args[0])
		localPath := ""
		if len(args) > 1 {
			localPath = args[1]
		}
		client, err := box.GetBoxClient(boxFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Box client")
		}
		fmt.Printf("Starting download of %s...\n", u.FDebug(boxPath))
		downloadedPath, err := box.DownloadBoxItem(client, boxPath, localPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to download")
		}
		if downloadedPath != "" {
			fmt.Printf("Successfully downloaded %s %s %s\n", u.FDebug(boxPath), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(downloadedPath))
		}
		fmt.Printf("%s Download completed successfully\n", u.FSuccess("✓"))
	},
}

var boxSyncCmd = &cobra.Command{
	Use:   "sync <local-dir> <remote-dir>",
	Short: "Sync local directory with Box remote directory",
	Long:  `Synchronizes a local directory with a remote Box directory. Uploads missing files, deletes remote-only files, and updates changed files.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		localDir := args[0]
		remotePath, _ := box.ResolvePath(args[1])
		client, err := box.GetBoxClient(boxFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Box client")
		}
		fmt.Printf("Starting sync of %s to %s...\n", u.FDebug(localDir), u.FDebug(remotePath))
		if err := box.SyncBoxDirectory(client, localDir, remotePath); err != nil {
			log.Fatal().Err(err).Msg("Failed to sync")
		}
		fmt.Printf("%s Sync completed successfully\n", u.FSuccess("✓"))
	},
}

var boxIndexCmd = &cobra.Command{
	Use:   "index [path]",
	Short: "Index file metadata for fast searching",
	Run: func(cmd *cobra.Command, args []string) {
		path := ""
		if len(args) > 0 {
			path, _ = box.ResolvePath(args[0])
		}
		client, err := box.GetBoxClient(boxFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Box client")
		}
		if err := box.GenerateIndex(client, path); err != nil {
			log.Fatal().Err(err).Msg("Indexing failed")
		}
		u.PrintSuccess("Box indexing completed.")
	},
}

var boxSearchFlags struct {
	excludeDirs  bool
	excludeFiles bool
}

var boxSearchCmd = &cobra.Command{
	Use:   "search <regex> [path]",
	Short: "Search indexed files using regex",
	Long: `Search across your indexed Box files using regex.
Requires running 'anbu box index' first.
If [path] is omitted, it defaults to the root of the index.
Shortcuts (e.g., %project%) are supported in the path.`,
	Args: cobra.MinimumNArgs(1),
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
		u.PrintInfo(fmt.Sprintf("Searching Box index in '%s'...", displayPath))
		items, err := box.SearchIndex(regex, path, boxSearchFlags.excludeDirs, boxSearchFlags.excludeFiles)
		if err != nil {
			log.Fatal().Err(err).Msg("Search failed")
		}
		table := u.NewTable([]string{"Type", "Name", "Path", "Size", "Modified"})
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
				item.Name,
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
	BoxCmd.PersistentFlags().StringVarP(&boxFlags.credentialsFile, "credentials", "c", "", "Path to Box credentials.json file (default ~/.anbu-box-credentials.json)")

	BoxCmd.AddCommand(boxListCmd)
	BoxCmd.AddCommand(boxUploadCmd)
	BoxCmd.AddCommand(boxDownloadCmd)
	BoxCmd.AddCommand(boxSyncCmd)
	BoxCmd.AddCommand(boxIndexCmd)
	BoxCmd.AddCommand(boxSearchCmd)

	boxSearchCmd.Flags().BoolVar(&boxSearchFlags.excludeDirs, "exclude-dirs", false, "Exclude directories from results")
	boxSearchCmd.Flags().BoolVar(&boxSearchFlags.excludeFiles, "exclude-files", false, "Exclude files from results")
}
