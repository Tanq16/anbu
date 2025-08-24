package anbuCrypto

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	u "github.com/tanq16/anbu/utils"
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

func ScanSecretsInPath(path string, printFalsePositives bool) {
	logger := u.NewManager()
	logger.StartDisplay()
	defer logger.StopDisplay()
	funcID := logger.Register("secrets-scanner")
	logger.SetMessage(funcID, "Scanning for secrets...")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.ReportError(funcID, err)
		logger.SetMessage(funcID, "Path doesn't exist")
		return
	}

	rules := make([]struct {
		Name    string
		Pattern *regexp.Regexp
	}, len(secretRules))
	for i, rule := range secretRules {
		logger.SetMessage(funcID, "Compiling patterns...")
		pattern, err := regexp.Compile(rule.Pattern)
		if err != nil {
			logger.ReportError(funcID, err)
			logger.SetMessage(funcID, fmt.Sprintf("Failed to compile pattern %s", rule.Name))
			return
		}
		rules[i].Name = rule.Name
		rules[i].Pattern = pattern
	}

	// Collect files to scan
	logger.SetMessage(funcID, "Collecting files to scan...")
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
		logger.ReportError(funcID, err)
		logger.SetMessage(funcID, "Error collecting files")
		return
	}
	logger.SetMessage(funcID, fmt.Sprintf("Found %d files to scan", len(filesToScan)))
	matches := &SecretMatches{
		matches: make([]SecretMatch, 0),
	}

	// Create scanner pool
	logger.SetMessage(funcID, "Scanning files...")
	var wg sync.WaitGroup
	var progWg sync.WaitGroup
	progressChan := make(chan int64)
	progWg.Add(1)
	go func(progCh <-chan int64) {
		defer progWg.Done()
		var completed int64
		for i := range progCh {
			completed += i
			logger.AddProgressBarToStream(funcID, completed, int64(len(filesToScan)), fmt.Sprintf("Scanned %d files", completed))
		}
	}(progressChan)
	workers := make(chan struct{}, 30) // Limit to 30
	errChan := make(chan error, len(filesToScan))
	for _, file := range filesToScan {
		skipped := false
		for _, skipMatch := range secretSkips {
			if strings.Contains(file, skipMatch) {
				skipped = true
				break
			}
		}
		if skipped {
			progressChan <- 1
			continue
		}
		wg.Add(1)
		workers <- struct{}{}
		go func(filepath string, progCh chan<- int64) {
			defer wg.Done()
			defer func() { <-workers }()
			if err := scanFile(filepath, rules, matches); err != nil {
				errChan <- fmt.Errorf("error scanning file %s: %v", filepath, err)
			}
			progressChan <- 1
		}(file, progressChan)
	}
	wg.Wait()
	close(progressChan)
	close(errChan)
	progWg.Wait()
	if len(matches.matches) == 0 {
		logger.SetMessage(funcID, "No secrets found")
	} else {
		logger.SetMessage(funcID, fmt.Sprintf("Found %d potential secrets", len(matches.matches)))
	}

	if len(errChan) > 0 {
		var errMsgs []string
		for err := range errChan {
			errMsgs = append(errMsgs, err.Error())
		}
		logger.ReportError(funcID, fmt.Errorf("scanning errors occurred: %s", strings.Join(errMsgs, "; ")))
		logger.SetMessage(funcID, "Errors occurred during scanning")
	} else {
		logger.Complete(funcID, "Scanning completed successfully")
	}

	table := logger.RegisterTable("Primary Matches", []string{"Type", "Match", "File", "Line"})
	falsepTable := logger.RegisterTable("Generic Matches", []string{"Match", "File", "Line"})
	for _, match := range matches.matches {
		if match.Type == "Generic Secrets & Keys" {
			if !printFalsePositives {
				continue
			}
			falsepTable.Rows = append(falsepTable.Rows, []string{
				match.Match,
				match.File,
				fmt.Sprintf("%d", match.Line),
			})
			continue
		} else {
			table.Rows = append(table.Rows, []string{
				match.Type,
				match.Match,
				match.File,
				fmt.Sprintf("%d", match.Line),
			})
		}
	}
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
				if len(matchStr) > 40 {
					matchStr = matchStr[:37] + "..."
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
