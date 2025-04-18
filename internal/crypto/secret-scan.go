package anbuCrypto

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/tanq16/anbu/utils"
)

type SecretMatch struct {
	Type  string
	Match string
	File  string
	Line  int
}

type SecretMatches struct {
	matches []SecretMatch
	mutex   sync.Mutex
}

func (s SecretMatch) Equals(other SecretMatch) bool {
	return s.Type == other.Type && s.Match == other.Match
}

func (sm *SecretMatches) Contains(match SecretMatch) bool {
	for _, m := range sm.matches {
		if m.Equals(match) {
			return true
		}
	}
	return false
}

func (sm *SecretMatches) Add(match SecretMatch) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	if !sm.Contains(match) {
		sm.matches = append(sm.matches, match)
	}
}

func ScanSecretsInPath(path string, printFalsePositives bool) error {
	logger := utils.GetLogger("secrets")
	rules := make([]struct {
		Name    string
		Pattern *regexp.Regexp
	}, len(secretRules))
	for i, rule := range secretRules {
		logger.Debug().Str("name", rule.Name).Str("pattern", rule.Pattern).Msg("compiling pattern")
		pattern, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern for %s: %v", rule.Name, err)
		}
		rules[i].Name = rule.Name
		rules[i].Pattern = pattern
	}

	// Collect files to scan
	var filesToScan []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && !isLikelyBinary(path) && info.Size() <= maxSize {
			filesToScan = append(filesToScan, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	logger.Debug().Int("file_count", len(filesToScan)).Msg("files to scan")
	matches := &SecretMatches{
		matches: make([]SecretMatch, 0),
	}

	// Create scanner pool
	var wg sync.WaitGroup
	workers := make(chan struct{}, 20) // Limit to 20
	errChan := make(chan error, len(filesToScan))
	for _, file := range filesToScan {
		if file == "" {
			continue
		}
		if strings.Contains(file, "node_modules") || strings.Contains(file, ".git") {
			continue
		}
		wg.Add(1)
		workers <- struct{}{}
		go func(filepath string) {
			defer wg.Done()
			defer func() { <-workers }()
			if err := scanFile(filepath, rules, matches); err != nil {
				errChan <- fmt.Errorf("error scanning file %s: %v", filepath, err)
			}
			logger.Debug().Str("file", filepath).Msg("scanned")
		}(file)
	}
	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		var errMsgs []string
		for err := range errChan {
			errMsgs = append(errMsgs, err.Error())
		}
		return fmt.Errorf("scanning errors occurred: %s", strings.Join(errMsgs, "; "))
	}

	table := utils.MarkdownTable{
		Headers: []string{"Type", "Match", "File", "Line"},
		Rows:    make([][]string, 0),
	}
	falsepTable := utils.MarkdownTable{
		Headers: []string{"Match", "File", "Line"},
		Rows:    make([][]string, 0),
	}
	for _, match := range matches.matches {
		if match.Type == "Generic Secrets & Keys" {
			falsepTable.Rows = append(falsepTable.Rows, []string{
				match.Match,
				match.File,
				fmt.Sprintf("%d", match.Line),
			})
			continue
		}
		table.Rows = append(table.Rows, []string{
			match.Type,
			match.Match,
			match.File,
			fmt.Sprintf("%d", match.Line),
		})
	}
	if len(table.Rows) > 0 {
		table.OutMDPrint(false)
	} else {
		fmt.Println(utils.OutSuccess("No secrets found."))
	}
	if len(falsepTable.Rows) > 0 && printFalsePositives {
		fmt.Println(utils.OutWarning("\nFalse positive matches detected. Please verify the results for the following:"))
		falsepTable.OutMDPrint(false)
	}
	return nil
}

func scanFile(filepath string, rules []struct {
	Name    string
	Pattern *regexp.Regexp
}, matches *SecretMatches) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	buf := make([]byte, maxSize)
	scanner.Buffer(buf, maxSize)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		for _, rule := range rules {
			if foundMatches := rule.Pattern.FindStringSubmatch(line); len(foundMatches) > 0 {
				// Get the actual match (first group if exists, otherwise full match)
				matchStr := foundMatches[0]
				if len(foundMatches) > 1 && foundMatches[1] != "" {
					matchStr = foundMatches[1]
				}
				if len(matchStr) > 50 {
					matchStr = matchStr[:47] + "..."
				}
				match := SecretMatch{
					Type:  rule.Name,
					Match: matchStr,
					File:  filepath,
					Line:  lineNum,
				}
				matches.Add(match)
			}
		}
	}
	return scanner.Err()
}

func isLikelyBinary(path string) bool {
	// Check extension first
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".pdf": true, ".zip": true, ".tar": true, ".gz": true,
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".bin": true, ".class": true, ".pyc": true, ".obj": true,
		".o": true, ".a": true, ".lib": true, ".jar": true,
	}
	if binaryExts[ext] {
		return true
	}
	// Check file content
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	buf := make([]byte, 512) // Read first 512 bytes
	n, err := file.Read(buf)
	if err != nil {
		return false
	}
	// Check for null bytes and non-printable chars as heuristic
	nullCount := 0
	nonPrintable := 0
	for i := range n {
		if buf[i] == 0 {
			nullCount++
		} else if buf[i] < 32 && buf[i] != '\n' && buf[i] != '\r' && buf[i] != '\t' {
			nonPrintable++
		}
	}
	return nullCount > 0 || float64(nonPrintable)/float64(n) > 0.3
}
