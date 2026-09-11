package archive

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func IsEncrypted(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	magic := make([]byte, len(archiveMagic))
	_, err = io.ReadFull(f, magic)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return false, nil
		}
		return false, err
	}
	return bytes.Equal(magic, archiveMagic), nil
}

func Extract(cfg ExtractConfig) error {
	destDir, err := filepath.Abs(cfg.Dest)
	if err != nil {
		return err
	}
	destDir = filepath.Clean(destDir)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	encrypted, err := IsEncrypted(cfg.Archive)
	if err != nil {
		return err
	}
	if encrypted {
		if cfg.Password == "" {
			return fmt.Errorf("password required")
		}
		src, err := os.Open(cfg.Archive)
		if err != nil {
			return err
		}
		tmp, err := os.CreateTemp("", "anbu-archive-*.zip")
		if err != nil {
			src.Close()
			return err
		}
		tmpName := tmp.Name()
		err = decryptTo(tmp, src, cfg.Password)
		closeErr := errors.Join(src.Close(), tmp.Close())
		if err != nil {
			os.Remove(tmpName)
			return err
		}
		if closeErr != nil {
			os.Remove(tmpName)
			return closeErr
		}
		defer os.Remove(tmpName)
		r, err := zip.OpenReader(tmpName)
		if err != nil {
			return err
		}
		defer r.Close()
		return extractFiles(r.File, destDir, cfg.Bare)
	}
	r, err := zip.OpenReader(cfg.Archive)
	if err != nil {
		return err
	}
	defer r.Close()
	return extractFiles(r.File, destDir, cfg.Bare)
}

func extractFiles(files []*zip.File, destDir string, bare bool) error {
	for _, f := range files {
		if err := extractOne(f, destDir, bare); err != nil {
			return err
		}
	}
	return nil
}

func extractOne(f *zip.File, destDir string, bare bool) error {
	mode := f.Mode()
	if mode&os.ModeSymlink != 0 {
		return nil
	}
	name := f.Name
	if bare {
		stripped, ok := stripFirst(name)
		if !ok {
			return nil
		}
		name = stripped
	}
	target, err := safeExtractPath(destDir, name)
	if err != nil {
		return err
	}
	info := f.FileInfo()
	if info.IsDir() || strings.HasSuffix(name, "/") {
		return os.MkdirAll(target, 0755)
	}
	if !info.Mode().IsRegular() && !mode.IsRegular() {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	perm := info.Mode().Perm()
	if perm == 0 {
		perm = 0644
	}
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	rc, err := f.Open()
	if err != nil {
		out.Close()
		return err
	}
	_, copyErr := io.Copy(out, rc)
	closeErr := errors.Join(out.Close(), rc.Close())
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func stripFirst(name string) (string, bool) {
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.TrimPrefix(name, "/")
	slash := strings.IndexByte(name, '/')
	if slash < 0 {
		return "", false
	}
	rest := name[slash+1:]
	if rest == "" {
		return "", false
	}
	return rest, true
}

func safeExtractPath(destDir, name string) (string, error) {
	name = strings.ReplaceAll(name, "\\", "/")
	target := filepath.Join(destDir, filepath.FromSlash(name))
	target = filepath.Clean(target)
	sep := string(os.PathSeparator)
	if target != destDir && !strings.HasPrefix(target, destDir+sep) {
		return "", fmt.Errorf("illegal path in archive: %s", name)
	}
	return target, nil
}
