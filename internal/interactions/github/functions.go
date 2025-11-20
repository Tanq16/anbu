package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v79/github"
	u "github.com/tanq16/anbu/utils"
)

func getClient(httpClient *http.Client) *github.Client {
	return github.NewClient(httpClient)
}

func ListIssues(client *http.Client, owner, repo string) error {
	ghClient := getClient(client)
	ctx := context.Background()
	issues, _, err := ghClient.Issues.ListByRepo(ctx, owner, repo, &github.IssueListByRepoOptions{
		State: "all",
	})
	if err != nil {
		return err
	}
	if len(issues) == 0 {
		u.PrintInfo("No issues found")
		return nil
	}

	table := u.NewTable([]string{"#", "Creator", "Tags", "Title"})
	for _, issue := range issues {
		var labels []string
		for _, label := range issue.Labels {
			if label.Name != nil {
				labels = append(labels, *label.Name)
			}
		}
		tags := strings.Join(labels, ", ")
		if tags == "" {
			tags = "--"
		}
		creator := "--"
		if issue.User != nil && issue.User.Login != nil {
			creator = *issue.User.Login
		}
		title := "--"
		if issue.Title != nil {
			title = *issue.Title
		}
		num := 0
		if issue.Number != nil {
			num = *issue.Number
		}
		table.Rows = append(table.Rows, []string{
			fmt.Sprintf("%d", num),
			creator,
			tags,
			title,
		})
	}
	table.PrintTable(false)
	return nil
}

func ListIssueComments(client *http.Client, owner, repo string, issueNum int) error {
	ghClient := getClient(client)
	ctx := context.Background()
	issue, _, err := ghClient.Issues.Get(ctx, owner, repo, issueNum)
	if err != nil {
		return err
	}
	author := "--"
	body := "--"
	if issue.User != nil && issue.User.Login != nil {
		author = *issue.User.Login
	}
	if issue.Body != nil {
		body = *issue.Body
	}
	if body != "" && body != "--" {
		fmt.Printf("%s: %s\n", author, body)
	}
	comments, _, err := ghClient.Issues.ListComments(ctx, owner, repo, issueNum, nil)
	if err != nil {
		return err
	}

	if len(comments) > 0 {
		if body != "" && body != "--" {
			fmt.Println(strings.Repeat(u.StyleSymbols["hline"], 80))
		}
		for i, comment := range comments {
			if i > 0 {
				fmt.Println(strings.Repeat(u.StyleSymbols["hline"], 80))
			}
			commentAuthor := "--"
			commentBody := "--"
			if comment.User != nil && comment.User.Login != nil {
				commentAuthor = *comment.User.Login
			}
			if comment.Body != nil {
				commentBody = *comment.Body
			}
			fmt.Printf("%s: %s\n", commentAuthor, commentBody)
		}
	}
	return nil
}

func ListPRs(client *http.Client, owner, repo string) error {
	ghClient := getClient(client)
	ctx := context.Background()
	prs, _, err := ghClient.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
		State: "all",
	})
	if err != nil {
		return err
	}
	if len(prs) == 0 {
		u.PrintInfo("No pull requests found")
		return nil
	}

	table := u.NewTable([]string{"#", "Creator", "Base", "Head", "Title"})
	for _, pr := range prs {
		creator := "--"
		baseBranch := "--"
		headBranch := "--"
		title := "--"
		if pr.User != nil && pr.User.Login != nil {
			creator = *pr.User.Login
		}
		if pr.Base != nil && pr.Base.Ref != nil {
			baseBranch = *pr.Base.Ref
		}
		if pr.Head != nil {
			if pr.Head.Ref != nil {
				headBranch = *pr.Head.Ref
			} else if pr.Head.Repo != nil && pr.Head.Repo.FullName != nil {
				headBranch = *pr.Head.Repo.FullName
			}
		}
		if pr.Title != nil {
			title = *pr.Title
		}
		num := 0
		if pr.Number != nil {
			num = *pr.Number
		}
		table.Rows = append(table.Rows, []string{
			fmt.Sprintf("%d", num),
			creator,
			baseBranch,
			headBranch,
			title,
		})
	}
	table.PrintTable(false)
	return nil
}

func ListPRComments(client *http.Client, owner, repo string, prNum int) error {
	ghClient := getClient(client)
	ctx := context.Background()
	pr, _, err := ghClient.PullRequests.Get(ctx, owner, repo, prNum)
	if err != nil {
		return err
	}
	author := "--"
	body := "--"
	if pr.User != nil && pr.User.Login != nil {
		author = *pr.User.Login
	}
	if pr.Body != nil {
		body = *pr.Body
	}
	if body != "" && body != "--" {
		fmt.Printf("%s: %s\n", author, body)
	}
	comments, _, err := ghClient.Issues.ListComments(ctx, owner, repo, prNum, nil)
	if err != nil {
		return err
	}

	if len(comments) > 0 {
		if body != "" && body != "--" {
			fmt.Println(strings.Repeat(u.StyleSymbols["hline"], 80))
		}
		for i, comment := range comments {
			if i > 0 {
				fmt.Println(strings.Repeat(u.StyleSymbols["hline"], 80))
			}
			commentAuthor := "--"
			commentBody := "--"
			if comment.User != nil && comment.User.Login != nil {
				commentAuthor = *comment.User.Login
			}
			if comment.Body != nil {
				commentBody = *comment.Body
			}
			fmt.Printf("%s: %s\n", commentAuthor, commentBody)
		}
	}
	return nil
}

func ListActions(client *http.Client, owner, repo string) error {
	ghClient := getClient(client)
	ctx := context.Background()
	runs, _, err := ghClient.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return err
	}
	if len(runs.WorkflowRuns) == 0 {
		u.PrintInfo("No workflow runs found")
		return nil
	}
	sort.Slice(runs.WorkflowRuns, func(i, j int) bool {
		timeI := time.Time{}
		timeJ := time.Time{}
		if runs.WorkflowRuns[i].CreatedAt != nil {
			timeI = runs.WorkflowRuns[i].CreatedAt.Time
		}
		if runs.WorkflowRuns[j].CreatedAt != nil {
			timeJ = runs.WorkflowRuns[j].CreatedAt.Time
		}
		return timeI.After(timeJ)
	})

	table := u.NewTable([]string{"#", "Time", "Workflow", "Status"})
	for i, run := range runs.WorkflowRuns {
		workflowName := "--"
		if run.Name != nil {
			workflowName = *run.Name
		}
		status := "--"
		if run.Status != nil {
			status = *run.Status
		}
		timeStr := "--"
		if run.CreatedAt != nil {
			timeStr = run.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		table.Rows = append(table.Rows, []string{
			fmt.Sprintf("%d", i+1),
			timeStr,
			workflowName,
			status,
		})
	}
	table.PrintTable(false)
	return nil
}

func ListActionJobs(client *http.Client, owner, repo string, runID int) error {
	ghClient := getClient(client)
	ctx := context.Background()
	runs, _, err := ghClient.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return err
	}
	if len(runs.WorkflowRuns) == 0 {
		return fmt.Errorf("no workflow runs found")
	}
	sort.Slice(runs.WorkflowRuns, func(i, j int) bool {
		timeI := time.Time{}
		timeJ := time.Time{}
		if runs.WorkflowRuns[i].CreatedAt != nil {
			timeI = runs.WorkflowRuns[i].CreatedAt.Time
		}
		if runs.WorkflowRuns[j].CreatedAt != nil {
			timeJ = runs.WorkflowRuns[j].CreatedAt.Time
		}
		return timeI.After(timeJ)
	})
	if runID < 1 || runID > len(runs.WorkflowRuns) {
		return fmt.Errorf("run ID %d out of range (1-%d)", runID, len(runs.WorkflowRuns))
	}
	run := runs.WorkflowRuns[runID-1]
	if run.ID == nil {
		return fmt.Errorf("run ID is nil")
	}
	jobs, _, err := ghClient.Actions.ListWorkflowJobs(ctx, owner, repo, *run.ID, &github.ListWorkflowJobsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return err
	}
	if len(jobs.Jobs) == 0 {
		u.PrintInfo("No jobs found for this workflow run")
		return nil
	}

	table := u.NewTable([]string{"#", "Name", "Status", "Conclusion"})
	for i, job := range jobs.Jobs {
		name := "--"
		status := "--"
		conclusion := "--"
		if job.Name != nil {
			name = *job.Name
		}
		if job.Status != nil {
			status = *job.Status
		}
		if job.Conclusion != nil {
			conclusion = *job.Conclusion
		}
		table.Rows = append(table.Rows, []string{
			fmt.Sprintf("%d", i+1),
			name,
			status,
			conclusion,
		})
	}
	table.PrintTable(false)
	return nil
}

func GetActionJobInfo(client *http.Client, owner, repo string, runID, jobID int) error {
	ghClient := getClient(client)
	ctx := context.Background()
	runs, _, err := ghClient.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return err
	}
	if len(runs.WorkflowRuns) == 0 {
		return fmt.Errorf("no workflow runs found")
	}
	sort.Slice(runs.WorkflowRuns, func(i, j int) bool {
		timeI := time.Time{}
		timeJ := time.Time{}
		if runs.WorkflowRuns[i].CreatedAt != nil {
			timeI = runs.WorkflowRuns[i].CreatedAt.Time
		}
		if runs.WorkflowRuns[j].CreatedAt != nil {
			timeJ = runs.WorkflowRuns[j].CreatedAt.Time
		}
		return timeI.After(timeJ)
	})

	if runID < 1 || runID > len(runs.WorkflowRuns) {
		return fmt.Errorf("run ID %d out of range (1-%d)", runID, len(runs.WorkflowRuns))
	}
	run := runs.WorkflowRuns[runID-1]
	if run.ID == nil {
		return fmt.Errorf("run ID is nil")
	}
	jobs, _, err := ghClient.Actions.ListWorkflowJobs(ctx, owner, repo, *run.ID, &github.ListWorkflowJobsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return err
	}
	if jobID < 1 || jobID > len(jobs.Jobs) {
		return fmt.Errorf("job ID %d out of range (1-%d)", jobID, len(jobs.Jobs))
	}
	job := jobs.Jobs[jobID-1]
	fmt.Printf("Job Name: %s\n", getString(job.Name))
	fmt.Printf("Status: %s\n", getString(job.Status))
	fmt.Printf("Conclusion: %s\n", getString(job.Conclusion))
	if job.StartedAt != nil {
		fmt.Printf("Started: %s\n", job.StartedAt.Time.Format("2006-01-02 15:04:05"))
	}
	if job.CompletedAt != nil {
		fmt.Printf("Completed: %s\n", job.CompletedAt.Time.Format("2006-01-02 15:04:05"))
	}
	if job.RunnerName != nil {
		fmt.Printf("Runner: %s\n", *job.RunnerName)
	}
	if job.Steps != nil {
		fmt.Printf("\nSteps:\n")
		for i, step := range job.Steps {
			fmt.Printf("  %d. %s - %s", i+1, getString(step.Name), getString(step.Status))
			if step.Conclusion != nil {
				fmt.Printf(" (%s)", *step.Conclusion)
			}
			fmt.Println()
		}
	}
	return nil
}

func StreamActionJobLogs(client *http.Client, owner, repo string, runID, jobID int) error {
	ghClient := getClient(client)
	ctx := context.Background()
	runs, _, err := ghClient.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return err
	}
	if len(runs.WorkflowRuns) == 0 {
		return fmt.Errorf("no workflow runs found")
	}

	sort.Slice(runs.WorkflowRuns, func(i, j int) bool {
		timeI := time.Time{}
		timeJ := time.Time{}
		if runs.WorkflowRuns[i].CreatedAt != nil {
			timeI = runs.WorkflowRuns[i].CreatedAt.Time
		}
		if runs.WorkflowRuns[j].CreatedAt != nil {
			timeJ = runs.WorkflowRuns[j].CreatedAt.Time
		}
		return timeI.After(timeJ)
	})

	if runID < 1 || runID > len(runs.WorkflowRuns) {
		return fmt.Errorf("run ID %d out of range (1-%d)", runID, len(runs.WorkflowRuns))
	}
	run := runs.WorkflowRuns[runID-1]
	if run.ID == nil {
		return fmt.Errorf("run ID is nil")
	}
	jobs, _, err := ghClient.Actions.ListWorkflowJobs(ctx, owner, repo, *run.ID, &github.ListWorkflowJobsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return err
	}
	if jobID < 1 || jobID > len(jobs.Jobs) {
		return fmt.Errorf("job ID %d out of range (1-%d)", jobID, len(jobs.Jobs))
	}

	job := jobs.Jobs[jobID-1]
	if job.ID == nil {
		return fmt.Errorf("job ID is nil")
	}
	logsURL, _, err := ghClient.Actions.GetWorkflowJobLogs(ctx, owner, repo, *job.ID, 3)
	if err != nil {
		return err
	}
	if logsURL == nil {
		return fmt.Errorf("logs URL is nil")
	}
	req, err := http.NewRequest("GET", logsURL.String(), nil)
	if err != nil {
		return err
	}

	httpClient := &http.Client{}
	httpResp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get logs: status %d", httpResp.StatusCode)
	}
	outFile, err := os.Create("anbu-github.log")
	if err != nil {
		return fmt.Errorf("failed to create log file: %v", err)
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to write logs: %v", err)
	}
	u.PrintSuccess("Logs successfully written to anbu-github.log")
	return nil
}

func AddIssueComment(client *http.Client, owner, repo string, issueNum int, body string) error {
	ghClient := getClient(client)
	ctx := context.Background()
	comment := &github.IssueComment{
		Body: github.Ptr(body),
	}
	_, _, err := ghClient.Issues.CreateComment(ctx, owner, repo, issueNum, comment)
	return err
}

func AddPRComment(client *http.Client, owner, repo string, prNum int, body string) error {
	return AddIssueComment(client, owner, repo, prNum, body)
}

func CreateIssue(client *http.Client, owner, repo string, title, body string) error {
	ghClient := getClient(client)
	ctx := context.Background()
	issue := &github.IssueRequest{
		Title: github.Ptr(title),
		Body:  github.Ptr(body),
	}
	_, _, err := ghClient.Issues.Create(ctx, owner, repo, issue)
	return err
}

func CreatePR(client *http.Client, owner, repo string, head, base string) error {
	ghClient := getClient(client)
	ctx := context.Background()
	pr := &github.NewPullRequest{
		Title: github.Ptr(fmt.Sprintf("PR: %s -> %s", head, base)),
		Head:  github.Ptr(head),
		Base:  github.Ptr(base),
	}
	_, _, err := ghClient.PullRequests.Create(ctx, owner, repo, pr)
	return err
}

func getString(s *string) string {
	if s == nil {
		return "--"
	}
	return *s
}
