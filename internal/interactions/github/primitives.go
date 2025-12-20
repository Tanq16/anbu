package github

const (
	githubTokenFile = ".anbu-github-token.json"
)

type RepoPath struct {
	Owner string
	Repo  string
	Path  string
}
