package anbuGenerics

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"

	u "github.com/tanq16/anbu/utils"
)

func BulkRename(pattern string, replacement string, renameDirectories bool, dryRun bool) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid regex pattern")
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
				continue
			}
			placeholder := fmt.Sprintf("\\%d", i)
			newName = strings.ReplaceAll(newName, placeholder, match)
		}
		if oldName == newName {
			continue
		}
		if dryRun {
			fmt.Printf("Dry Run: Renaming %s %s %s\n", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(newName))
		} else {
			err := os.Rename(filepath.Join(currentDir, oldName), filepath.Join(currentDir, newName))
			if err != nil {
				u.PrintError(fmt.Sprintf("Failed to rename %s to %s: %v", oldName, newName, err))
				continue
			}
			fmt.Printf("Renamed: %s %s %s\n", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(newName))
		}
		renameCount++
	}
	fmt.Println()
	if renameCount == 0 {
		log.Warn().Msg("no items were renamed")
	} else {
		fmt.Printf("%s %s\n", u.FDebug("Operation completed:"), u.FSuccess(fmt.Sprintf("%d %s", renameCount, map[bool]string{true: "directories", false: "files"}[renameDirectories])))
	}
}
