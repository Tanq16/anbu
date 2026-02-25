package anbuGenerics

import (
	"fmt"
	"os"
	"path/filepath"

	u "github.com/tanq16/anbu/internal/utils"
)

const sizeThreshold = 100 * 1024 * 1024

func FindDuplicates(recursive bool, delete bool) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	var files []string
	if recursive {
		err = filepath.WalkDir(currentDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk directory: %w", err)
		}
	} else {
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, filepath.Join(currentDir, entry.Name()))
			}
		}
	}

	sizeMap := make(map[int64][]string)
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		size := info.Size()
		sizeMap[size] = append(sizeMap[size], file)
	}

	var hashedDuplicates [][]string
	var unhashedDuplicates [][]string
	for size, fileList := range sizeMap {
		if len(fileList) < 2 {
			continue
		}
		if size >= sizeThreshold {
			unhashedDuplicates = append(unhashedDuplicates, fileList)
			continue
		}
		hashMap := make(map[string][]string)
		for _, file := range fileList {
			hash, err := u.ComputeFileHash(file)
			if err != nil {
				continue
			}
			hashMap[hash] = append(hashMap[hash], file)
		}
		for _, fileList := range hashMap {
			if len(fileList) >= 2 {
				hashedDuplicates = append(hashedDuplicates, fileList)
			}
		}
	}

	if len(hashedDuplicates) == 0 && len(unhashedDuplicates) == 0 {
		u.PrintInfo("No duplicate files found")
		return nil
	}
	if len(hashedDuplicates) > 0 {
		table := u.NewTable([]string{"Set ID", "Files"})
		for i, fileList := range hashedDuplicates {
			filesStr := ""
			for j, file := range fileList {
				relPath, err := filepath.Rel(currentDir, file)
				if err != nil {
					relPath = file
				}
				if j > 0 {
					filesStr += ", "
				}
				filesStr += relPath
			}
			table.Rows = append(table.Rows, []string{fmt.Sprintf("%d", i+1), filesStr})
		}
		table.PrintTable(false)
		if delete {
			for _, fileList := range hashedDuplicates {
				for i := 1; i < len(fileList); i++ {
					if err := os.Remove(fileList[i]); err != nil {
						relPath, _ := filepath.Rel(currentDir, fileList[i])
						u.PrintError(fmt.Sprintf("Failed to delete %s", relPath), err)
					} else {
						relPath, _ := filepath.Rel(currentDir, fileList[i])
						u.PrintGeneric(fmt.Sprintf("Deleted: %s", u.FSuccess(relPath)))
					}
				}
			}
		}
	}

	if len(unhashedDuplicates) > 0 {
		u.LineBreak()
		u.PrintWarn("Unhashed duplicates due to huge size:", nil)
		table := u.NewTable([]string{"Set ID", "Files"})
		startID := len(hashedDuplicates) + 1
		for i, fileList := range unhashedDuplicates {
			filesStr := ""
			for j, file := range fileList {
				relPath, err := filepath.Rel(currentDir, file)
				if err != nil {
					relPath = file
				}
				if j > 0 {
					filesStr += ", "
				}
				filesStr += relPath
			}
			table.Rows = append(table.Rows, []string{fmt.Sprintf("%d", startID+i), filesStr})
		}
		table.PrintTable(false)
		if delete {
			for _, fileList := range unhashedDuplicates {
				for i := 1; i < len(fileList); i++ {
					if err := os.Remove(fileList[i]); err != nil {
						relPath, _ := filepath.Rel(currentDir, fileList[i])
						u.PrintError(fmt.Sprintf("Failed to delete %s", relPath), err)
					} else {
						relPath, _ := filepath.Rel(currentDir, fileList[i])
						u.PrintGeneric(fmt.Sprintf("Deleted: %s", u.FSuccess(relPath)))
					}
				}
			}
		}
	}
	return nil
}
