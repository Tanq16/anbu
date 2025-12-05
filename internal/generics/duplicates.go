package anbuGenerics

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
)

const sizeThreshold = 100 * 1024 * 1024

func FindDuplicates(recursive bool) {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get current directory")
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
			log.Fatal().Err(err).Msg("failed to walk directory")
		}
	} else {
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to read directory")
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
			hash, err := computeFileHash(file)
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
		return
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
	}

	if len(unhashedDuplicates) > 0 {
		fmt.Println()
		u.PrintWarning("Unhashed duplicates due to huge size:")
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
	}
}

func computeFileHash(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}
