package interactions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"mime/multipart"

	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
	"golang.org/x/oauth2"
)

const (
	boxTokenFile    = ".anbu-box-token.json"
	redirectURI     = "http://localhost:8080"
	apiBaseURL      = "https://api.box.com/2.0"
	uploadBaseURL   = "https://upload.box.com/api/2.0"
	folderItemsURL  = apiBaseURL + "/folders/%s/items"
	fileContentURL  = apiBaseURL + "/files/%s/content"
	uploadFileURL   = uploadBaseURL + "/files/content"
	uploadFolderURL = apiBaseURL + "/folders"
)

type BoxCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type BoxItem struct {
	Type   string `json:"type"`
	ID     string `json:"id"`
	Name   string `json:"name"`
	Size   *int64 `json:"size"`
	Parent *struct {
		ID string `json:"id"`
	} `json:"parent"`
	ModifiedAt *string `json:"modified_at"`
}

type BoxFolderItems struct {
	TotalCount int       `json:"total_count"`
	Entries    []BoxItem `json:"entries"`
}

type BoxError struct {
	Type    string `json:"type"`
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type BoxItemDisplay struct {
	Name         string
	ModifiedTime string
	Size         int64
	Type         string
}

func GetBoxClient(credentialsFile string) (*http.Client, error) {
	ctx := context.Background()
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %v", err)
	}
	var creds BoxCredentials
	if err := json.Unmarshal(b, &creds); err != nil {
		return nil, fmt.Errorf("unable to parse credentials file: %v", err)
	}
	if creds.ClientID == "" || creds.ClientSecret == "" {
		return nil, fmt.Errorf("credentials file must contain client_id and client_secret")
	}
	config := &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://account.box.com/api/oauth2/authorize",
			TokenURL: "https://api.box.com/oauth2/token",
		},
		RedirectURL: redirectURI,
		Scopes:      []string{"root_readwrite"},
	}
	token, err := getBoxOAuthToken(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get OAuth token: %v", err)
	}
	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("unable to refresh token: %v", err)
	}
	if newToken.AccessToken != token.AccessToken {
		log.Debug().Str("op", "box/auth").Msg("access token was refreshed")
		saveBoxToken(newToken)
	}
	return oauth2.NewClient(ctx, tokenSource), nil
}

func getBoxOAuthToken(config *oauth2.Config) (*oauth2.Token, error) {
	tokenFile, err := getBoxTokenFilePath()
	if err != nil {
		return nil, err
	}
	token, err := boxTokenFromFile(tokenFile)
	if err == nil {
		if token.Valid() {
			log.Debug().Str("op", "box/auth").Msg("existing token retrieved and valid")
			return token, nil
		}
		if token.RefreshToken != "" {
			log.Debug().Str("op", "box/auth").Msg("refreshing expired token")
			tokenSource := config.TokenSource(context.Background(), token)
			newToken, err := tokenSource.Token()
			if err != nil {
				return nil, fmt.Errorf("unable to refresh token: %v", err)
			}
			token = newToken
			if err := saveBoxToken(token); err != nil {
				log.Warn().Str("op", "box/auth").Msgf("unable to save refreshed token: %v", err)
			}
			return token, nil
		}
	}
	log.Debug().Str("op", "box/auth").Msg("no valid token, starting new OAuth flow")
	state := fmt.Sprintf("st%d", os.Getpid())
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("\nVisit this URL to authorize Anbu:\n\n%s\n", u.FInfo(authURL))
	fmt.Printf("\nAfter authorizing, you will be redirected to a 'localhost' URL.\n")
	fmt.Printf("Copy the *entire* 'localhost' URL from your browser and paste it here: ")
	var redirectURLStr string
	if _, err := fmt.Scanln(&redirectURLStr); err != nil {
		return nil, fmt.Errorf("unable to read redirect URL: %v", err)
	}
	parsedURL, err := url.Parse(redirectURLStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse the pasted URL: %v", err)
	}
	code := parsedURL.Query().Get("code")
	returnedState := parsedURL.Query().Get("state")
	if code == "" {
		return nil, fmt.Errorf("pasted URL did not contain an authorization 'code'")
	}
	if returnedState != state {
		return nil, fmt.Errorf("CSRF state mismatch. Expected '%s' but got '%s'", state, returnedState)
	}
	fmt.Println("Trading code for token...")
	token, err = config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("unable to exchange auth code for token: %v", err)
	}
	if err := saveBoxToken(token); err != nil {
		log.Warn().Str("op", "box/auth").Msgf("unable to save new token: %v", err)
	}
	fmt.Println(u.FSuccess("\nAuthentication successful. Token saved."))
	return token, nil
}

func getBoxTokenFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, boxTokenFile), nil
}

func boxTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func saveBoxToken(token *oauth2.Token) error {
	tokenFile, err := getBoxTokenFilePath()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
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
		req, err := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, currentID), nil)
		if err != nil {
			return "", "", fmt.Errorf("http error creating request: %w", err)
		}
		q := req.URL.Query()
		q.Add("fields", "type,name")
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
		found := false
		for _, item := range items.Entries {
			if strings.EqualFold(item.Name, segment) {
				currentID = item.ID
				currentType = item.Type
				found = true
				break
			}
		}
		if !found {
			return "", "", fmt.Errorf("path not found: '%s' in '%s'", segment, path)
		}
		isLastSegment := (i == len(segments)-1)
		if isLastSegment {
			if expectedType != "" && currentType != expectedType {
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

func ListBoxContents(client *http.Client, folderPath string) ([]BoxItemDisplay, []BoxItemDisplay, error) {
	folderID, _, err := resolvePathToID(client, folderPath, "folder")
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, folderID), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %v", err)
	}
	q := req.URL.Query()
	q.Add("fields", "type,name,size,modified_at")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list items: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil, handleBoxAPIError("list items", resp)
	}
	var items BoxFolderItems
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, nil, fmt.Errorf("failed to parse list response: %v", err)
	}
	var folders []BoxItemDisplay
	var files []BoxItemDisplay
	for _, item := range items.Entries {
		var modTime string
		if item.ModifiedAt != nil {
			t, err := time.Parse(time.RFC3339, *item.ModifiedAt)
			if err == nil {
				modTime = t.Format("2006-01-02 15:04")
			}
		}
		display := BoxItemDisplay{
			Name:         item.Name,
			ModifiedTime: modTime,
			Type:         item.Type,
		}
		if item.Size != nil {
			display.Size = *item.Size
		}
		if item.Type == "folder" {
			folders = append(folders, display)
		} else {
			files = append(files, display)
		}
	}
	sort.Slice(folders, func(i, j int) bool { return folders[i].Name < folders[j].Name })
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	return folders, files, nil
}

func UploadBoxFile(client *http.Client, localPath string, boxFolderPath string) error {
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

func DownloadBoxFile(client *http.Client, boxFilePath string, localPath string) (string, error) {
	fileID, itemType, err := resolvePathToID(client, boxFilePath, "file")
	if err != nil {
		return "", fmt.Errorf("failed to find file '%s': %v", boxFilePath, err)
	}
	if itemType != "file" {
		return "", fmt.Errorf("path '%s' is a %s, not a file", boxFilePath, itemType)
	}
	if localPath == "" {
		parts := strings.Split(strings.Trim(boxFilePath, "/"), "/")
		if len(parts) > 0 {
			localPath = parts[len(parts)-1]
		} else {
			localPath = "downloaded_file"
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
				log.Error().Err(err).Msgf("Failed to create directory %s, skipping...", currentLocalPath)
				return nil
			}
			folderIdMap[currentLocalPath] = boxFolderId
		} else {
			file, err := os.Open(currentLocalPath)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to open local file %s, skipping...", currentLocalPath)
				return nil
			}
			defer file.Close()
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			attributesJSON := fmt.Sprintf(`{"name":"%s", "parent":{"id":"%s"}}`, d.Name(), parentBoxId)
			if err := writer.WriteField("attributes", attributesJSON); err != nil {
				log.Error().Err(err).Msgf("Failed to write attributes for %s, skipping...", currentLocalPath)
				file.Close()
				return nil
			}
			part, err := writer.CreateFormFile("file", d.Name())
			if err != nil {
				log.Error().Err(err).Msgf("Failed to create form file for %s, skipping...", currentLocalPath)
				file.Close()
				return nil
			}
			if _, err := io.Copy(part, file); err != nil {
				log.Error().Err(err).Msgf("Failed to copy file %s, skipping...", currentLocalPath)
				file.Close()
				return nil
			}
			writer.Close()
			file.Close()
			req, err := http.NewRequest("POST", uploadFileURL, body)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to create upload request for %s, skipping...", currentLocalPath)
				return nil
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())
			resp, err := client.Do(req)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to upload file %s, skipping...", currentLocalPath)
				return nil
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusCreated {
				log.Error().Msgf("Failed to upload file %s (status %d), skipping...", currentLocalPath, resp.StatusCode)
				return nil
			}
			fmt.Printf("Uploaded %s %s %s\n", u.FDebug(currentLocalPath), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(d.Name()))
		}
		return nil
	})
}

func findOrCreateBoxFolder(client *http.Client, folderName string, parentId string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(folderItemsURL, parentId), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	q := req.URL.Query()
	q.Add("fields", "type,name")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to list folder: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		var items BoxFolderItems
		if err := json.NewDecoder(resp.Body).Decode(&items); err == nil {
			for _, item := range items.Entries {
				if strings.EqualFold(item.Name, folderName) && item.Type == "folder" {
					return item.ID, nil
				}
			}
		}
	}
	log.Debug().Msgf("Folder '%s' not found, creating it...", folderName)
	folderJSON := fmt.Sprintf(`{"name":"%s", "parent":{"id":"%s"}}`, folderName, parentId)
	req, err = http.NewRequest("POST", uploadFolderURL, bytes.NewBufferString(folderJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create folder request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
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
				log.Error().Err(err).Msgf("Failed to create local directory %s, skipping...", currentLocalPath)
				continue
			}
			if err := downloadBoxFolderContents(client, item.ID, currentLocalPath); err != nil {
				log.Error().Err(err).Msgf("Failed to download subfolder %s, skipping...", item.Name)
			}
		} else {
			out, err := os.Create(currentLocalPath)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to create local file %s, skipping...", currentLocalPath)
				continue
			}
			req, err := http.NewRequest("GET", fmt.Sprintf(fileContentURL, item.ID), nil)
			if err != nil {
				out.Close()
				log.Error().Err(err).Msgf("Failed to create download request for %s, skipping...", item.Name)
				continue
			}
			resp, err := client.Do(req)
			if err != nil {
				out.Close()
				log.Error().Err(err).Msgf("Failed to download file %s, skipping...", item.Name)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				out.Close()
				log.Error().Msgf("Failed to download file %s (status %d), skipping...", item.Name, resp.StatusCode)
				continue
			}
			if _, err := io.Copy(out, resp.Body); err != nil {
				resp.Body.Close()
				out.Close()
				log.Error().Err(err).Msgf("Failed to write file %s, skipping...", item.Name)
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
