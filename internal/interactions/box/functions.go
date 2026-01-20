package box

import (
	"bytes"
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

	"mime/multipart"

	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
)

func resolvePathToID(client *http.Client, path string, expectedType string) (string, string, error) {
	if path == "" || path == "/" || path == "root" {
		return "0", "folder", nil
	}
	segments := strings.Split(strings.Trim(path, "/"), "/")
	currentID := "0"
	var currentType string
	for i, segment := range segments {
		if segment == "" {
			continue
		}
		isFirstSegment := (i == 0)
		isLastSegment := (i == len(segments)-1)
		isNumeric := strings.TrimSpace(segment) != "" && isAllDigits(segment)
		if isNumeric {
			if isFirstSegment {
				folderID, folderType, err := tryResolveFolderByID(client, segment)
				if err == nil {
					currentID = folderID
					currentType = folderType
					if isLastSegment {
						if expectedType != "" && currentType != expectedType {
							return "", "", fmt.Errorf("path error: ID '%s' is a %s, but expected a %s", segment, currentType, expectedType)
						}
						return currentID, currentType, nil
					}
					continue
				}
			}
			if isLastSegment {
				fileID, fileType, err := tryResolveFileByID(client, segment)
				if err == nil {
					if expectedType != "" && fileType != expectedType {
						return "", "", fmt.Errorf("path error: ID '%s' is a %s, but expected a %s", segment, fileType, expectedType)
					}
					return fileID, fileType, nil
				}
				folderID, folderType, err := tryResolveFolderByID(client, segment)
				if err == nil {
					if expectedType != "" && folderType != expectedType {
						return "", "", fmt.Errorf("path error: ID '%s' is a %s, but expected a %s", segment, folderType, expectedType)
					}
					return folderID, folderType, nil
				}
			}
		}
		found := false
		offset := 0
		limit := 1000
		for !found {
			req, err := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, currentID), nil)
			if err != nil {
				return "", "", fmt.Errorf("http error creating request: %w", err)
			}
			q := req.URL.Query()
			q.Add("fields", "type,name")
			q.Add("limit", fmt.Sprintf("%d", limit))
			q.Add("offset", fmt.Sprintf("%d", offset))
			req.URL.RawQuery = q.Encode()
			resp, err := client.Do(req)
			if err != nil {
				return "", "", fmt.Errorf("http error listing folder %s: %w", currentID, err)
			}
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				return "", "", fmt.Errorf("api error listing folder %s (status %d): %s", currentID, resp.StatusCode, string(body))
			}
			var items BoxFolderItems
			err = json.NewDecoder(resp.Body).Decode(&items)
			resp.Body.Close()
			if err != nil {
				return "", "", fmt.Errorf("json parse error: %w", err)
			}
			for _, item := range items.Entries {
				if strings.EqualFold(item.Name, segment) {
					currentID = item.ID
					currentType = item.Type
					found = true
					break
				}
			}
			if found {
				break
			}
			offset += len(items.Entries)
			if offset >= items.TotalCount || len(items.Entries) == 0 {
				break
			}
		}
		if !found {
			log.Debug().Str("segment", segment).Str("path", path).Str("currentID", currentID).Msg("path segment not found during traversal")
			return "", "", fmt.Errorf("path not found: '%s' in '%s'", segment, path)
		}
		if isLastSegment {
			if expectedType != "" && currentType != expectedType {
				log.Debug().Str("segment", segment).Str("actualType", currentType).Str("expectedType", expectedType).Msg("type mismatch in path resolution")
				return "", "", fmt.Errorf("path error: '%s' is a %s, but expected a %s", segment, currentType, expectedType)
			}
			return currentID, currentType, nil
		}
		if currentType != "folder" {
			return "", "", fmt.Errorf("path error: '%s' in '%s' is a file, not a folder", segment, path)
		}
	}
	return "0", "folder", nil
}

func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func tryResolveFolderByID(client *http.Client, folderID string) (string, string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/folders/%s", apiBaseURL, folderID), nil)
	if err != nil {
		return "", "", err
	}
	q := req.URL.Query()
	q.Add("fields", "type,id")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("status %d", resp.StatusCode)
	}
	var folder BoxItem
	if err := json.NewDecoder(resp.Body).Decode(&folder); err != nil {
		return "", "", err
	}
	return folder.ID, folder.Type, nil
}

func tryResolveFileByID(client *http.Client, fileID string) (string, string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/files/%s", apiBaseURL, fileID), nil)
	if err != nil {
		return "", "", err
	}
	q := req.URL.Query()
	q.Add("fields", "type,id")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("status %d", resp.StatusCode)
	}
	var file BoxItem
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return "", "", err
	}
	return file.ID, file.Type, nil
}

func ListBoxContents(client *http.Client, path string) ([]BoxItemDisplay, []BoxItemDisplay, error) {
	if path == "" || path == "/" || path == "root" {
		return listBoxFolderContents(client, "0")
	}
	itemID, itemType, err := resolvePathToID(client, path, "")
	if err != nil {
		return nil, nil, err
	}
	if itemType == "file" {
		return listBoxFileInfo(client, itemID)
	}
	return listBoxFolderContents(client, itemID)
}

func listBoxFileInfo(client *http.Client, fileID string) ([]BoxItemDisplay, []BoxItemDisplay, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/files/%s", apiBaseURL, fileID), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %v", err)
	}
	q := req.URL.Query()
	q.Add("fields", "type,name,size,modified_at")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file info: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil, handleBoxAPIError("get file info", resp)
	}
	var item BoxItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, nil, fmt.Errorf("failed to parse file response: %v", err)
	}
	var modTime string
	if item.ModifiedAt != nil {
		t, err := time.Parse(time.RFC3339, *item.ModifiedAt)
		if err == nil {
			modTime = t.Format("2006-01-02 15:04")
		}
	}
	display := BoxItemDisplay{
		ID:           item.ID,
		Name:         item.Name,
		ModifiedTime: modTime,
		Type:         item.Type,
	}
	if item.Size != nil {
		display.Size = *item.Size
	}
	return nil, []BoxItemDisplay{display}, nil
}

func listBoxFolderContents(client *http.Client, folderID string) ([]BoxItemDisplay, []BoxItemDisplay, error) {
	var allFolders []BoxItemDisplay
	var allFiles []BoxItemDisplay
	offset := 0
	limit := 1000
	for {
		req, err := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, folderID), nil)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create request: %v", err)
		}
		q := req.URL.Query()
		q.Add("fields", "type,name,id,size,modified_at")
		q.Add("limit", fmt.Sprintf("%d", limit))
		q.Add("offset", fmt.Sprintf("%d", offset))
		req.URL.RawQuery = q.Encode()
		resp, err := client.Do(req)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list items: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, nil, handleBoxAPIError("list items", resp)
		}
		var items BoxFolderItems
		err = json.NewDecoder(resp.Body).Decode(&items)
		resp.Body.Close()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse list response: %v", err)
		}
		for _, item := range items.Entries {
			var modTime string
			if item.ModifiedAt != nil {
				t, err := time.Parse(time.RFC3339, *item.ModifiedAt)
				if err == nil {
					modTime = t.Format("2006-01-02 15:04")
				}
			}
			display := BoxItemDisplay{
				ID:           item.ID,
				Name:         item.Name,
				ModifiedTime: modTime,
				Type:         item.Type,
			}
			if item.Size != nil {
				display.Size = *item.Size
			}
			if item.Type == "folder" {
				allFolders = append(allFolders, display)
			} else {
				allFiles = append(allFiles, display)
			}
		}
		offset += len(items.Entries)
		if offset >= items.TotalCount || len(items.Entries) == 0 {
			break
		}
	}
	sort.Slice(allFolders, func(i, j int) bool { return allFolders[i].Name < allFolders[j].Name })
	sort.Slice(allFiles, func(i, j int) bool { return allFiles[i].Name < allFiles[j].Name })
	return allFolders, allFiles, nil
}

func UploadBoxItem(client *http.Client, localPath string, boxFolderPath string) error {
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("failed to stat local path '%s': %v", localPath, err)
	}
	if fileInfo.IsDir() {
		return UploadBoxFolder(client, localPath, boxFolderPath)
	}
	return uploadBoxFile(client, localPath, boxFolderPath)
}

func uploadBoxFile(client *http.Client, localPath string, boxFolderPath string) error {
	parentFolderID := "0"
	if boxFolderPath != "" {
		var err error
		parentFolderID, _, err = resolvePathToID(client, boxFolderPath, "folder")
		if err != nil {
			return fmt.Errorf("failed to find parent folder '%s': %v", boxFolderPath, err)
		}
	}
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file '%s': %v", localPath, err)
	}
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileName := filepath.Base(localPath)
	attributesJSON := fmt.Sprintf(`{"name":"%s", "parent":{"id":"%s"}}`, fileName, parentFolderID)
	if err := writer.WriteField("attributes", attributesJSON); err != nil {
		return fmt.Errorf("failed to write attributes field: %v", err)
	}
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file to form: %v", err)
	}
	writer.Close()
	req, err := http.NewRequest("POST", uploadFileURL, body)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return handleBoxAPIError("upload file", resp)
	}
	var result BoxFolderItems
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse upload response: %v", err)
	}
	if len(result.Entries) > 0 {
		fmt.Printf("Successfully uploaded %s %s %s (ID: %s)\n", u.FDebug(localPath), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(result.Entries[0].Name), u.FDebug(result.Entries[0].ID))
	} else {
		fmt.Printf("Successfully uploaded %s\n", u.FSuccess(fileName))
	}
	return nil
}

func DownloadBoxItem(client *http.Client, boxPath string, localPath string) (string, error) {
	itemID, itemType, err := resolvePathToID(client, boxPath, "")
	if err != nil {
		return "", fmt.Errorf("failed to find item '%s': %v", boxPath, err)
	}
	if itemType == "folder" {
		if localPath != "" {
			return "", fmt.Errorf("local path cannot be specified when downloading a folder")
		}
		return "", DownloadBoxFolder(client, boxPath)
	}
	return downloadBoxFile(client, itemID, boxPath, localPath)
}

func downloadBoxFile(client *http.Client, fileID string, boxFilePath string, localPath string) (string, error) {
	if localPath == "" {
		isNumericID := isAllDigits(strings.TrimSpace(boxFilePath))
		if isNumericID {
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/files/%s", apiBaseURL, fileID), nil)
			if err != nil {
				return "", fmt.Errorf("failed to create request: %v", err)
			}
			q := req.URL.Query()
			q.Add("fields", "name")
			req.URL.RawQuery = q.Encode()
			resp, err := client.Do(req)
			if err != nil {
				return "", fmt.Errorf("failed to get file info: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var file BoxItem
				if err := json.NewDecoder(resp.Body).Decode(&file); err == nil && file.Name != "" {
					localPath = file.Name
				}
			}
		}
		if localPath == "" {
			parts := strings.Split(strings.Trim(boxFilePath, "/"), "/")
			if len(parts) > 0 {
				localPath = parts[len(parts)-1]
			} else {
				localPath = "downloaded_file"
			}
		}
	}
	out, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file '%s': %v", localPath, err)
	}
	defer out.Close()
	req, err := http.NewRequest("GET", fmt.Sprintf(fileContentURL, fileID), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create download request: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", handleBoxAPIError("download file", resp)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", fmt.Errorf("failed to write to local file: %v", err)
	}
	fmt.Printf("Downloaded %s %s %s\n", u.FDebug(boxFilePath), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(localPath))
	return localPath, nil
}

func UploadBoxFolder(client *http.Client, localPath string, boxFolderPath string) error {
	parentFolderID := "0"
	if boxFolderPath != "" {
		var err error
		parentFolderID, _, err = resolvePathToID(client, boxFolderPath, "folder")
		if err != nil {
			return fmt.Errorf("failed to find parent folder '%s': %v", boxFolderPath, err)
		}
	}
	rootFolderName := filepath.Base(localPath)
	driveRootFolderID, err := findOrCreateBoxFolder(client, rootFolderName, parentFolderID)
	if err != nil {
		return fmt.Errorf("failed to create root box folder '%s': %v", rootFolderName, err)
	}
	folderIdMap := make(map[string]string)
	folderIdMap[localPath] = driveRootFolderID
	return filepath.WalkDir(localPath, func(currentLocalPath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if currentLocalPath == localPath {
			return nil
		}
		parentLocalDir := filepath.Dir(currentLocalPath)
		parentBoxId, ok := folderIdMap[parentLocalDir]
		if !ok {
			return fmt.Errorf("could not find parent Box ID for local path: %s", parentLocalDir)
		}
		if d.IsDir() {
			boxFolderId, err := findOrCreateBoxFolder(client, d.Name(), parentBoxId)
			if err != nil {
				u.PrintError(fmt.Sprintf("Failed to create directory %s, skipping", currentLocalPath))
				log.Debug().Err(err).Msgf("Failed to create directory %s, skipping", currentLocalPath)
				return nil
			}
			folderIdMap[currentLocalPath] = boxFolderId
		} else {
			file, err := os.Open(currentLocalPath)
			if err != nil {
				u.PrintError(fmt.Sprintf("Failed to open local file %s, skipping", currentLocalPath))
				log.Debug().Err(err).Msgf("Failed to open local file %s, skipping", currentLocalPath)
				return nil
			}
			defer file.Close()
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			attributesJSON := fmt.Sprintf(`{"name":"%s", "parent":{"id":"%s"}}`, d.Name(), parentBoxId)
			if err := writer.WriteField("attributes", attributesJSON); err != nil {
				u.PrintError(fmt.Sprintf("Failed to write attributes for %s, skipping", currentLocalPath))
				log.Debug().Err(err).Msgf("Failed to write attributes for %s, skipping", currentLocalPath)
				file.Close()
				return nil
			}
			part, err := writer.CreateFormFile("file", d.Name())
			if err != nil {
				u.PrintError(fmt.Sprintf("Failed to create form file for %s, skipping", currentLocalPath))
				log.Debug().Err(err).Msgf("Failed to create form file for %s, skipping", currentLocalPath)
				file.Close()
				return nil
			}
			if _, err := io.Copy(part, file); err != nil {
				u.PrintError(fmt.Sprintf("Failed to copy file %s, skipping", currentLocalPath))
				log.Debug().Err(err).Msgf("Failed to copy file %s, skipping", currentLocalPath)
				file.Close()
				return nil
			}
			writer.Close()
			file.Close()
			req, err := http.NewRequest("POST", uploadFileURL, body)
			if err != nil {
				u.PrintError(fmt.Sprintf("Failed to create upload request for %s, skipping", currentLocalPath))
				log.Debug().Err(err).Msgf("Failed to create upload request for %s, skipping", currentLocalPath)
				return nil
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())
			resp, err := client.Do(req)
			if err != nil {
				u.PrintError(fmt.Sprintf("Failed to upload file %s, skipping", currentLocalPath))
				log.Debug().Err(err).Msgf("Failed to upload file %s, skipping", currentLocalPath)
				return nil
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusCreated {
				u.PrintError(fmt.Sprintf("Failed to upload file %s (status %d), skipping", currentLocalPath, resp.StatusCode))
				log.Debug().Msgf("Failed to upload file %s (status %d), skipping", currentLocalPath, resp.StatusCode)
				return nil
			}
			fmt.Printf("Uploaded %s %s %s\n", u.FDebug(currentLocalPath), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(d.Name()))
		}
		return nil
	})
}

func findOrCreateBoxFolder(client *http.Client, folderName string, parentId string) (string, error) {
	offset := 0
	limit := 1000
	for {
		req, err := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, parentId), nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %v", err)
		}
		q := req.URL.Query()
		q.Add("fields", "type,name")
		q.Add("limit", fmt.Sprintf("%d", limit))
		q.Add("offset", fmt.Sprintf("%d", offset))
		req.URL.RawQuery = q.Encode()
		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to list folder: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			break
		}
		var items BoxFolderItems
		err = json.NewDecoder(resp.Body).Decode(&items)
		resp.Body.Close()
		if err != nil {
			break
		}
		for _, item := range items.Entries {
			if strings.EqualFold(item.Name, folderName) && item.Type == "folder" {
				return item.ID, nil
			}
		}
		offset += len(items.Entries)
		if offset >= items.TotalCount || len(items.Entries) == 0 {
			break
		}
	}
	log.Debug().Msgf("Folder '%s' not found, creating it", folderName)
	folderJSON := fmt.Sprintf(`{"name":"%s", "parent":{"id":"%s"}}`, folderName, parentId)
	req, err := http.NewRequest("POST", uploadFolderURL, bytes.NewBufferString(folderJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create folder request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create folder: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", handleBoxAPIError("create folder", resp)
	}
	var folder BoxItem
	if err := json.NewDecoder(resp.Body).Decode(&folder); err != nil {
		return "", fmt.Errorf("failed to parse folder response: %v", err)
	}
	fmt.Printf("Created Box folder %s\n", u.FSuccess(folderName))
	return folder.ID, nil
}

func DownloadBoxFolder(client *http.Client, boxFolderPath string) error {
	folderID, itemType, err := resolvePathToID(client, boxFolderPath, "folder")
	if err != nil {
		return fmt.Errorf("failed to find folder '%s': %v", boxFolderPath, err)
	}
	if itemType != "folder" {
		return errors.New("path is not a folder. Use 'download' instead")
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/folders/%s", apiBaseURL, folderID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	q := req.URL.Query()
	q.Add("fields", "name")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get folder info: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return handleBoxAPIError("get folder info", resp)
	}
	var folder BoxItem
	if err := json.NewDecoder(resp.Body).Decode(&folder); err != nil {
		return fmt.Errorf("failed to parse folder response: %v", err)
	}
	localFolderPath := folder.Name
	if err := os.Mkdir(localFolderPath, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create local root folder '%s': %v", localFolderPath, err)
	}
	return downloadBoxFolderContents(client, folderID, localFolderPath)
}

func downloadBoxFolderContents(client *http.Client, folderID string, localDestPath string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, folderID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	q := req.URL.Query()
	q.Add("fields", "type,name")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to list folder contents: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return handleBoxAPIError("list folder contents", resp)
	}
	var items BoxFolderItems
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return fmt.Errorf("failed to parse folder items: %v", err)
	}
	for _, item := range items.Entries {
		currentLocalPath := filepath.Join(localDestPath, item.Name)
		if item.Type == "folder" {
			if err := os.Mkdir(currentLocalPath, 0755); err != nil && !os.IsExist(err) {
				u.PrintError(fmt.Sprintf("Failed to create local directory %s, skipping", currentLocalPath))
				log.Debug().Err(err).Msgf("Failed to create local directory %s, skipping", currentLocalPath)
				continue
			}
			if err := downloadBoxFolderContents(client, item.ID, currentLocalPath); err != nil {
				u.PrintError(fmt.Sprintf("Failed to download subfolder %s, skipping", item.Name))
				log.Debug().Err(err).Msgf("Failed to download subfolder %s, skipping", item.Name)
			}
		} else {
			out, err := os.Create(currentLocalPath)
			if err != nil {
				u.PrintError(fmt.Sprintf("Failed to create local file %s, skipping", currentLocalPath))
				log.Debug().Err(err).Msgf("Failed to create local file %s, skipping", currentLocalPath)
				continue
			}
			req, err := http.NewRequest("GET", fmt.Sprintf(fileContentURL, item.ID), nil)
			if err != nil {
				out.Close()
				u.PrintError(fmt.Sprintf("Failed to create download request for %s, skipping", item.Name))
				log.Debug().Err(err).Msgf("Failed to create download request for %s, skipping", item.Name)
				continue
			}
			resp, err := client.Do(req)
			if err != nil {
				out.Close()
				u.PrintError(fmt.Sprintf("Failed to download file %s, skipping", item.Name))
				log.Debug().Err(err).Msgf("Failed to download file %s, skipping", item.Name)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				out.Close()
				u.PrintError(fmt.Sprintf("Failed to download file %s (status %d), skipping", item.Name, resp.StatusCode))
				log.Debug().Msgf("Failed to download file %s (status %d), skipping", item.Name, resp.StatusCode)
				continue
			}
			if _, err := io.Copy(out, resp.Body); err != nil {
				resp.Body.Close()
				out.Close()
				u.PrintError(fmt.Sprintf("Failed to write file %s, skipping", item.Name))
				log.Debug().Err(err).Msgf("Failed to write file %s, skipping", item.Name)
				continue
			}
			resp.Body.Close()
			out.Close()
			fmt.Printf("Downloaded %s %s %s\n", u.FDebug(item.Name), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(currentLocalPath))
		}
	}
	return nil
}

func handleBoxAPIError(action string, resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("api request to '%s' failed with status %s (could not read error body)", action, resp.Status)
	}
	var boxErr BoxError
	if json.Unmarshal(body, &boxErr) == nil {
		return fmt.Errorf("api request to '%s' failed: %s - %s", action, boxErr.Code, boxErr.Message)
	}
	return fmt.Errorf("api request to '%s' failed with status %s: %s", action, resp.Status, string(body))
}

func HumanReadableBoxSize(size int64) string {
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
