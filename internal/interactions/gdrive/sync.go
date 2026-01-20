package gdrive

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	u "github.com/tanq16/anbu/utils"
	"google.golang.org/api/drive/v3"
)

type FileTree struct {
	Files map[string]FileInfo
	Dirs  map[string]*FileTree
}

type FileInfo struct {
	Path string
	Hash string
	ID   string
}

func computeLocalHash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func buildLocalTree(rootDir string, ignoreSet map[string]struct{}) (*FileTree, error) {
	tree := &FileTree{
		Files: make(map[string]FileInfo),
		Dirs:  make(map[string]*FileTree),
	}
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if _, skip := ignoreSet[d.Name()]; skip {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		if d.IsDir() {
			parts := strings.Split(relPath, string(filepath.Separator))
			current := tree
			for _, part := range parts {
				if current.Dirs == nil {
					current.Dirs = make(map[string]*FileTree)
				}
				if _, exists := current.Dirs[part]; !exists {
					current.Dirs[part] = &FileTree{
						Files: make(map[string]FileInfo),
						Dirs:  make(map[string]*FileTree),
					}
				}
				current = current.Dirs[part]
			}
		} else {
			hash, err := computeLocalHash(path)
			if err != nil {
				u.PrintError(fmt.Sprintf("Failed to compute hash for %s", path))
				log.Debug().Err(err).Msgf("Failed to compute hash for %s", path)
				return nil
			}
			parts := strings.Split(relPath, string(filepath.Separator))
			fileName := parts[len(parts)-1]
			dirPath := strings.Join(parts[:len(parts)-1], string(filepath.Separator))
			current := tree
			if dirPath != "" {
				for part := range strings.SplitSeq(dirPath, string(filepath.Separator)) {
					if current.Dirs == nil {
						current.Dirs = make(map[string]*FileTree)
					}
					if _, exists := current.Dirs[part]; !exists {
						current.Dirs[part] = &FileTree{
							Files: make(map[string]FileInfo),
							Dirs:  make(map[string]*FileTree),
						}
					}
					current = current.Dirs[part]
				}
			}
			current.Files[fileName] = FileInfo{
				Path: relPath,
				Hash: hash,
			}
		}
		return nil
	})
	return tree, err
}

func buildRemoteTree(srv *drive.Service, folderID string, basePath string, ignoreSet map[string]struct{}) (*FileTree, error) {
	tree := &FileTree{
		Files: make(map[string]FileInfo),
		Dirs:  make(map[string]*FileTree),
	}
	query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	var pageToken string
	for {
		r, err := srv.Files.List().Q(query).
			PageSize(100).
			Fields("nextPageToken, files(id, name, mimeType, md5Checksum)").
			PageToken(pageToken).Do()
		if err != nil {
			return nil, err
		}
		for _, f := range r.Files {
			if _, skip := ignoreSet[f.Name]; skip {
				continue
			}
			itemPath := filepath.Join(basePath, f.Name)
			if f.MimeType == googleFolderMimeType {
				subTree, err := buildRemoteTree(srv, f.Id, itemPath, ignoreSet)
				if err != nil {
					u.PrintError(fmt.Sprintf("Failed to build tree for folder %s", itemPath))
					log.Debug().Err(err).Msgf("Failed to build tree for folder %s", itemPath)
					continue
				}
				tree.Dirs[f.Name] = subTree
			} else {
				hash := f.Md5Checksum
				if hash == "" {
					u.PrintWarning(fmt.Sprintf("No MD5 hash for file %s", itemPath))
					log.Debug().Msgf("No MD5 hash for file %s", itemPath)
					continue
				}
				tree.Files[f.Name] = FileInfo{
					Path: itemPath,
					Hash: hash,
					ID:   f.Id,
				}
			}
		}
		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return tree, nil
}

func deleteDriveFile(srv *drive.Service, fileID string) error {
	return srv.Files.Delete(fileID).Do()
}

func uploadDriveFileToFolder(srv *drive.Service, localPath string, parentFolderID string) (*drive.File, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileName := filepath.Base(localPath)
	fileMetadata := &drive.File{
		Name:    fileName,
		Parents: []string{parentFolderID},
	}
	driveFile, err := srv.Files.Create(fileMetadata).Media(file).Fields("id, name").Do()
	if err != nil {
		return nil, err
	}
	return driveFile, nil
}

func updateDriveFile(srv *drive.Service, fileID string, localPath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = srv.Files.Update(fileID, nil).Media(file).Do()
	return err
}

func syncTree(srv *drive.Service, localTree *FileTree, remoteTree *FileTree, localBase string, remoteFolderID string, sem chan struct{}, wg *sync.WaitGroup) error {
	for fileName, localFile := range localTree.Files {
		remoteFile, exists := remoteTree.Files[fileName]
		localPath := filepath.Join(localBase, fileName)
		switch {
		case !exists:
			sem <- struct{}{}
			wg.Add(1)
			go func(localPath string, localFile FileInfo) {
				defer wg.Done()
				defer func() { <-sem }()
				fmt.Printf("Uploading %s\n", u.FDebug(localFile.Path))
				if _, err := uploadDriveFileToFolder(srv, localPath, remoteFolderID); err != nil {
					u.PrintError(fmt.Sprintf("Failed to upload %s", localFile.Path))
					log.Debug().Err(err).Msgf("Failed to upload %s", localFile.Path)
				}
			}(localPath, localFile)
		case localFile.Hash != remoteFile.Hash:
			sem <- struct{}{}
			wg.Add(1)
			go func(localPath string, localFile FileInfo, remoteFile FileInfo) {
				defer wg.Done()
				defer func() { <-sem }()
				fmt.Printf("Updating %s\n", u.FDebug(localFile.Path))
				if err := updateDriveFile(srv, remoteFile.ID, localPath); err != nil {
					u.PrintError(fmt.Sprintf("Failed to update %s", localFile.Path))
					log.Debug().Err(err).Msgf("Failed to update %s", localFile.Path)
				}
			}(localPath, localFile, remoteFile)
		}
	}
	for fileName, remoteFile := range remoteTree.Files {
		if _, exists := localTree.Files[fileName]; !exists {
			sem <- struct{}{}
			wg.Add(1)
			go func(remoteFile FileInfo) {
				defer wg.Done()
				defer func() { <-sem }()
				fmt.Printf("%s %s\n", u.FError("Deleting"), u.FDebug(remoteFile.Path))
				if err := deleteDriveFile(srv, remoteFile.ID); err != nil {
					u.PrintError(fmt.Sprintf("Failed to delete %s", remoteFile.Path))
					log.Debug().Err(err).Msgf("Failed to delete %s", remoteFile.Path)
				}
			}(remoteFile)
		}
	}
	for dirName, localSubTree := range localTree.Dirs {
		remoteSubTree, exists := remoteTree.Dirs[dirName]
		var subFolderID string
		if !exists {
			folderMetadata := &drive.File{
				Name:     dirName,
				MimeType: googleFolderMimeType,
				Parents:  []string{remoteFolderID},
			}
			f, err := srv.Files.Create(folderMetadata).Fields("id").Do()
			if err != nil {
				u.PrintError(fmt.Sprintf("Failed to create folder %s", dirName))
				log.Debug().Err(err).Msgf("Failed to create folder %s", dirName)
				continue
			}
			subFolderID = f.Id
			remoteSubTree = &FileTree{
				Files: make(map[string]FileInfo),
				Dirs:  make(map[string]*FileTree),
			}
		} else {
			query := fmt.Sprintf("name = '%s' and '%s' in parents and mimeType = '%s' and trashed = false", dirName, remoteFolderID, googleFolderMimeType)
			r, err := srv.Files.List().Q(query).Fields("files(id)").PageSize(1).Do()
			if err != nil || len(r.Files) == 0 {
				u.PrintError(fmt.Sprintf("Failed to find folder ID for %s", dirName))
				log.Debug().Msgf("Failed to find folder ID for %s", dirName)
				continue
			}
			subFolderID = r.Files[0].Id
		}
		subLocalBase := filepath.Join(localBase, dirName)
		if err := syncTree(srv, localSubTree, remoteSubTree, subLocalBase, subFolderID, sem, wg); err != nil {
			u.PrintError(fmt.Sprintf("Failed to sync subtree %s", dirName))
			log.Debug().Err(err).Msgf("Failed to sync subtree %s", dirName)
		}
	}
	for dirName := range remoteTree.Dirs {
		if _, exists := localTree.Dirs[dirName]; !exists {
			sem <- struct{}{}
			wg.Add(1)
			go func(dirName string) {
				defer wg.Done()
				defer func() { <-sem }()
				query := fmt.Sprintf("name = '%s' and '%s' in parents and mimeType = '%s' and trashed = false", dirName, remoteFolderID, googleFolderMimeType)
				r, err := srv.Files.List().Q(query).Fields("files(id)").PageSize(1).Do()
				if err == nil && len(r.Files) > 0 {
					if err := deleteDriveFolderRecursive(srv, r.Files[0].Id); err != nil {
						u.PrintError(fmt.Sprintf("Failed to delete folder %s", dirName))
						log.Debug().Err(err).Msgf("Failed to delete folder %s", dirName)
					} else {
						fmt.Printf("%s %s\n", u.FError("Deleting folder"), u.FDebug(dirName))
					}
				}
			}(dirName)
		}
	}
	return nil
}

func deleteDriveFolderRecursive(srv *drive.Service, folderID string) error {
	query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	var pageToken string
	for {
		r, err := srv.Files.List().Q(query).
			PageSize(100).
			Fields("nextPageToken, files(id, mimeType)").
			PageToken(pageToken).Do()
		if err != nil {
			return err
		}
		for _, f := range r.Files {
			if f.MimeType == googleFolderMimeType {
				if err := deleteDriveFolderRecursive(srv, f.Id); err != nil {
					return err
				}
			} else {
				if err := srv.Files.Delete(f.Id).Do(); err != nil {
					return err
				}
			}
		}
		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return srv.Files.Delete(folderID).Do()
}

func SyncDriveDirectory(srv *drive.Service, localDir string, remotePath string, concurrency int, ignore []string) error {
	ignoreSet := make(map[string]struct{})
	for _, v := range ignore {
		name := strings.TrimSpace(v)
		if name != "" {
			ignoreSet[name] = struct{}{}
		}
	}
	log.Debug().Str("localDir", localDir).Msg("building local tree")
	localTree, err := buildLocalTree(localDir, ignoreSet)
	if err != nil {
		return fmt.Errorf("failed to build local tree: %v", err)
	}
	folderID := "root"
	if remotePath != "" && remotePath != "/" && remotePath != "root" {
		item, err := getItemIdByPath(srv, remotePath)
		if err != nil {
			return fmt.Errorf("failed to resolve remote path: %v", err)
		}
		if item.MimeType != googleFolderMimeType {
			return fmt.Errorf("remote path is not a folder")
		}
		folderID = item.Id
	}
	log.Debug().Str("remotePath", remotePath).Str("folderID", folderID).Msg("building remote tree")
	remoteTree, err := buildRemoteTree(srv, folderID, "", ignoreSet)
	if err != nil {
		return fmt.Errorf("failed to build remote tree: %v", err)
	}
	if concurrency < 1 {
		concurrency = 1
	}
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	err = syncTree(srv, localTree, remoteTree, localDir, folderID, sem, &wg)
	wg.Wait()
	return err
}
