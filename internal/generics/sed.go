package anbuGenerics

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"

	u "github.com/tanq16/anbu/utils"
)

func Sed(pattern string, replacement string, path string, dryRun bool) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid regex pattern")
	}

	info, err := os.Stat(path)
	if err != nil {
		log.Fatal().Err(err).Msgf("Path does not exist: %s", path)
	}

	var wg sync.WaitGroup
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
			wg.Add(1)
			go func(fp string) {
				defer wg.Done()
				if processFile(fp, re, replacement, dryRun) {
					mu.Lock()
					processedCount++
					mu.Unlock()
				}
			}(filePath)
			return nil
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to walk directory")
		}
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if processFile(path, re, replacement, dryRun) {
				mu.Lock()
				processedCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if processedCount == 0 {
		log.Warn().Msg("no files were processed")
	} else {
		fmt.Printf("%s %s\n", u.FDebug("Operation completed:"),
			u.FSuccess(fmt.Sprintf("%d file(s) processed", processedCount)))
	}
}

func processFile(filePath string, re *regexp.Regexp, replacement string, dryRun bool) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		u.PrintError(fmt.Sprintf("Failed to read file %s: %v", filePath, err))
		return false
	}

	originalContent := string(content)
	modifiedContent := replaceWithGroups(originalContent, re, replacement)

	if originalContent == modifiedContent {
		return false
	}

	if dryRun {
		fmt.Printf("Dry Run: %s\n", u.FDebug(filePath))
		fmt.Println(modifiedContent)
		fmt.Println()
	} else {
		err := os.WriteFile(filePath, []byte(modifiedContent), 0644)
		if err != nil {
			u.PrintError(fmt.Sprintf("Failed to write file %s: %v", filePath, err))
			return false
		}
		fmt.Printf("Modified: %s\n", u.FSuccess(filePath))
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
