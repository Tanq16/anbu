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
	Use:     "list [folder-path]",
	Aliases: []string{"ls"},
	Short:   "List files and folders in Box",
	Long:    `Lists files and folders. If [folder-path] is provided, lists content of that folder. Otherwise, lists the root folder.`,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		folderPath := ""
		if len(args) > 0 {
			folderPath = args[0]
		}
		client, err := box.GetBoxClient(boxFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Box client")
		}
		folders, files, err := box.ListBoxContents(client, folderPath)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to list contents for '%s'", folderPath)
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
			u.PrintInfo(fmt.Sprintf("No items found in '%s'", folderPath))
			return
		}
		table.PrintTable(false)
	},
}

var boxUploadCmd = &cobra.Command{
	Use:     "upload <local-file> [box-folder-path]",
	Aliases: []string{"up"},
	Short:   "Upload a local file to Box",
	Long:    `Uploads a single local file. If [box-folder-path] is provided, uploads to that folder. Otherwise, uploads to the root folder.`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		localPath := args[0]
		boxFolderPath := ""
		if len(args) > 1 {
			boxFolderPath = args[1]
		}
		client, err := box.GetBoxClient(boxFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Box client")
		}
		u.PrintInfo(fmt.Sprintf("Starting upload of %s to %s...", u.FDebug(localPath), u.FDebug(boxFolderPath)))
		if err := box.UploadBoxFile(client, localPath, boxFolderPath); err != nil {
			log.Fatal().Err(err).Msg("Failed to upload file")
		}
	},
}

var boxDownloadCmd = &cobra.Command{
	Use:     "download <box-file-path> [local-path]",
	Aliases: []string{"dl"},
	Short:   "Download a file from Box",
	Long:    `Downloads a file from Box to the current directory. If [local-path] is provided, saves to that path. Otherwise, uses the file name from Box.`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		boxFilePath := args[0]
		localPath := ""
		if len(args) > 1 {
			localPath = args[1]
		}
		client, err := box.GetBoxClient(boxFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Box client")
		}
		u.PrintInfo(fmt.Sprintf("Starting download of %s...", u.FDebug(boxFilePath)))
		downloadedPath, err := box.DownloadBoxFile(client, boxFilePath, localPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to download file")
		}
		fmt.Printf("Successfully downloaded %s %s %s\n", u.FDebug(boxFilePath), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(downloadedPath))
	},
}

var boxUploadFolderCmd = &cobra.Command{
	Use:     "upload-folder <local-folder> [box-folder-path]",
	Aliases: []string{"up-f"},
	Short:   "Upload a local folder recursively to Box",
	Long:    `Uploads a local folder recursively. If [box-folder-path] is provided, uploads into that folder. Otherwise, creates the new folder in the root.`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		localPath := args[0]
		boxFolderPath := ""
		if len(args) > 1 {
			boxFolderPath = args[1]
		}
		client, err := box.GetBoxClient(boxFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Box client")
		}
		u.PrintInfo(fmt.Sprintf("Starting folder upload of %s to %s...", u.FDebug(localPath), u.FDebug(boxFolderPath)))
		if err := box.UploadBoxFolder(client, localPath, boxFolderPath); err != nil {
			log.Fatal().Err(err).Msg("Failed to upload folder")
		}
		u.PrintSuccess(fmt.Sprintf("Successfully uploaded folder %s", localPath))
	},
}

var boxDownloadFolderCmd = &cobra.Command{
	Use:     "download-folder <box-folder-path>",
	Aliases: []string{"dl-f"},
	Short:   "Download a folder recursively from Box",
	Long:    `Downloads a folder recursively from Box to the current directory. <box-folder-path> is the full path to the folder (e.g., 'MyFolder/MySubFolder').`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		boxFolderPath := args[0]
		client, err := box.GetBoxClient(boxFlags.credentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get Box client")
		}
		u.PrintInfo(fmt.Sprintf("Starting folder download of %s...", u.FDebug(boxFolderPath)))
		if err := box.DownloadBoxFolder(client, boxFolderPath); err != nil {
			log.Fatal().Err(err).Msg("Failed to download folder")
		}
		u.PrintSuccess(fmt.Sprintf("Successfully downloaded folder %s", boxFolderPath))
	},
}

func init() {
	BoxCmd.PersistentFlags().StringVarP(&boxFlags.credentialsFile, "credentials", "c", "", "Path to Box credentials.json file (default ~/.anbu-box-credentials.json)")

	BoxCmd.AddCommand(boxListCmd)
	BoxCmd.AddCommand(boxUploadCmd)
	BoxCmd.AddCommand(boxDownloadCmd)
	BoxCmd.AddCommand(boxUploadFolderCmd)
	BoxCmd.AddCommand(boxDownloadFolderCmd)
}
