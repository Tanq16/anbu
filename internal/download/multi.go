package download

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type chunk struct {
	id         int
	start      int64
	end        int64
	downloaded int64
	completed  bool
	path       string
}

func multi(ctx context.Context, url, outputPath string, connections int, client *Client, fileSize int64, progress io.Writer) error {
	tempDir := filepath.Join(filepath.Dir(outputPath), tempDirName)
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return fmt.Errorf("error creating temp directory: %w", err)
	}

	chunks := splitChunks(fileSize, connections)
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(connections)
	var mu sync.Mutex
	for i := range chunks {
		g.Go(func() error {
			return downloadChunk(ctx, url, outputPath, &chunks[i], client, tempDir, progress, &mu)
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	for _, c := range chunks {
		if !c.completed {
			return fmt.Errorf("download incomplete: chunk %d failed", c.id)
		}
	}
	return assemble(outputPath, fileSize, chunks)
}

func splitChunks(fileSize int64, connections int) []chunk {
	chunkSize := fileSize / int64(connections)
	var chunks []chunk
	var pos int64
	for i := range connections {
		start := pos
		end := start + chunkSize - 1
		if i == connections-1 {
			end = fileSize - 1
		}
		end = min(end, fileSize-1)
		if end >= start {
			chunks = append(chunks, chunk{id: i, start: start, end: end})
		}
		pos = end + 1
	}
	return chunks
}

func downloadChunk(ctx context.Context, url, outputPath string, ch *chunk, client *Client, tempDir string, progress io.Writer, mu *sync.Mutex) error {
	tempFileName := filepath.Join(tempDir, fmt.Sprintf("%s.part%d", filepath.Base(outputPath), ch.id))
	expectedSize := ch.end - ch.start + 1

	reconcile := func() int64 {
		currentSize := int64(0)
		if fi, err := os.Stat(tempFileName); err == nil {
			currentSize = fi.Size()
		}
		if currentSize > expectedSize {
			os.Remove(tempFileName)
			currentSize = 0
		}
		if delta := currentSize - ch.downloaded; delta != 0 {
			addProgress(progress, delta)
			ch.downloaded = currentSize
		}
		return currentSize
	}

	resumeOffset := reconcile()
	if resumeOffset == expectedSize {
		mu.Lock()
		ch.path = tempFileName
		ch.completed = true
		mu.Unlock()
		return nil
	}

	var lastErr error
	for retry := range maxRetries {
		if retry > 0 {
			timer := time.NewTimer(time.Duration(retry+1) * 500 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
			resumeOffset = reconcile()
		}
		if err := writeChunk(ctx, url, ch, client, tempFileName, progress, resumeOffset); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			lastErr = err
			resumeOffset = reconcile()
			continue
		}
		mu.Lock()
		ch.path = tempFileName
		ch.completed = true
		mu.Unlock()
		return nil
	}
	if lastErr != nil {
		return fmt.Errorf("chunk %d failed after %d retries: %w", ch.id, maxRetries, lastErr)
	}
	return fmt.Errorf("chunk %d failed after %d retries", ch.id, maxRetries)
}

func writeChunk(ctx context.Context, url string, ch *chunk, client *Client, tempFileName string, progress io.Writer, resumeOffset int64) error {
	flag := os.O_WRONLY | os.O_CREATE
	if resumeOffset > 0 {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}
	tempFile, err := os.OpenFile(tempFileName, flag, 0o644)
	if err != nil {
		return fmt.Errorf("error opening temp file: %w", err)
	}
	defer tempFile.Close()

	startByte := ch.start + resumeOffset
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", startByte, ch.end))
	req.Header.Set("Connection", "keep-alive")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	if resp.Header.Get("Content-Range") == "" {
		return fmt.Errorf("missing Content-Range header")
	}

	remaining := ch.end - startByte + 1
	var newBytes int64
	counter := countingWriter{w: progress, n: &newBytes}
	written, err := io.Copy(tempFile, io.TeeReader(io.LimitReader(resp.Body, remaining), counter))
	if err != nil {
		ch.downloaded += newBytes
		return err
	}
	ch.downloaded += newBytes
	if written != remaining {
		return fmt.Errorf("size mismatch: expected %d remaining bytes, got %d bytes this session", remaining, written)
	}
	if ch.downloaded != ch.end-ch.start+1 {
		return fmt.Errorf("total size mismatch: expected %d total bytes, got %d bytes", ch.end-ch.start+1, ch.downloaded)
	}
	return nil
}

func assemble(outputPath string, fileSize int64, chunks []chunk) error {
	for _, c := range chunks {
		if !c.completed {
			return fmt.Errorf("not all chunks were completed successfully")
		}
	}
	ordered := slices.Clone(chunks)
	slices.SortFunc(ordered, func(a, b chunk) int { return cmp.Compare(a.id, b.id) })

	destFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	var totalWritten int64
	for _, c := range ordered {
		tempFile, err := os.Open(c.path)
		if err != nil {
			return fmt.Errorf("error opening chunk file %s: %w", c.path, err)
		}
		fileInfo, err := tempFile.Stat()
		if err != nil {
			tempFile.Close()
			return fmt.Errorf("error getting chunk file info: %w", err)
		}
		written, err := io.Copy(destFile, tempFile)
		tempFile.Close()
		if err != nil {
			return fmt.Errorf("error copying chunk data: %w", err)
		}
		if written != fileInfo.Size() {
			return fmt.Errorf("error: wrote %d bytes but chunk size is %d", written, fileInfo.Size())
		}
		totalWritten += written
	}
	if totalWritten != fileSize {
		return fmt.Errorf("error: total written bytes (%d) doesn't match expected file size (%d)", totalWritten, fileSize)
	}
	for _, c := range ordered {
		os.Remove(c.path)
	}
	removeTempDirIfEmpty(filepath.Dir(ordered[0].path))
	return nil
}
