package anbuGenerics

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
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

// same logic as GenerateUUIDString but returns instead of printing
func generateUUIDString() string {
	uuid, _ := uuid.NewRandom()
	return uuid.String()
}

// same logic as GenerateRUIDString but returns instead of printing
func generateRUIDString(length int) string {
	if length <= 0 || length > 30 {
		log.Warn().Msg("length must be between 1 and 30; using 18")
		length = 18
	}
	uuid, _ := uuid.NewRandom()
	// remove version and variant bits from UUID
	shortUUID := uuid.String()[0:8] + uuid.String()[9:13] + uuid.String()[15:18] + uuid.String()[20:23] + uuid.String()[24:]
	return shortUUID[:length]
}
