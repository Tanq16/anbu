package archive

import (
	"archive/zip"
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type zipEntry struct {
	diskPath string
	zipName  string
	info     os.FileInfo
}

func Create(cfg CreateConfig) error {
	entries, err := collect(cfg)
	if err != nil {
		return err
	}
	if dir := filepath.Dir(cfg.Output); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	if !cfg.Encrypt {
		f, err := os.OpenFile(cfg.Output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		err = writeZip(f, entries)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		return closeErr
	}
	var buf bytes.Buffer
	if err := writeZip(&buf, entries); err != nil {
		return err
	}
	encrypted, err := encryptArchive(buf.Bytes(), cfg.Password)
	if err != nil {
		return err
	}
	return os.WriteFile(cfg.Output, encrypted, 0600)
}

func collect(cfg CreateConfig) ([]zipEntry, error) {
	outAbs, err := filepath.Abs(cfg.Output)
	if err != nil {
		return nil, err
	}
	wrapper := ""
	if !cfg.Bare {
		wrapper = baseName(cfg.Output)
	}
	var entries []zipEntry
	seen := map[string]struct{}{}
	for _, arg := range cfg.Paths {
		more, err := collectArg(arg, wrapper, outAbs, cfg.Include, cfg.Exclude)
		if err != nil {
			return nil, err
		}
		for _, e := range more {
			if _, ok := seen[e.zipName]; ok {
				continue
			}
			seen[e.zipName] = struct{}{}
			entries = append(entries, e)
		}
	}
	return entries, nil
}

func collectArg(arg, wrapper, outAbs string, include, exclude []*regexp.Regexp) ([]zipEntry, error) {
	info, err := os.Lstat(arg)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, nil
	}
	absArg, err := filepath.Abs(arg)
	if err != nil {
		return nil, err
	}
	if absArg == outAbs {
		return nil, nil
	}
	prefix := argumentPrefix(arg)
	var entries []zipEntry
	err = filepath.WalkDir(arg, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if absPath == outAbs {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(absArg, absPath)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		if relSlash == ".." || strings.HasPrefix(relSlash, "../") {
			return nil
		}
		name := prefix
		if rel != "." {
			if name == "" {
				name = relSlash
			} else {
				name = name + "/" + relSlash
			}
		}
		if wrapper != "" {
			if name == "" {
				name = wrapper
			} else {
				name = wrapper + "/" + name
			}
		}
		if info.IsDir() && name != "" && !strings.HasSuffix(name, "/") {
			name += "/"
		}
		if name == "" {
			return nil
		}
		if !keep(name, include, exclude) {
			if info.IsDir() && matchAny(exclude, name) {
				return fs.SkipDir
			}
			return nil
		}
		entries = append(entries, zipEntry{diskPath: absPath, zipName: name, info: info})
		return nil
	})
	return entries, err
}

func writeZip(w io.Writer, entries []zipEntry) error {
	zw := zip.NewWriter(w)
	for _, e := range entries {
		if err := addEntry(zw, e); err != nil {
			zw.Close()
			return err
		}
	}
	return zw.Close()
}

func addEntry(zw *zip.Writer, e zipEntry) error {
	header, err := zip.FileInfoHeader(e.info)
	if err != nil {
		return err
	}
	header.Name = e.zipName
	if e.info.IsDir() {
		_, err := zw.CreateHeader(header)
		return err
	}
	header.Method = zip.Deflate
	hw, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	f, err := os.Open(e.diskPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(hw, f)
	return err
}

func keep(name string, include, exclude []*regexp.Regexp) bool {
	if len(include) > 0 && !matchAny(include, name) {
		return false
	}
	return !matchAny(exclude, name)
}

func matchAny(res []*regexp.Regexp, name string) bool {
	return slices.ContainsFunc(res, func(re *regexp.Regexp) bool {
		return re.MatchString(name)
	})
}

func argumentPrefix(arg string) string {
	clean := filepath.Clean(arg)
	if clean == "." {
		return ""
	}
	slash := filepath.ToSlash(clean)
	if filepath.IsAbs(clean) || slash == ".." || strings.HasPrefix(slash, "../") {
		return filepath.ToSlash(filepath.Base(clean))
	}
	return slash
}

func baseName(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, ".enc")
	name := strings.TrimSuffix(base, filepath.Ext(base))
	if name == "" || name == "." {
		return "archive"
	}
	return name
}
