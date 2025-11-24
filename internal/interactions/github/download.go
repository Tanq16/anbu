package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v79/github"
	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
)

func ParseGitHubURL(url string) (owner, repo, ref, path string, err error) {
	parts := strings.Split(strings.Trim(url, "/"), "/")
	if len(parts) < 4 {
		return "", "", "", "", fmt.Errorf("invalid URL format: expected OWNER/REPO/tree/REF/PATH")
	}
	if parts[2] != "tree" {
		return "", "", "", "", fmt.Errorf("invalid URL format: expected OWNER/REPO/tree/REF/PATH")
	}
	owner = parts[0]
	repo = parts[1]
	ref = parts[3]
	if len(parts) > 4 {
		path = strings.Join(parts[4:], "/")
	}
	return owner, repo, ref, path, nil
}

func DownloadFromURL(client *http.Client, url string) error {
	owner, repo, ref, path, err := ParseGitHubURL(url)
	if err != nil {
		return err
	}
	return Download(client, owner, repo, ref, path)
}

func Download(client *http.Client, owner, repo, ref, path string) error {
	ghClient := getClient(client)
	ctx := context.Background()
	opts := &github.RepositoryContentGetOptions{
		Ref: ref,
	}
	var localPath string
	if path == "" {
		localPath = repo
	} else {
		localPath = filepath.Base(path)
	}
	fileContent, directoryContent, _, err := ghClient.Repositories.GetContents(ctx, owner, repo, path, opts)
	if err != nil {
		return fmt.Errorf("failed to get repository contents: %v", err)
	}
	if fileContent != nil {
		return downloadFile(client, fileContent, localPath)
	}
	if directoryContent != nil {
		if err := os.MkdirAll(localPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
		return downloadDirectory(ghClient, client, ctx, owner, repo, ref, path, localPath)
	}
	return fmt.Errorf("no content found at path")
}

func downloadFile(client *http.Client, content *github.RepositoryContent, localPath string) error {
	if content.DownloadURL == nil || *content.DownloadURL == "" {
		if content.Content == nil {
			return fmt.Errorf("file content is empty")
		}
		decoded, err := base64.StdEncoding.DecodeString(*content.Content)
		if err != nil {
			return fmt.Errorf("failed to decode base64 content: %v", err)
		}
		if err := os.WriteFile(localPath, decoded, 0644); err != nil {
			return fmt.Errorf("failed to write file: %v", err)
		}
		fmt.Printf("Downloaded %s %s %s\n", u.FDebug(*content.Name), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(localPath))
		return nil
	}

	req, err := http.NewRequest("GET", *content.DownloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	fmt.Printf("Downloaded %s %s %s\n", u.FDebug(*content.Name), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(localPath))
	return nil
}

func downloadDirectory(ghClient *github.Client, httpClient *http.Client, ctx context.Context, owner, repo, ref, repoPath, localPath string) error {
	opts := &github.RepositoryContentGetOptions{
		Ref: ref,
	}
	_, contents, _, err := ghClient.Repositories.GetContents(ctx, owner, repo, repoPath, opts)
	if err != nil {
		return fmt.Errorf("failed to get directory contents: %v", err)
	}
	if contents == nil {
		return fmt.Errorf("directory is empty")
	}
	for _, item := range contents {
		if item.Name == nil {
			continue
		}
		itemLocalPath := filepath.Join(localPath, *item.Name)
		itemRepoPath := repoPath
		if repoPath != "" {
			itemRepoPath = repoPath + "/" + *item.Name
		} else {
			itemRepoPath = *item.Name
		}
		if item.Type == nil {
			continue
		}
		switch *item.Type {
		case "file":
			if err := downloadFile(httpClient, item, itemLocalPath); err != nil {
				log.Error().Err(err).Msgf("Failed to download file %s, skipping...", *item.Name)
				continue
			}
		case "dir":
			if err := os.MkdirAll(itemLocalPath, 0755); err != nil {
				log.Error().Err(err).Msgf("Failed to create directory %s, skipping...", itemLocalPath)
				continue
			}
			if err := downloadDirectory(ghClient, httpClient, ctx, owner, repo, ref, itemRepoPath, itemLocalPath); err != nil {
				log.Error().Err(err).Msgf("Failed to download directory %s, skipping...", *item.Name)
				continue
			}
		}
	}
	return nil
}
