package anbuGenerics

// import (
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"regexp"
// 	"strings"

// 	"github.com/tanq16/anbu/utils"
// )

// func BulkRename(pattern string, replacement string, renameDirectories bool) error {

// 	re, err := regexp.Compile(pattern)
// 	if err != nil {
// 		return fmt.Errorf("invalid regular expression: %w", err)
// 	}
// 	currentDir, err := os.Getwd()
// 	if err != nil {
// 		return fmt.Errorf("failed to get current directory: %w", err)
// 	}
// 	entries, err := os.ReadDir(currentDir)
// 	if err != nil {
// 		return fmt.Errorf("failed to read directory: %w", err)
// 	}

// 	renameCount := 0
// 	for _, entry := range entries {
// 		// Skip based on whether it's handling files or directories
// 		if renameDirectories && !entry.IsDir() {
// 			continue
// 		}
// 		if !renameDirectories && entry.IsDir() {
// 			continue
// 		}
// 		oldName := entry.Name()

// 		matches := re.FindStringSubmatch(oldName)
// 		if matches == nil {
// 			continue
// 		}
// 		newName := replacement
// 		for i, match := range matches {
// 			if i == 0 {
// 				continue // Skip the full match
// 			}
// 			placeholder := fmt.Sprintf("\\%d", i)
// 			newName = strings.Replace(newName, placeholder, match, -1)
// 		}

// 		// If nothing changed, skip, otherwise rename
// 		if oldName == newName {
// 			continue
// 		}
// 		err := os.Rename(
// 			filepath.Join(currentDir, oldName),
// 			filepath.Join(currentDir, newName),
// 		)
// 		if err != nil {
// 			logger.Debug().Str("old", oldName).Str("new", newName).Err(err).Msg("failed to rename")
// 			continue
// 		}
// 		logger.Debug().Str("old", oldName).Str("new", newName).Msg("renamed successfully")
// 		fmt.Printf("%s %s %s %s\n", utils.OutSuccess("Renamed:"), utils.OutDebug(oldName), utils.OutSuccess("→"), utils.OutInfo(newName))
// 		renameCount++
// 	}

// 	// Provide a summary
// 	fmt.Println()
// 	if renameCount == 0 {
// 		fmt.Println(utils.OutWarning("No items were renamed."))
// 	} else {
// 		fmt.Printf("%s %s\n", utils.OutDetail("Operation completed:"), utils.OutSuccess(fmt.Sprintf("%d %s renamed", renameCount, map[bool]string{true: "directories", false: "files"}[renameDirectories])))
// 	}
// 	return nil
// }
