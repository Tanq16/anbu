package interactions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	gdriveTokenFile      = ".anbu-gdrive-token.json"
	googleFolderMimeType = "application/vnd.google-apps.folder"
)

type DriveItem struct {
	Name         string
	ModifiedTime string
	Size         int64
}

func GetDriveService(credentialsFile string) (*drive.Service, error) {
	ctx := context.Background()
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file: %v", err)
	}

	token, err := getOAuthToken(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get OAuth token: %v", err)
	}

	client := config.Client(ctx, token)
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Drive client: %v", err)
	}

	return srv, nil
}

func getOAuthToken(config *oauth2.Config) (*oauth2.Token, error) {
	tokenFile, err := getTokenFilePath()
	if err != nil {
		return nil, err
	}

	token, err := tokenFromFile(tokenFile)
	if err == nil {
		if token.Valid() {
			log.Debug().Str("op", "google-drive/auth").Msgf("existing token retrieved and valid")
			return token, nil
		}
		if token.RefreshToken != "" {
			log.Debug().Str("op", "google-drive/auth").Msgf("refreshing expired token")
			tokenSource := config.TokenSource(context.Background(), token)
			newToken, err := tokenSource.Token()
			if err != nil {
				return nil, fmt.Errorf("unable to refresh token: %v", err)
			}
			token = newToken
			if err := saveToken(tokenFile, token); err != nil {
				log.Warn().Str("op", "google-drive/auth").Msgf("unable to save refreshed token: %v", err)
			}
			return token, nil
		}
	}

	log.Debug().Str("op", "google-drive/auth").Msgf("no valid token, starting new OAuth flow")
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Printf("\nVisit this URL to authorize Anbu:\n\n%s\n", u.FInfo(authURL))
	fmt.Printf("\nAfter authorizing, enter the authorization code: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %v", err)
	}

	token, err = config.Exchange(context.Background(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to exchange auth code for token: %v", err)
	}
	if err := saveToken(tokenFile, token); err != nil {
		log.Warn().Str("op", "google-drive/auth").Msgf("unable to save new token: %v", err)
	}
	fmt.Println(u.FSuccess("\nAuthentication successful. Token saved."))
	return token, nil
}

func getTokenFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, gdriveTokenFile), nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func saveToken(file string, token *oauth2.Token) error {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return fmt.Errorf("unable to encode token: %v", err)
	}
	return nil
}

func getFolderIdByName(srv *drive.Service, name string, parentId string) (string, error) {
	if name == "root" || name == "" {
		return "root", nil
	}

	query := fmt.Sprintf("name = '%s' and '%s' in parents and mimeType = '%s' and trashed = false", name, parentId, googleFolderMimeType)
	r, err := srv.Files.List().Q(query).Fields("files(id, name)").PageSize(1).Do()
	if err != nil {
		return "", fmt.Errorf("failed to query for folder '%s': %v", name, err)
	}

	if len(r.Files) == 0 {
		return "", fmt.Errorf("folder not found: '%s'", name)
	}

	return r.Files[0].Id, nil
}

func getItemIdByPath(srv *drive.Service, drivePath string) (*drive.File, error) {
	if drivePath == "root" || drivePath == "" || drivePath == "/" {
		f, err := srv.Files.Get("root").Fields("id, name, mimeType").Do()
		if err != nil {
			return nil, err
		}
		return f, nil
	}

	parts := strings.Split(strings.Trim(drivePath, "/"), "/")
	currentParentId := "root"
	var currentFile *drive.File

	for i, part := range parts {
		isLastPart := (i == len(parts)-1)
		mimeTypeQuery := ""
		if !isLastPart {
			mimeTypeQuery = fmt.Sprintf("and mimeType = '%s'", googleFolderMimeType)
		}

		query := fmt.Sprintf("name = '%s' and '%s' in parents %s and trashed = false", part, currentParentId, mimeTypeQuery)
		r, err := srv.Files.List().Q(query).Fields("files(id, name, mimeType)").PageSize(1).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to search for path part '%s': %v", part, err)
		}

		if len(r.Files) == 0 {
			return nil, fmt.Errorf("path not found: '%s' (at part '%s')", drivePath, part)
		}

		currentFile = r.Files[0]
		currentParentId = currentFile.Id
	}

	return currentFile, nil
}

func findOrCreateFolder(srv *drive.Service, folderName string, parentId string) (string, error) {
	folderId, err := getFolderIdByName(srv, folderName, parentId)
	if err == nil {
		return folderId, nil // Folder exists
	}

	log.Debug().Msgf("Folder '%s' not found, creating it...", folderName)
	folderMetadata := &drive.File{
		Name:     folderName,
		MimeType: googleFolderMimeType,
		Parents:  []string{parentId},
	}

	f, err := srv.Files.Create(folderMetadata).Fields("id").Do()
	if err != nil {
		return "", fmt.Errorf("failed to create folder '%s': %v", folderName, err)
	}

	fmt.Printf("Created Drive folder %s\n", u.FSuccess(folderName))
	return f.Id, nil
}

func ListDriveContents(srv *drive.Service, folderName string) ([]DriveItem, []DriveItem, error) {
	folderId, err := getFolderIdByName(srv, folderName, "root")
	if err != nil {
		if folderName != "root" {
			// Try finding by path if name fails
			file, err := getItemIdByPath(srv, folderName)
			if err != nil {
				return nil, nil, err
			}
			if file.MimeType != googleFolderMimeType {
				return nil, nil, fmt.Errorf("path provided is not a folder: %s", folderName)
			}
			folderId = file.Id
		} else {
			return nil, nil, err
		}
	}

	query := fmt.Sprintf("'%s' in parents and trashed = false", folderId)
	var folders []DriveItem
	var files []DriveItem
	var pageToken string

	for {
		r, err := srv.Files.List().Q(query).
			PageSize(100).
			Fields("nextPageToken, files(id, name, mimeType, size, modifiedTime)").
			PageToken(pageToken).Do()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list files: %v", err)
		}

		for _, f := range r.Files {
			modTime, _ := time.Parse(time.RFC3339, f.ModifiedTime)
			item := DriveItem{
				Name:         f.Name,
				ModifiedTime: modTime.Format("2006-01-02 15:04"),
				Size:         f.Size,
			}
			if f.MimeType == googleFolderMimeType {
				folders = append(folders, item)
			} else {
				files = append(files, item)
			}
		}
		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}

	sort.Slice(folders, func(i, j int) bool { return folders[i].Name < folders[j].Name })
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })

	return folders, files, nil
}

func UploadFile(srv *drive.Service, localPath string, driveFolder string) (*drive.File, error) {
	folderId, err := getFolderIdByName(srv, driveFolder, "root")
	if err != nil {
		if driveFolder != "root" {
			// Try finding by path if name fails
			file, err := getItemIdByPath(srv, driveFolder)
			if err != nil {
				return nil, err
			}
			if file.MimeType != googleFolderMimeType {
				return nil, fmt.Errorf("path provided is not a folder: %s", driveFolder)
			}
			folderId = file.Id
		} else {
			return nil, err
		}
	}

	file, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local file %s: %v", localPath, err)
	}
	defer file.Close()

	fileName := filepath.Base(localPath)
	fileMetadata := &drive.File{
		Name:    fileName,
		Parents: []string{folderId},
	}

	driveFile, err := srv.Files.Create(fileMetadata).Media(file).Fields("id, name").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create file in drive: %v", err)
	}

	return driveFile, nil
}

func DownloadFile(srv *drive.Service, drivePath string) (string, error) {
	file, err := getItemIdByPath(srv, drivePath)
	if err != nil {
		return "", err
	}

	if file.MimeType == googleFolderMimeType {
		return "", errors.New("path is a folder, not a file. Use 'download-folder' instead")
	}

	localPath := filepath.Base(file.Name)
	if err := downloadFileById(srv, file, localPath); err != nil {
		return "", err
	}

	return localPath, nil
}

func downloadFileById(srv *drive.Service, file *drive.File, localPath string) error {
	var resp *http.Response
	var err error

	if strings.HasPrefix(file.MimeType, "application/vnd.google-apps") {
		// Handle Google Doc formats (Docs, Sheets, Slides)
		var exportMimeType string
		switch file.MimeType {
		case "application/vnd.google-apps.document":
			exportMimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
			localPath += ".docx"
		case "application/vnd.google-apps.spreadsheet":
			exportMimeType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
			localPath += ".xlsx"
		case "application/vnd.google-apps.presentation":
			exportMimeType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
			localPath += ".pptx"
		default:
			// Try to export as PDF for other types like Drawings
			exportMimeType = "application/pdf"
			localPath += ".pdf"
		}
		log.Debug().Msgf("Exporting Google Doc %s as %s", file.Name, exportMimeType)
		resp, err = srv.Files.Export(file.Id, exportMimeType).Download()
	} else {
		// Handle regular binary files
		log.Debug().Msgf("Downloading binary file %s", file.Name)
		resp, err = srv.Files.Get(file.Id).Download()
	}

	if err != nil {
		return fmt.Errorf("failed to start download for '%s': %v", file.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download request failed for '%s' with status: %s", file.Name, resp.Status)
	}

	out, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file %s: %v", localPath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to local file %s: %v", localPath, err)
	}

	fmt.Printf("Downloaded %s %s %s\n",
		u.FDebug(file.Name),
		u.FInfo(u.StyleSymbols["arrow"]),
		u.FSuccess(localPath),
	)
	return nil
}

func UploadFolder(srv *drive.Service, localPath string, driveFolder string) error {
	parentFolderId, err := getFolderIdByName(srv, driveFolder, "root")
	if err != nil {
		if driveFolder != "root" {
			file, err := getItemIdByPath(srv, driveFolder)
			if err != nil {
				return err
			}
			if file.MimeType != googleFolderMimeType {
				return fmt.Errorf("path provided is not a folder: %s", driveFolder)
			}
			parentFolderId = file.Id
		} else {
			return err
		}
	}

	rootFolderName := filepath.Base(localPath)
	driveRootFolderId, err := findOrCreateFolder(srv, rootFolderName, parentFolderId)
	if err != nil {
		return fmt.Errorf("failed to create root drive folder '%s': %v", rootFolderName, err)
	}

	folderIdMap := make(map[string]string)
	folderIdMap[localPath] = driveRootFolderId

	return filepath.WalkDir(localPath, func(currentLocalPath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if currentLocalPath == localPath {
			return nil // Skip the root folder itself
		}

		parentLocalDir := filepath.Dir(currentLocalPath)
		parentDriveId, ok := folderIdMap[parentLocalDir]
		if !ok {
			return fmt.Errorf("could not find parent Drive ID for local path: %s", parentLocalDir)
		}

		if d.IsDir() {
			driveFolderId, err := findOrCreateFolder(srv, d.Name(), parentDriveId)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to create directory %s, skipping...", currentLocalPath)
				return nil
			}
			folderIdMap[currentLocalPath] = driveFolderId
		} else {
			file, err := os.Open(currentLocalPath)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to open local file %s, skipping...", currentLocalPath)
				return nil
			}
			defer file.Close()

			fileMetadata := &drive.File{
				Name:    d.Name(),
				Parents: []string{parentDriveId},
			}

			_, err = srv.Files.Create(fileMetadata).Media(file).Fields("id").Do()
			if err != nil {
				log.Error().Err(err).Msgf("Failed to upload file %s, skipping...", currentLocalPath)
				return nil
			}
			fmt.Printf("Uploaded %s %s %s\n",
				u.FDebug(currentLocalPath),
				u.FInfo(u.StyleSymbols["arrow"]),
				u.FSuccess(d.Name()),
			)
		}
		return nil
	})
}

func DownloadFolder(srv *drive.Service, drivePath string) error {
	folder, err := getItemIdByPath(srv, drivePath)
	if err != nil {
		return err
	}

	if folder.MimeType != googleFolderMimeType {
		return errors.New("path is not a folder. Use 'download' instead")
	}

	localFolderPath := filepath.Base(folder.Name)
	if err := os.Mkdir(localFolderPath, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create local root folder '%s': %v", localFolderPath, err)
	}

	return downloadDriveFolderContents(srv, folder.Id, localFolderPath)
}

func downloadDriveFolderContents(srv *drive.Service, folderId string, localDestPath string) error {
	query := fmt.Sprintf("'%s' in parents and trashed = false", folderId)
	var pageToken string

	for {
		r, err := srv.Files.List().Q(query).
			PageSize(100).
			Fields("nextPageToken, files(id, name, mimeType)").
			PageToken(pageToken).Do()
		if err != nil {
			return fmt.Errorf("failed to list contents of folder %s: %v", folderId, err)
		}

		for _, f := range r.Files {
			currentLocalPath := filepath.Join(localDestPath, f.Name)
			if f.MimeType == googleFolderMimeType {
				if err := os.Mkdir(currentLocalPath, 0755); err != nil && !os.IsExist(err) {
					log.Error().Err(err).Msgf("Failed to create local directory %s, skipping...", currentLocalPath)
					continue
				}
				if err := downloadDriveFolderContents(srv, f.Id, currentLocalPath); err != nil {
					log.Error().Err(err).Msgf("Failed to download subfolder %s, skipping...", f.Name)
				}
			} else {
				if err := downloadFileById(srv, f, currentLocalPath); err != nil {
					log.Error().Err(err).Msgf("Failed to download file %s, skipping...", f.Name)
				}
			}
		}

		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return nil
}

func HumanReadableSize(size int64) string {
	if size == 0 {
		return "0 B"
	}
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
