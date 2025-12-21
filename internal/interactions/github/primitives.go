package github

const (
	githubTokenFile = "github-token.json"
)

type RepoPath struct {
	Owner string
	Repo  string
	Path  string
}
