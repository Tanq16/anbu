package interactionsCmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tanq16/anbu/internal/interactions/github"
	u "github.com/tanq16/anbu/utils"
)

var githubFlags struct {
	credentialsFile string
	pat             string
}

var GitHubCmd = &cobra.Command{
	Use:     "github",
	Aliases: []string{"gh"},
	Short:   "Interact with GitHub repositories and resources with OAuth app or Personal Access Token authentication",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip credentials file check if PAT is provided
		if githubFlags.pat != "" {
			return nil
		}
		if githubFlags.credentialsFile == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get user home directory: %w", err)
			}
			githubFlags.credentialsFile = filepath.Join(homeDir, ".anbu", "github-credentials.json")
		}
		if _, err := os.Stat(githubFlags.credentialsFile); os.IsNotExist(err) {
			return fmt.Errorf("credentials file not found at %s. Please provide one using the --credentials flag or place it at the default location", githubFlags.credentialsFile)
		}
		return nil
	},
}

var githubListCmd = &cobra.Command{
	Use:     "list [owner/repo/PATHS]",
	Aliases: []string{"ls"},
	Short:   "List GitHub resources (issues, PRs, actions)",
	Long: `Lists issues, pull requests, workflow runs, and their comments. Supports nested paths for detailed views and log streaming.
Examples:
  anbu gh ls owner/repo/i          - list all issues
  anbu gh ls owner/repo/i/23       - list comments for issue 23
  anbu gh ls owner/repo/pr         - list all PRs
  anbu gh ls owner/repo/pr/24      - list comments for PR 24
  anbu gh ls owner/repo/a          - list all workflow runs
  anbu gh ls owner/repo/a/3        - list jobs in workflow run 3
  anbu gh ls owner/repo/a/3/4     - get info for job 4 in run 3
  anbu gh ls owner/repo/a/3/4/logs - stream logs for job 4 in run 3`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := ""
		if len(args) > 0 {
			path = args[0]
		}
		client, err := github.GetGitHubClient(githubFlags.credentialsFile, githubFlags.pat)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get GitHub client")
		}
		parts := strings.Split(path, "/")
		if len(parts) < 2 {
			log.Fatal().Msg("Invalid path format. Expected: owner/repo/PATHS")
		}
		owner := parts[0]
		repo := parts[1]
		resourcePath := strings.Join(parts[2:], "/")
		if resourcePath == "" {
			log.Fatal().Msg("No resource path specified")
		}
		if err := handleList(client, owner, repo, resourcePath); err != nil {
			log.Fatal().Err(err).Msg("Failed to list resource")
		}
	},
}

var githubAddCmd = &cobra.Command{
	Use:   "add [owner/repo/PATHS]",
	Short: "Add comments to issues or PRs",
	Long: `Adds comments to GitHub issues or pull requests. Supports multi-line input terminated with 'EOF'.
Examples:
  anbu gh add owner/repo/i/23  - add comment to issue 23
  anbu gh add owner/repo/pr/24 - add comment to PR 24`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		client, err := github.GetGitHubClient(githubFlags.credentialsFile, githubFlags.pat)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get GitHub client")
		}
		parts := strings.Split(path, "/")
		if len(parts) < 3 {
			log.Fatal().Msg("Invalid path format. Expected: owner/repo/i/NUMBER or owner/repo/pr/NUMBER")
		}
		owner := parts[0]
		repo := parts[1]
		resourceType := parts[2]
		if len(parts) < 4 {
			log.Fatal().Msg("Missing issue/PR number")
		}
		var num int
		if _, err := fmt.Sscanf(parts[3], "%d", &num); err != nil {
			log.Fatal().Err(err).Msg("Invalid issue/PR number")
		}
		fmt.Println("Enter your comment (type 'EOF' on a new line to finish):")
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			line := scanner.Text()
			if line == "EOF" {
				break
			}
			lines = append(lines, line)
		}
		body := strings.Join(lines, "\n")
		switch resourceType {
		case "i":
			if err := github.AddIssueComment(client, owner, repo, num, body); err != nil {
				log.Fatal().Err(err).Msg("Failed to add issue comment")
			}
		case "pr":
			if err := github.AddPRComment(client, owner, repo, num, body); err != nil {
				log.Fatal().Err(err).Msg("Failed to add PR comment")
			}
		default:
			log.Fatal().Msg("Invalid resource type. Use 'i' for issues or 'pr' for PRs")
		}
		u.PrintSuccess("Comment added successfully")
	},
}

var githubMakeCmd = &cobra.Command{
	Use:   "make [owner/repo/PATHS]",
	Short: "Create issues or PRs",
	Long: `Creates new GitHub issues or pull requests. For PRs, supports custom base branches (defaults to main).
Examples:
  anbu gh make owner/repo/i              - create a new issue
  anbu gh make owner/repo/pr/branch      - create PR from branch to main
  anbu gh make owner/repo/pr/branch/base - create PR from branch to base`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		client, err := github.GetGitHubClient(githubFlags.credentialsFile, githubFlags.pat)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get GitHub client")
		}
		parts := strings.Split(path, "/")
		if len(parts) < 3 {
			log.Fatal().Msg("Invalid path format")
		}
		owner := parts[0]
		repo := parts[1]
		resourceType := parts[2]
		if resourceType == "i" {
			fmt.Print("Enter issue title: ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			title := scanner.Text()
			fmt.Println("Enter issue body (type 'EOF' on a new line to finish):")
			var lines []string
			for scanner.Scan() {
				line := scanner.Text()
				if line == "EOF" {
					break
				}
				lines = append(lines, line)
			}
			body := strings.Join(lines, "\n")
			if err := github.CreateIssue(client, owner, repo, title, body); err != nil {
				log.Fatal().Err(err).Msg("Failed to create issue")
			}
			u.PrintSuccess("Issue created successfully")
		} else if resourceType == "pr" {
			if len(parts) < 4 {
				log.Fatal().Msg("Missing branch name")
			}
			head := parts[3]
			base := "main"
			if len(parts) >= 5 {
				base = parts[4]
			}
			if err := github.CreatePR(client, owner, repo, head, base); err != nil {
				log.Fatal().Err(err).Msg("Failed to create PR")
			}
			u.PrintSuccess("PR created successfully")
		} else {
			log.Fatal().Msg("Invalid resource type. Use 'i' for issues or 'pr' for PRs")
		}
	},
}

var githubDownloadCmd = &cobra.Command{
	Use:     "download [OWNER/REPO/tree/REF/PATH]",
	Aliases: []string{"dl"},
	Short:   "Download files or folders from GitHub",
	Long: `Downloads files or folders from GitHub repositories. Supports branch names, commit SHAs, and recursive folder downloads.
The URL format is: OWNER/REPO/tree/BRANCH|COMMIT/PATH

Examples:
  anbu gh download owner/repo/tree/main/src/file.go     - download a single file
  anbu gh download owner/repo/tree/main/src             - download a folder
  anbu gh download owner/repo/tree/abc123def/path/to/dir - download from specific commit`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		client, err := github.GetGitHubClient(githubFlags.credentialsFile, githubFlags.pat)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get GitHub client")
		}
		if err := github.DownloadFromURL(client, url); err != nil {
			log.Fatal().Err(err).Msg("Failed to download")
		}
		u.PrintSuccess("Download completed successfully")
	},
}

func handleList(client *http.Client, owner, repo, resourcePath string) error {
	parts := strings.Split(resourcePath, "/")
	if len(parts) == 0 {
		return fmt.Errorf("empty resource path")
	}
	resourceType := parts[0]
	switch resourceType {
	case "i":
		if len(parts) == 1 {
			return github.ListIssues(client, owner, repo)
		} else if len(parts) == 2 {
			var issueNum int
			if _, err := fmt.Sscanf(parts[1], "%d", &issueNum); err != nil {
				return fmt.Errorf("invalid issue number: %v", err)
			}
			return github.ListIssueComments(client, owner, repo, issueNum)
		}
	case "pr":
		if len(parts) == 1 {
			return github.ListPRs(client, owner, repo)
		} else if len(parts) == 2 {
			var prNum int
			if _, err := fmt.Sscanf(parts[1], "%d", &prNum); err != nil {
				return fmt.Errorf("invalid PR number: %v", err)
			}
			return github.ListPRComments(client, owner, repo, prNum)
		}
	case "a":
		if len(parts) == 1 {
			return github.ListActions(client, owner, repo)
		} else if len(parts) == 2 {
			var runID int
			if _, err := fmt.Sscanf(parts[1], "%d", &runID); err != nil {
				return fmt.Errorf("invalid run ID: %v", err)
			}
			return github.ListActionJobs(client, owner, repo, runID)
		} else if len(parts) == 3 {
			var runID, jobID int
			if _, err := fmt.Sscanf(parts[1], "%d", &runID); err != nil {
				return fmt.Errorf("invalid run ID: %v", err)
			}
			if _, err := fmt.Sscanf(parts[2], "%d", &jobID); err != nil {
				return fmt.Errorf("invalid job ID: %v", err)
			}
			return github.GetActionJobInfo(client, owner, repo, runID, jobID)
		} else if len(parts) == 4 && parts[3] == "logs" {
			var runID, jobID int
			if _, err := fmt.Sscanf(parts[1], "%d", &runID); err != nil {
				return fmt.Errorf("invalid run ID: %v", err)
			}
			if _, err := fmt.Sscanf(parts[2], "%d", &jobID); err != nil {
				return fmt.Errorf("invalid job ID: %v", err)
			}
			return github.StreamActionJobLogs(client, owner, repo, runID, jobID)
		}
	}
	return fmt.Errorf("invalid resource path: %s", resourcePath)
}

func init() {
	GitHubCmd.PersistentFlags().StringVarP(&githubFlags.credentialsFile, "credentials", "c", "", "Path to GitHub credentials.json file (default ~/.anbu/github-credentials.json)")
	GitHubCmd.PersistentFlags().StringVar(&githubFlags.pat, "pat", "", "GitHub Personal Access Token (classic or fine-grained)")

	GitHubCmd.AddCommand(githubListCmd)
	GitHubCmd.AddCommand(githubAddCmd)
	GitHubCmd.AddCommand(githubMakeCmd)
	GitHubCmd.AddCommand(githubDownloadCmd)
}
