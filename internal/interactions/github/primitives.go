package github

import (
	"fmt"
	"strings"
)

const (
	githubTokenFile = ".anbu-github-token.json"
)

type RepoPath struct {
	Owner string
	Repo  string
	Path  string
}

func ParsePath(path string) (*RepoPath, error) {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid path format: expected owner/repo/path")
	}
	rp := &RepoPath{
		Owner: parts[0],
		Repo:  parts[1],
	}
	if len(parts) > 2 {
		rp.Path = strings.Join(parts[2:], "/")
	}
	return rp, nil
}
