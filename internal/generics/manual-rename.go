package anbuGenerics

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	u "github.com/tanq16/anbu/utils"
)

func ManualRename(includeDir bool, hidden bool, includeExtension bool) {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get current directory")
	}
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read directory")
	}
	var items []os.DirEntry
	for _, entry := range entries {
		if !hidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if includeDir && entry.IsDir() {
			items = append(items, entry)
		}
		if !entry.IsDir() {
			items = append(items, entry)
		}
	}
	if len(items) == 0 {
		log.Warn().Msg("no items found to rename")
		return
	}
	log.Debug().Int("count", len(items)).Msg("items to process")

	reader := bufio.NewReader(os.Stdin)
	renameCount := 0
	for _, entry := range items {
		oldName := entry.Name()
		fmt.Printf("%s %s ", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"]))
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Error().Err(err).Msg("failed to read input")
			if !u.GlobalDebugFlag {
				fmt.Print("\033[A\r\033[K")
			}
			fmt.Printf("%s %s %s %s\n", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"]), u.FDebug(oldName), u.FWarning("(skipped)"))
			continue
		}
		newName := strings.TrimSpace(input)
		if !u.GlobalDebugFlag {
			fmt.Print("\033[A\r\033[K")
		}
		if newName == "" || newName == oldName {
			fmt.Printf("%s %s %s %s\n", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"]), u.FDebug(oldName), u.FWarning("(skipped)"))
			continue
		}
		if !includeExtension && !entry.IsDir() {
			ext := filepath.Ext(oldName)
			if ext != "" && !strings.HasSuffix(newName, ext) {
				newName = newName + ext
			}
		}
		if oldName == newName {
			fmt.Printf("%s %s %s %s\n", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"]), u.FDebug(oldName), u.FWarning("(skipped)"))
			continue
		}
		oldPath := filepath.Join(currentDir, oldName)
		newPath := filepath.Join(currentDir, newName)
		log.Debug().Str("old", oldName).Str("new", newName).Msg("renaming")

		err = os.Rename(oldPath, newPath)
		if err != nil {
			log.Error().Err(err).Str("old", oldName).Str("new", newName).Msg("failed to rename")
			fmt.Printf("%s %s %s %s\n", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"]), u.FError(newName), u.FWarning("(failed)"))
			continue
		}
		fmt.Printf("%s %s %s\n", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(newName))
		renameCount++
	}
	fmt.Println()
	if renameCount == 0 {
		log.Warn().Msg("no items were renamed")
	}
	fmt.Printf("%s %s\n", u.FDebug("Operation completed:"), u.FSuccess(fmt.Sprintf("%d items renamed", renameCount)))
}
