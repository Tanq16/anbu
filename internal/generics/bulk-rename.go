package anbuGenerics

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	u "github.com/tanq16/anbu/utils"
)

// BulkRename renames files/directories based on a regex pattern and replacement string
func BulkRename(pattern string, replacement string, renameDirectories bool) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		u.PrintError(fmt.Sprintf("Invalid regex pattern: %v", err))
		return
	}
	currentDir, _ := os.Getwd()
	entries, _ := os.ReadDir(currentDir)

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
				continue // Skip the full match
			}
			placeholder := fmt.Sprintf("\\%d", i)
			newName = strings.ReplaceAll(newName, placeholder, match)
		}
		if oldName == newName {
			continue
		}
		err := os.Rename(filepath.Join(currentDir, oldName), filepath.Join(currentDir, newName))
		if err != nil {
			u.PrintError(fmt.Sprintf("Failed to rename %s to %s: %v", oldName, newName, err))
			continue
		}
		fmt.Printf("Renamed: %s %s %s\n", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(newName))
		renameCount++
	}

	// Provide a summary
	fmt.Println()
	if renameCount == 0 {
		u.PrintWarning("No items were renamed")
	} else {
		fmt.Printf("%s %s\n", u.FDebug("Operation completed: Renamed"),
			u.FSuccess(fmt.Sprintf("%d %s", renameCount, map[bool]string{true: "directories", false: "files"}[renameDirectories])))
	}
}
