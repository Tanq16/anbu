package gdrive

const (
	gdriveTokenFile      = ".anbu-gdrive-token.json"
	googleFolderMimeType = "application/vnd.google-apps.folder"
)

type DriveItem struct {
	Name         string
	ModifiedTime string
	Size         int64
}
