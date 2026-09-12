package anbuGenerics

import (
	"fmt"
	"os"
	"path/filepath"

	u "github.com/tanq16/anbu/utils"
)

const sizeThreshold = 100 * 1024 * 1024

type DuplicateSet struct {
	Files []string
}

type DeleteFailure struct {
	Path string
	Err  error
}

type DuplicateResult struct {
	Hashed     []DuplicateSet
	Unhashed   []DuplicateSet
	Deleted    []string
	DeleteErrs []DeleteFailure
}

func FindDuplicates(recursive bool, delete bool) (DuplicateResult, error) {
	var result DuplicateResult
	currentDir, err := os.Getwd()
	if err != nil {
		return result, fmt.Errorf("failed to get current directory: %w", err)
	}
	var files []string
	if recursive {
		err = filepath.WalkDir(currentDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return result, fmt.Errorf("failed to walk directory: %w", err)
		}
	} else {
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return result, fmt.Errorf("failed to read directory: %w", err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, filepath.Join(currentDir, entry.Name()))
			}
		}
	}

	sizeMap := make(map[int64][]string)
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		sizeMap[info.Size()] = append(sizeMap[info.Size()], file)
	}

	for size, fileList := range sizeMap {
		if len(fileList) < 2 {
			continue
		}
		if size >= sizeThreshold {
			result.Unhashed = append(result.Unhashed, DuplicateSet{Files: relPaths(currentDir, fileList)})
			continue
		}
		hashMap := make(map[string][]string)
		for _, file := range fileList {
			hash, err := u.ComputeFileHash(file)
			if err != nil {
				continue
			}
			hashMap[hash] = append(hashMap[hash], file)
		}
		for _, hashed := range hashMap {
			if len(hashed) >= 2 {
				result.Hashed = append(result.Hashed, DuplicateSet{Files: relPaths(currentDir, hashed)})
			}
		}
	}

	if !delete {
		return result, nil
	}
	result.Deleted, result.DeleteErrs = deleteExtras(currentDir, result.Hashed)
	unhashedDeleted, unhashedErrs := deleteExtras(currentDir, result.Unhashed)
	result.Deleted = append(result.Deleted, unhashedDeleted...)
	result.DeleteErrs = append(result.DeleteErrs, unhashedErrs...)
	return result, nil
}

func relPaths(base string, files []string) []string {
	out := make([]string, len(files))
	for i, file := range files {
		relPath, err := filepath.Rel(base, file)
		if err != nil {
			out[i] = file
			continue
		}
		out[i] = relPath
	}
	return out
}

func deleteExtras(base string, sets []DuplicateSet) ([]string, []DeleteFailure) {
	var deleted []string
	var failures []DeleteFailure
	for _, set := range sets {
		for i := 1; i < len(set.Files); i++ {
			path := set.Files[i]
			abs := path
			if !filepath.IsAbs(path) {
				abs = filepath.Join(base, path)
			}
			if err := os.Remove(abs); err != nil {
				failures = append(failures, DeleteFailure{Path: path, Err: err})
				continue
			}
			deleted = append(deleted, path)
		}
	}
	return deleted, failures
}
