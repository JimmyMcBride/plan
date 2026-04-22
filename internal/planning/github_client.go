package planning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitHubClient interface {
	Preflight(projectDir string) (*GitHubRepoInfo, error)
	CurrentContext(projectDir string) (*GitHubContext, error)
	CreateIssue(projectDir, repo string, input GitHubIssueInput) (*GitHubIssue, error)
	UpdateIssue(projectDir, repo string, issueNumber int, input GitHubIssueInput) (*GitHubIssue, error)
	GetIssue(projectDir, repo string, issueNumber int) (*GitHubIssue, error)
	FindMilestone(projectDir, repo, title string) (*GitHubMilestone, error)
}

type GitHubRepoInfo struct {
	Repo          string
	RepoURL       string
	DefaultBranch string
}

type GitHubPullRequest struct {
	Number     int
	URL        string
	State      string
	HeadRef    string
	BaseRef    string
	IsDraft    bool
	IsMerged   bool
	MergedAt   string
	HeadSHA    string
	DefaultRef string
}

type GitHubContext struct {
	Repo          GitHubRepoInfo
	CurrentBranch string
	CurrentSHA    string
	PlanningPR    *GitHubPullRequest
}

type GitHubIssueInput struct {
	Title     string
	Body      string
	State     string
	Labels    []string
	Milestone *int
}

type GitHubIssue struct {
	Number    int
	URL       string
	Title     string
	Body      string
	State     string
	Labels    []string
	Milestone *GitHubMilestone
}

type GitHubMilestone struct {
	Number int
	Title  string
}

var newGitHubClient = func() GitHubClient {
	return &cliGitHubClient{}
}

func SetGitHubClientFactoryForTesting(factory func() GitHubClient) func() {
	previous := newGitHubClient
	newGitHubClient = factory
	return func() {
		newGitHubClient = previous
	}
}

type cliGitHubClient struct{}

func (c *cliGitHubClient) Preflight(projectDir string) (*GitHubRepoInfo, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return nil, fmt.Errorf("gh is required for GitHub mode; install GitHub CLI from https://cli.github.com/ and retry")
	}
	if _, err := c.run(projectDir, nil, "gh", "auth", "status"); err != nil {
		return nil, fmt.Errorf("gh auth status failed; run `gh auth login` before enabling GitHub mode")
	}

	type repoView struct {
		NameWithOwner    string `json:"nameWithOwner"`
		URL              string `json:"url"`
		HasIssues        bool   `json:"hasIssuesEnabled"`
		DefaultBranchRef struct {
			Name string `json:"name"`
		} `json:"defaultBranchRef"`
	}

	out, err := c.run(projectDir, nil, "gh", "repo", "view", "--json", "nameWithOwner,url,hasIssuesEnabled,defaultBranchRef")
	if err != nil {
		return nil, fmt.Errorf("gh repo view failed; make sure this project is a GitHub checkout with an accessible origin")
	}
	var payload repoView
	if err := json.Unmarshal(out, &payload); err != nil {
		return nil, fmt.Errorf("parse GitHub repo metadata: %w", err)
	}
	if strings.TrimSpace(payload.NameWithOwner) == "" || strings.TrimSpace(payload.URL) == "" {
		return nil, fmt.Errorf("current repo did not resolve to a GitHub repository")
	}
	if !payload.HasIssues {
		return nil, fmt.Errorf("GitHub Issues are disabled for %s; enable Issues before turning on GitHub story mode", payload.NameWithOwner)
	}
	if strings.TrimSpace(payload.DefaultBranchRef.Name) == "" {
		return nil, fmt.Errorf("could not determine the default branch for %s", payload.NameWithOwner)
	}
	return &GitHubRepoInfo{
		Repo:          payload.NameWithOwner,
		RepoURL:       payload.URL,
		DefaultBranch: payload.DefaultBranchRef.Name,
	}, nil
}

func (c *cliGitHubClient) CurrentContext(projectDir string) (*GitHubContext, error) {
	repo, err := c.Preflight(projectDir)
	if err != nil {
		return nil, err
	}
	branchRaw, err := c.run(projectDir, nil, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("resolve current git branch: %w", err)
	}
	shaRaw, err := c.run(projectDir, nil, "git", "rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("resolve current git commit: %w", err)
	}
	context := &GitHubContext{
		Repo:          *repo,
		CurrentBranch: strings.TrimSpace(string(branchRaw)),
		CurrentSHA:    strings.TrimSpace(string(shaRaw)),
	}
	if context.CurrentBranch == "" {
		return nil, fmt.Errorf("could not determine current git branch")
	}

	type prPayload struct {
		Number      int    `json:"number"`
		URL         string `json:"url"`
		State       string `json:"state"`
		IsDraft     bool   `json:"isDraft"`
		MergedAt    string `json:"mergedAt"`
		HeadRefName string `json:"headRefName"`
		BaseRefName string `json:"baseRefName"`
	}

	out, err := c.run(projectDir, nil, "gh", "pr", "list", "--head", context.CurrentBranch, "--json", "number,url,state,isDraft,mergedAt,headRefName,baseRefName", "--limit", "1")
	if err != nil {
		return nil, fmt.Errorf("inspect pull request context: %w", err)
	}
	var prs []prPayload
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, fmt.Errorf("parse pull request context: %w", err)
	}
	if len(prs) > 0 {
		context.PlanningPR = &GitHubPullRequest{
			Number:     prs[0].Number,
			URL:        prs[0].URL,
			State:      prs[0].State,
			HeadRef:    prs[0].HeadRefName,
			BaseRef:    prs[0].BaseRefName,
			IsDraft:    prs[0].IsDraft,
			IsMerged:   strings.TrimSpace(prs[0].MergedAt) != "",
			MergedAt:   prs[0].MergedAt,
			HeadSHA:    context.CurrentSHA,
			DefaultRef: repo.DefaultBranch,
		}
	}
	return context, nil
}

func (c *cliGitHubClient) CreateIssue(projectDir, repo string, input GitHubIssueInput) (*GitHubIssue, error) {
	return c.upsertIssue(projectDir, fmt.Sprintf("repos/%s/issues", repo), input)
}

func (c *cliGitHubClient) UpdateIssue(projectDir, repo string, issueNumber int, input GitHubIssueInput) (*GitHubIssue, error) {
	return c.upsertIssue(projectDir, fmt.Sprintf("repos/%s/issues/%d", repo, issueNumber), input)
}

func (c *cliGitHubClient) upsertIssue(projectDir, apiPath string, input GitHubIssueInput) (*GitHubIssue, error) {
	payload := map[string]any{
		"title": input.Title,
		"body":  input.Body,
	}
	if strings.TrimSpace(input.State) != "" {
		payload["state"] = input.State
	}
	if input.Labels != nil {
		payload["labels"] = input.Labels
	}
	if input.Milestone != nil {
		payload["milestone"] = *input.Milestone
	}
	method := "POST"
	if strings.Contains(apiPath, "/issues/") {
		method = "PATCH"
	}
	out, err := c.api(projectDir, method, apiPath, payload)
	if err != nil {
		return nil, err
	}
	return parseGitHubIssue(out)
}

func (c *cliGitHubClient) GetIssue(projectDir, repo string, issueNumber int) (*GitHubIssue, error) {
	out, err := c.api(projectDir, "GET", fmt.Sprintf("repos/%s/issues/%d", repo, issueNumber), nil)
	if err != nil {
		return nil, err
	}
	return parseGitHubIssue(out)
}

func (c *cliGitHubClient) FindMilestone(projectDir, repo, title string) (*GitHubMilestone, error) {
	type milestonePayload struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
	}
	out, err := c.api(projectDir, "GET", fmt.Sprintf("repos/%s/milestones?state=all&per_page=100", repo), nil)
	if err != nil {
		return nil, err
	}
	var milestones []milestonePayload
	if err := json.Unmarshal(out, &milestones); err != nil {
		return nil, fmt.Errorf("parse milestones: %w", err)
	}
	for _, milestone := range milestones {
		if strings.EqualFold(strings.TrimSpace(milestone.Title), strings.TrimSpace(title)) {
			return &GitHubMilestone{Number: milestone.Number, Title: milestone.Title}, nil
		}
	}
	return nil, nil
}

func (c *cliGitHubClient) api(projectDir, method, apiPath string, payload any) ([]byte, error) {
	args := []string{"api", "--method", method, apiPath}
	var stdin []byte
	if payload != nil {
		args = append(args, "--input", "-")
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		stdin = raw
	}
	out, err := c.run(projectDir, stdin, "gh", args...)
	if err != nil {
		return nil, fmt.Errorf("gh api %s %s: %w", method, apiPath, err)
	}
	return out, nil
}

func (c *cliGitHubClient) run(projectDir string, stdin []byte, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = filepath.Clean(projectDir)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(out))
		if message == "" {
			message = err.Error()
		}
		return nil, fmt.Errorf("%s", message)
	}
	return out, nil
}

func parseGitHubIssue(raw []byte) (*GitHubIssue, error) {
	type label struct {
		Name string `json:"name"`
	}
	type payload struct {
		Number    int     `json:"number"`
		URL       string  `json:"html_url"`
		Title     string  `json:"title"`
		Body      string  `json:"body"`
		State     string  `json:"state"`
		Labels    []label `json:"labels"`
		Milestone *struct {
			Number int    `json:"number"`
			Title  string `json:"title"`
		} `json:"milestone"`
	}
	var item payload
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, fmt.Errorf("parse issue payload: %w", err)
	}
	issue := &GitHubIssue{
		Number: item.Number,
		URL:    item.URL,
		Title:  item.Title,
		Body:   item.Body,
		State:  item.State,
	}
	for _, current := range item.Labels {
		if strings.TrimSpace(current.Name) == "" {
			continue
		}
		issue.Labels = append(issue.Labels, current.Name)
	}
	if item.Milestone != nil {
		issue.Milestone = &GitHubMilestone{
			Number: item.Milestone.Number,
			Title:  item.Milestone.Title,
		}
	}
	return issue, nil
}
