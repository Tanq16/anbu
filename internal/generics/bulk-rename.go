package anbuGenerics

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type RenameResult struct {
	Old string
	New string
	Err error
}

func BulkRename(pattern string, replacement string, renameDirectories bool, dryRun bool) ([]RenameResult, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var results []RenameResult
	for _, entry := range entries {
		if renameDirectories && !entry.IsDir() {
			continue
		}
		if !renameDirectories && entry.IsDir() {
			continue
		}
		oldName := entry.Name()
		matches := re.FindStringSubmatch(oldName)
		if matches == nil {
			continue
		}
		newName := replacement
		for i, match := range matches {
			if i == 0 {
				continue
			}
			placeholder := fmt.Sprintf("\\%d", i)
			newName = strings.ReplaceAll(newName, placeholder, match)
		}
		if strings.Contains(newName, "\\uuid") {
			uuidStr, err := GenerateUUIDString(true)
			if err != nil {
				return results, err
			}
			newName = strings.ReplaceAll(newName, "\\uuid", uuidStr)
		}
		if strings.Contains(newName, "\\suid") {
			suidStr, err := GenerateRUIDString(18)
			if err != nil {
				return results, err
			}
			newName = strings.ReplaceAll(newName, "\\suid", suidStr)
		}
		if oldName == newName {
			continue
		}
		result := RenameResult{Old: oldName, New: newName}
		if !dryRun {
			if err := os.Rename(filepath.Join(currentDir, oldName), filepath.Join(currentDir, newName)); err != nil {
				result.Err = err
			}
		}
		results = append(results, result)
	}
	return results, nil
}
