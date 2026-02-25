package anbuGenerics

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	u "github.com/tanq16/anbu/internal/utils"
	"golang.org/x/sync/errgroup"
)

func Sed(pattern string, replacement string, path string, dryRun bool) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %s: %w", path, err)
	}

	g := new(errgroup.Group)
	g.SetLimit(30)
	var mu sync.Mutex
	processedCount := 0
	if info.IsDir() {
		err := filepath.Walk(path, func(filePath string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fileInfo.IsDir() {
				return nil
			}
			g.Go(func() error {
				if processFile(filePath, re, replacement, dryRun) {
					mu.Lock()
					processedCount++
					mu.Unlock()
				}
				return nil
			})
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk directory: %w", err)
		}
	} else {
		g.Go(func() error {
			if processFile(path, re, replacement, dryRun) {
				mu.Lock()
				processedCount++
				mu.Unlock()
			}
			return nil
		})
	}

	g.Wait()
	if processedCount == 0 {
		u.PrintWarn("no files were processed", nil)
	} else {
		u.PrintGeneric(fmt.Sprintf("%s %s", u.FDebug("Operation completed:"), u.FSuccess(fmt.Sprintf("%d file(s) processed", processedCount))))
	}
	return nil
}

func processFile(filePath string, re *regexp.Regexp, replacement string, dryRun bool) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		u.PrintError("Failed to read file", err)
		return false
	}
	originalContent := string(content)
	modifiedContent := replaceWithGroups(originalContent, re, replacement)
	if originalContent == modifiedContent {
		return false
	}

	if dryRun {
		u.PrintGeneric(fmt.Sprintf("Dry Run: %s", u.FDebug(filePath)))
		u.PrintGeneric(modifiedContent)
		u.LineBreak()
	} else {
		err := os.WriteFile(filePath, []byte(modifiedContent), 0644)
		if err != nil {
			u.PrintError("Failed to write file", err)
			return false
		}
		u.PrintGeneric(fmt.Sprintf("Modified: %s", u.FSuccess(filePath)))
	}
	return true
}

func replaceWithGroups(content string, re *regexp.Regexp, replacement string) string {
	var result strings.Builder
	lastIndex := 0
	allIndices := re.FindAllStringSubmatchIndex(content, -1)
	allSubmatches := re.FindAllStringSubmatch(content, -1)

	for i, indices := range allIndices {
		if len(indices) < 2 {
			continue
		}
		result.WriteString(content[lastIndex:indices[0]])
		if i < len(allSubmatches) && len(allSubmatches[i]) > 0 {
			repl := replacement
			for j := 1; j < len(allSubmatches[i]); j++ {
				placeholder := fmt.Sprintf("\\%d", j)
				repl = strings.ReplaceAll(repl, placeholder, allSubmatches[i][j])
			}
			result.WriteString(repl)
		}
		lastIndex = indices[1]
	}
	result.WriteString(content[lastIndex:])
	return result.String()
}
