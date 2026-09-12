package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func simple(ctx context.Context, url, outputPath string, client *Client, progress io.Writer) error {
	tempDir := filepath.Join(filepath.Dir(outputPath), tempDirName)
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return fmt.Errorf("error creating temp directory: %w", err)
	}
	tempOutputPath := filepath.Join(tempDir, filepath.Base(outputPath)) + ".part"

	var reported int64
	reconcile := func() {
		currentSize := int64(0)
		if fi, err := os.Stat(tempOutputPath); err == nil {
			currentSize = fi.Size()
		}
		if delta := currentSize - reported; delta != 0 {
			addProgress(progress, delta)
			reported = currentSize
		}
	}
	reconcile()

	var lastErr error
	for retry := range maxRetries {
		if retry > 0 {
			time.Sleep(time.Duration(retry+1) * 500 * time.Millisecond)
			reconcile()
		}
		err := simpleAttempt(ctx, url, tempOutputPath, client, progress, &reported)
		if err != nil {
			lastErr = err
			reconcile()
			continue
		}
		if err := os.Rename(tempOutputPath, outputPath); err != nil {
			return fmt.Errorf("error renaming (finalizing) output file: %w", err)
		}
		removeTempDirIfEmpty(tempDir)
		return nil
	}
	return fmt.Errorf("download failed after %d retries: %w", maxRetries, lastErr)
}

func simpleAttempt(ctx context.Context, url, tempOutputPath string, client *Client, progress io.Writer, reported *int64) error {
	var resumeOffset int64
	fileMode := os.O_CREATE | os.O_WRONLY
	if fileInfo, err := os.Stat(tempOutputPath); err == nil {
		resumeOffset = fileInfo.Size()
		fileMode |= os.O_APPEND
	} else {
		fileMode |= os.O_TRUNC
	}

	outFile, err := os.OpenFile(tempOutputPath, fileMode, 0o644)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer func() {
		if outFile != nil {
			outFile.Close()
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("error creating GET request: %w", err)
	}
	if resumeOffset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeOffset))
	}
	req.Header.Set("Connection", "keep-alive")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error executing GET request: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resumeOffset > 0 && resp.StatusCode == http.StatusPartialContent:
	case resumeOffset > 0 && resp.StatusCode == http.StatusOK:
		outFile.Close()
		outFile = nil
		outFile, err = os.OpenFile(tempOutputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return fmt.Errorf("error creating output file: %w", err)
		}
		if *reported > 0 {
			addProgress(progress, -*reported)
			*reported = 0
		}
		resumeOffset = 0
	case resumeOffset == 0 && resp.StatusCode == http.StatusOK:
	case resumeOffset == 0 && resp.StatusCode == http.StatusPartialContent:
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if _, err := io.Copy(outFile, io.TeeReader(resp.Body, countingWriter{w: progress, n: reported})); err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	if err := outFile.Sync(); err != nil {
		return err
	}
	return nil
}

type countingWriter struct {
	w io.Writer
	n *int64
}

func (c countingWriter) Write(p []byte) (int, error) {
	if c.w != nil {
		if _, err := c.w.Write(p); err != nil {
			return 0, err
		}
	}
	if c.n != nil {
		*c.n += int64(len(p))
	}
	return len(p), nil
}
