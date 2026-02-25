package anbuGenerics

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"

	u "github.com/tanq16/anbu/internal/utils"
)

func BulkRename(pattern string, replacement string, renameDirectories bool, dryRun bool) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	renameCount := 0
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
			uuidStr := generateUUIDString()
			newName = strings.ReplaceAll(newName, "\\uuid", uuidStr)
		}
		if strings.Contains(newName, "\\suid") {
			suidStr := generateRUIDString(18)
			newName = strings.ReplaceAll(newName, "\\suid", suidStr)
		}
		if oldName == newName {
			continue
		}
		if dryRun {
			u.PrintGeneric(fmt.Sprintf("Dry Run: Renaming %s %s %s", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(newName)))
		} else {
			err := os.Rename(filepath.Join(currentDir, oldName), filepath.Join(currentDir, newName))
			if err != nil {
				u.PrintError(fmt.Sprintf("Failed to rename %s to %s", oldName, newName), err)
				continue
			}
			u.PrintGeneric(fmt.Sprintf("Renamed: %s %s %s", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(newName)))
		}
		renameCount++
	}
	u.LineBreak()
	if renameCount == 0 {
		u.PrintWarn("no items were renamed", nil)
	} else {
		u.PrintGeneric(fmt.Sprintf("%s %s", u.FDebug("Operation completed:"), u.FSuccess(fmt.Sprintf("%d %s", renameCount, map[bool]string{true: "directories", false: "files"}[renameDirectories]))))
	}
	return nil
}

// same logic as GenerateUUIDString but returns instead of printing
func generateUUIDString() string {
	uuid, _ := uuid.NewRandom()
	return uuid.String()
}

// same logic as GenerateRUIDString but returns instead of printing
func generateRUIDString(length int) string {
	if length <= 0 || length > 30 {
		u.PrintWarn("length must be between 1 and 30; using 18", nil)
		length = 18
	}
	uuid, _ := uuid.NewRandom()
	// remove version and variant bits from UUID
	shortUUID := uuid.String()[0:8] + uuid.String()[9:13] + uuid.String()[15:18] + uuid.String()[20:23] + uuid.String()[24:]
	return shortUUID[:length]
}
