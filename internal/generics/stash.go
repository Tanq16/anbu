package anbuGenerics

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
)

type StashType string

const (
	TypeFS   StashType = "fs"
	TypeText StashType = "text"
)

type StashEntry struct {
	ID           int       `json:"id"`
	Type         StashType `json:"type"`
	Name         string    `json:"name"`
	OriginalPath string    `json:"original_path"`
	BlobName     string    `json:"blob_name"`
	CreatedAt    time.Time `json:"created_at"`
}

type StashIndex struct {
	NextID  int          `json:"next_id"`
	Entries []StashEntry `json:"entries"`
}

func getStashDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	stashDir := filepath.Join(homeDir, ".anbu-stash")
	if err := os.MkdirAll(stashDir, 0755); err != nil {
		return "", err
	}
	blobsDir := filepath.Join(stashDir, "blobs")
	if err := os.MkdirAll(blobsDir, 0755); err != nil {
		return "", err
	}
	return stashDir, nil
}

func getIndexPath() (string, error) {
	stashDir, err := getStashDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stashDir, "index.json"), nil
}

func loadIndex() (*StashIndex, error) {
	indexPath, err := getIndexPath()
	if err != nil {
		return nil, err
	}
	index := &StashIndex{
		NextID:  1,
		Entries: []StashEntry{},
	}
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return index, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, index); err != nil {
		return nil, err
	}
	return index, nil
}

func saveIndex(index *StashIndex) error {
	indexPath, err := getIndexPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(indexPath, data, 0644)
}

func findEntryByID(index *StashIndex, id int) *StashEntry {
	for i := range index.Entries {
		if index.Entries[i].ID == id {
			return &index.Entries[i]
		}
	}
	return nil
}

func removeEntryByID(index *StashIndex, id int) bool {
	for i, entry := range index.Entries {
		if entry.ID == id {
			index.Entries = append(index.Entries[:i], index.Entries[i+1:]...)
			return true
		}
	}
	return false
}

func StashFS(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}
	stashDir, err := getStashDir()
	if err != nil {
		return err
	}
	blobUUID := uuid.New().String()
	blobName := blobUUID
	if info.IsDir() {
		blobName += ".zip"
	} else {
		blobName += ".zip"
	}
	blobPath := filepath.Join(stashDir, "blobs", blobName)
	if info.IsDir() {
		if err := zipDir(absPath, blobPath); err != nil {
			return fmt.Errorf("failed to zip directory: %w", err)
		}
	} else {
		if err := zipFile(absPath, blobPath); err != nil {
			return fmt.Errorf("failed to zip file: %w", err)
		}
	}
	index, err := loadIndex()
	if err != nil {
		return err
	}
	entry := StashEntry{
		ID:           index.NextID,
		Type:         TypeFS,
		Name:         filepath.Base(absPath),
		OriginalPath: absPath,
		BlobName:     blobName,
		CreatedAt:    time.Now(),
	}
	index.Entries = append(index.Entries, entry)
	index.NextID++
	if err := saveIndex(index); err != nil {
		return err
	}
	fmt.Printf("%s %s %s\n", u.FDebug(absPath), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(fmt.Sprintf("Stashed (ID: %d)", entry.ID)))
	return nil
}

func StashText(name string) error {
	input, err := io.ReadAll(os.Stdin)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read stdin: %w", err)
	}
	if len(input) == 0 {
		return fmt.Errorf("no input provided")
	}
	stashDir, err := getStashDir()
	if err != nil {
		return err
	}
	blobUUID := uuid.New().String()
	blobName := blobUUID + ".txt"
	blobPath := filepath.Join(stashDir, "blobs", blobName)
	if err := os.WriteFile(blobPath, input, 0644); err != nil {
		return fmt.Errorf("failed to write blob: %w", err)
	}
	index, err := loadIndex()
	if err != nil {
		return err
	}
	entry := StashEntry{
		ID:        index.NextID,
		Type:      TypeText,
		Name:      name,
		BlobName:  blobName,
		CreatedAt: time.Now(),
	}
	index.Entries = append(index.Entries, entry)
	index.NextID++
	if err := saveIndex(index); err != nil {
		return err
	}
	fmt.Printf("%s %s %s\n", u.FDebug(name), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(fmt.Sprintf("Stashed (ID: %d)", entry.ID)))
	return nil
}

func StashList() error {
	index, err := loadIndex()
	if err != nil {
		return err
	}
	if len(index.Entries) == 0 {
		fmt.Println("No stashed entries.")
		return nil
	}
	sort.Slice(index.Entries, func(i, j int) bool {
		return index.Entries[i].ID > index.Entries[j].ID
	})
	table := u.NewTable([]string{"ID", "Type", "Name", "Time Ago"})
	for _, entry := range index.Entries {
		timeAgo := time.Since(entry.CreatedAt)
		timeAgoStr := formatTimeAgo(timeAgo)
		table.Rows = append(table.Rows, []string{
			fmt.Sprintf("%d", entry.ID),
			string(entry.Type),
			entry.Name,
			timeAgoStr,
		})
	}
	table.PrintTable(false)
	return nil
}

func formatTimeAgo(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	if days < 30 {
		return fmt.Sprintf("%dd ago", days)
	}
	months := days / 30
	if months < 12 {
		return fmt.Sprintf("%dmo ago", months)
	}
	years := months / 12
	return fmt.Sprintf("%dy ago", years)
}

func StashApply(id int) error {
	index, err := loadIndex()
	if err != nil {
		return err
	}
	entry := findEntryByID(index, id)
	if entry == nil {
		return fmt.Errorf("stash ID not found")
	}
	stashDir, err := getStashDir()
	if err != nil {
		return err
	}
	blobPath := filepath.Join(stashDir, "blobs", entry.BlobName)
	if entry.Type == TypeText {
		data, err := os.ReadFile(blobPath)
		if err != nil {
			return fmt.Errorf("failed to read blob: %w", err)
		}
		fmt.Print(string(data))
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		if entry.OriginalPath != "" {
			baseName := filepath.Base(entry.OriginalPath)
			targetPath := filepath.Join(cwd, baseName)
			if _, err := os.Stat(targetPath); err == nil {
				return fmt.Errorf("file or folder already exists: %s", targetPath)
			}
		}
		if err := unzipDir(blobPath, cwd); err != nil {
			return fmt.Errorf("failed to unzip: %w", err)
		}
		fmt.Printf("%s %s %s\n", u.FDebug(cwd), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(fmt.Sprintf("Applied stash (ID: %d)", id)))
	}
	return nil
}

func StashPop(id int) error {
	if err := StashApply(id); err != nil {
		return err
	}
	index, err := loadIndex()
	if err != nil {
		return err
	}
	entry := findEntryByID(index, id)
	if entry == nil {
		return fmt.Errorf("stash ID not found")
	}
	stashDir, err := getStashDir()
	if err != nil {
		return err
	}
	blobPath := filepath.Join(stashDir, "blobs", entry.BlobName)
	if err := os.Remove(blobPath); err != nil {
		log.Warn().Err(err).Msg("failed to remove blob file")
	}
	if !removeEntryByID(index, id) {
		return fmt.Errorf("failed to remove entry from index")
	}
	if len(index.Entries) == 0 {
		index.NextID = 1
	}
	if err := saveIndex(index); err != nil {
		return err
	}
	fmt.Printf("%s %s %s\n", u.FDebug(fmt.Sprintf("%d", id)), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(fmt.Sprintf("Popped stash (ID: %d)", id)))
	return nil
}

func StashClear(id int) error {
	index, err := loadIndex()
	if err != nil {
		return err
	}
	entry := findEntryByID(index, id)
	if entry == nil {
		return fmt.Errorf("stash ID not found")
	}
	stashDir, err := getStashDir()
	if err != nil {
		return err
	}
	blobPath := filepath.Join(stashDir, "blobs", entry.BlobName)
	if err := os.Remove(blobPath); err != nil {
		log.Warn().Err(err).Msg("failed to remove blob file")
	}
	if !removeEntryByID(index, id) {
		return fmt.Errorf("failed to remove entry from index")
	}
	if len(index.Entries) == 0 {
		index.NextID = 1
	}
	if err := saveIndex(index); err != nil {
		return err
	}
	fmt.Printf("%s %s %s\n", u.FDebug(fmt.Sprintf("%d", id)), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(fmt.Sprintf("Cleared stash (ID: %d)", id)))
	return nil
}

func zipDir(source, destZip string) error {
	zipFile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	archive := zip.NewWriter(zipFile)
	defer archive.Close()
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(source, path)
		header.Name = filepath.ToSlash(relPath)
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}
		writer, _ := archive.CreateHeader(header)
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
}

func unzipDir(srcZip, destDir string) error {
	r, err := zip.OpenReader(srcZip)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			continue
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
	}
	return nil
}

func zipFile(source, destZip string) error {
	zipFile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	archive := zip.NewWriter(zipFile)
	defer archive.Close()
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.Base(source)
	header.Method = zip.Deflate
	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(writer, file)
	return err
}
