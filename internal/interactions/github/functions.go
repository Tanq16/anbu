package github

import (
	"net/http"
)

func ListIssues(client *http.Client, owner, repo string) error {
	return nil
}

func ListIssueComments(client *http.Client, owner, repo string, issueNum int) error {
	return nil
}

func ListPRs(client *http.Client, owner, repo string) error {
	return nil
}

func ListPRComments(client *http.Client, owner, repo string, prNum int) error {
	return nil
}

func ListActions(client *http.Client, owner, repo string) error {
	return nil
}

func ListActionJobs(client *http.Client, owner, repo string, runID int) error {
	return nil
}

func GetActionJobInfo(client *http.Client, owner, repo string, runID, jobID int) error {
	return nil
}

func StreamActionJobLogs(client *http.Client, owner, repo string, runID, jobID int) error {
	return nil
}

func AddIssueComment(client *http.Client, owner, repo string, issueNum int, body string) error {
	return nil
}

func AddPRComment(client *http.Client, owner, repo string, prNum int, body string) error {
	return nil
}

func CreateIssue(client *http.Client, owner, repo string, title, body string) error {
	return nil
}

func CreatePR(client *http.Client, owner, repo string, head, base string) error {
	return nil
}
