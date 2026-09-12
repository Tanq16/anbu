package ghrelease

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"net/http"
	"regexp"
	"runtime"
	"strings"

	"github.com/tanq16/anbu/internal/download"
)

type Asset struct {
	Name string
	Size int64
	URL  string
}

type Release struct {
	Tag    string
	Assets []Asset
}

var assetSelectMap = map[string][]string{
	"linuxamd64":   {"linux-amd64", "linux_amd64", "linux-x86_64", "linux-x86-64", "linux_x86_64", "linux_x86-64", "amd64-linux", "x86_64-linux", "x86-64-linux", "amd64_linux", "x86_64_linux", "x86-64_linux"},
	"linuxarm64":   {"linux-arm64", "linux_arm64", "linux-aarch64", "linux_aarch64", "arm64-linux", "aarch64-linux", "arm64_linux", "aarch64_linux"},
	"windowsamd64": {"windows-amd64", "windows_amd64", "windows-x86_64", "windows-x86-64", "windows_x86_64", "windows_x86-64", "amd64-windows", "x86_64-windows", "x86-64-windows", "amd64_windows", "x86_64_windows", "x86-64_windows"},
	"windowsarm64": {"windows-arm64", "windows_arm64", "windows-aarch64", "windows_aarch64", "arm64-windows", "aarch64-windows", "arm64_windows", "aarch64_windows"},
	"darwinamd64":  {"darwin-amd64", "darwin_amd64", "darwin-x86_64", "darwin-x86-64", "darwin_x86_64", "darwin_x86-64", "amd64-darwin", "x86_64-darwin", "x86-64-darwin", "amd64_darwin", "x86_64_darwin", "x86-64_darwin"},
	"darwinarm64":  {"darwin-arm64", "darwin_arm64", "darwin-aarch64", "darwin_aarch64", "arm64-darwin", "aarch64-darwin", "arm64_darwin", "aarch64_darwin"},
}

var repoPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^https?://github\.com/([^/]+)/([^/]+)/?.*$`),
	regexp.MustCompile(`^github\.com/([^/]+)/([^/]+)/?.*$`),
	regexp.MustCompile(`^([^/]+)/([^/]+)$`),
}

var ignoredAssets = []string{
	"license", "readme", "changelog", "checksums", "sha256checksum", ".sha256",
}

func ParseRepo(raw string) (string, string, error) {
	raw = strings.TrimSuffix(strings.TrimSpace(raw), "/")
	for _, pattern := range repoPatterns {
		matches := pattern.FindStringSubmatch(raw)
		if len(matches) >= 3 {
			return matches[1], matches[2], nil
		}
	}
	return "", "", fmt.Errorf("invalid GitHub repository format: %s", raw)
}

func Latest(ctx context.Context, owner, repo string, client *download.Client) (Release, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return Release{}, fmt.Errorf("error creating API request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("error making API request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var payload struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			Size               int64  `json:"size"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.UnmarshalRead(resp.Body, &payload); err != nil {
		return Release{}, fmt.Errorf("error decoding API response: %w", err)
	}
	if len(payload.Assets) == 0 {
		return Release{}, fmt.Errorf("no assets found in the release")
	}
	assets := make([]Asset, 0, len(payload.Assets))
	for _, a := range payload.Assets {
		assets = append(assets, Asset{Name: a.Name, Size: a.Size, URL: a.BrowserDownloadURL})
	}
	return Release{Tag: payload.TagName, Assets: assets}, nil
}

func SelectForPlatform(assets []Asset) (Asset, bool) {
	platformKey := runtime.GOOS + runtime.GOARCH
	for _, asset := range assets {
		if ignored(asset.Name) {
			continue
		}
		lower := strings.ToLower(asset.Name)
		for _, key := range assetSelectMap[platformKey] {
			if strings.Contains(lower, key) {
				return asset, true
			}
		}
	}

	osKeys, archKeys := platformKeywords()
	conflicts := conflictingKeywords()
	bestScore := -1000
	var best Asset
	found := false
	for _, asset := range assets {
		if ignored(asset.Name) {
			continue
		}
		lower := strings.ToLower(asset.Name)
		score := 0
		for _, key := range osKeys {
			if strings.Contains(lower, key) {
				score += 5
				break
			}
		}
		for _, key := range archKeys {
			if strings.Contains(lower, key) {
				score += 5
				break
			}
		}
		for _, key := range conflicts {
			if strings.Contains(lower, key) {
				score -= 100
			}
		}
		if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".zip") {
			score += 1
		}
		if score > bestScore {
			bestScore = score
			best = asset
			found = true
		}
	}
	if found && bestScore > 0 {
		return best, true
	}
	return Asset{}, false
}

func FindAsset(assets []Asset, name string) (Asset, bool) {
	for _, asset := range assets {
		if asset.Name == name {
			return asset, true
		}
	}
	lower := strings.ToLower(name)
	for _, asset := range assets {
		if strings.EqualFold(asset.Name, name) || strings.Contains(strings.ToLower(asset.Name), lower) {
			return asset, true
		}
	}
	return Asset{}, false
}

func ignored(name string) bool {
	lower := strings.ToLower(name)
	for _, token := range ignoredAssets {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}

func platformKeywords() ([]string, []string) {
	var osKeys, archKeys []string
	switch runtime.GOOS {
	case "linux":
		osKeys = []string{"linux", "gnu"}
	case "windows":
		osKeys = []string{"windows", "win", ".exe"}
	case "darwin":
		osKeys = []string{"darwin", "mac", "apple", "osx"}
	}
	switch runtime.GOARCH {
	case "amd64":
		archKeys = []string{"amd64", "x86_64", "x86-64", "x64", "64-bit", "64bit"}
	case "arm64":
		archKeys = []string{"arm64", "aarch64"}
	}
	return osKeys, archKeys
}

func conflictingKeywords() []string {
	var conflicts []string
	if runtime.GOOS != "linux" {
		conflicts = append(conflicts, "linux", "gnu")
	}
	if runtime.GOOS != "windows" {
		conflicts = append(conflicts, "windows", "win", ".exe")
	}
	if runtime.GOOS != "darwin" {
		conflicts = append(conflicts, "darwin", "mac", "apple", "osx")
	}
	if runtime.GOARCH != "amd64" {
		conflicts = append(conflicts, "amd64", "x86_64", "x86-64", "x64")
	}
	if runtime.GOARCH != "arm64" {
		conflicts = append(conflicts, "arm64", "aarch64")
	}
	if runtime.GOARCH != "386" {
		conflicts = append(conflicts, "i386", "386", "x86_32", "x86-32")
	}
	if runtime.GOARCH != "arm" {
		conflicts = append(conflicts, "armv6", "armv7", "arm32")
	}
	return conflicts
}
