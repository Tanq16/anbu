package github

const (
	githubTokenFile = ".anbu-github-token.json"
	apiBaseURL      = "https://api.github.com"
)

type RepoPath struct {
	Owner string
	Repo  string
	Path  string
}

func ParsePath(path string) (*RepoPath, error) {
	return nil, nil
}
