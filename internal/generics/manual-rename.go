package anbuGenerics

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	u "github.com/tanq16/anbu/internal/utils"
)

func ManualRename(includeDir bool, hidden bool, includeExtension bool) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
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
		u.PrintWarn("no items found to rename", nil)
		return nil
	}
	log.Debug().Str("package", "generics").Int("count", len(items)).Msg("items to process")

	reader := bufio.NewReader(os.Stdin)
	renameCount := 0
	for _, entry := range items {
		oldName := entry.Name()
		u.PrintGeneric(fmt.Sprintf("%s %s ", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"])))
		input, err := reader.ReadString('\n')
		if err != nil {
			u.PrintError("failed to read input", err)
			if !u.GlobalDebugFlag {
				u.ClearTerminal(1)
			}
			u.PrintGeneric(fmt.Sprintf("%s %s %s %s", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"]), u.FDebug(oldName), u.FWarning("(skipped)")))
			continue
		}
		newName := strings.TrimSpace(input)
		if !u.GlobalDebugFlag {
			fmt.Print("\033[A\r\033[K")
		}
		if newName == "" || newName == oldName {
			u.PrintGeneric(fmt.Sprintf("%s %s %s %s", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"]), u.FDebug(oldName), u.FWarning("(skipped)")))
			continue
		}
		if !includeExtension && !entry.IsDir() {
			ext := filepath.Ext(oldName)
			if ext != "" && !strings.HasSuffix(newName, ext) {
				newName = newName + ext
			}
		}
		if oldName == newName {
			u.PrintGeneric(fmt.Sprintf("%s %s %s %s", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"]), u.FDebug(oldName), u.FWarning("(skipped)")))
			continue
		}
		oldPath := filepath.Join(currentDir, oldName)
		newPath := filepath.Join(currentDir, newName)
		log.Debug().Str("package", "generics").Str("old", oldName).Str("new", newName).Msg("renaming")

		err = os.Rename(oldPath, newPath)
		if err != nil {
			u.PrintError("failed to rename", err)
			u.PrintGeneric(fmt.Sprintf("%s %s %s %s", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"]), u.FError(newName), u.FWarning("(failed)")))
			continue
		}
		u.PrintGeneric(fmt.Sprintf("%s %s %s", u.FDebug(oldName), u.FInfo(u.StyleSymbols["arrow"]), u.FSuccess(newName)))
		renameCount++
	}
	u.LineBreak()
	if renameCount == 0 {
		u.PrintWarn("no items were renamed", nil)
	}
	u.PrintGeneric(fmt.Sprintf("%s %s", u.FDebug("Operation completed:"), u.FSuccess(fmt.Sprintf("%d items renamed", renameCount))))
	return nil
}
