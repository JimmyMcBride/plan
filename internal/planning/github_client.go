package planning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type GitHubClient interface {
	Preflight(projectDir string) (*GitHubRepoInfo, error)
	CurrentContext(projectDir string) (*GitHubContext, error)
	CreateIssue(projectDir, repo string, input GitHubIssueInput) (*GitHubIssue, error)
	UpdateIssue(projectDir, repo string, issueNumber int, input GitHubIssueInput) (*GitHubIssue, error)
	GetIssue(projectDir, repo string, issueNumber int) (*GitHubIssue, error)
	ListIssuesByLabel(projectDir, repo string, labels []string) ([]GitHubIssue, error)
	EnsureLabel(projectDir, repo string, input GitHubLabelInput) error
	FindMilestone(projectDir, repo, title string) (*GitHubMilestone, error)
	CreateMilestone(projectDir, repo string, input GitHubMilestoneInput) (*GitHubMilestone, error)
	GetDiscussion(projectDir, repo string, number int) (*GitHubDiscussion, error)
	UpdateDiscussionBody(projectDir, repo string, number int, body string) (*GitHubDiscussion, error)
	AddSubIssue(projectDir, repo string, issueNumber, subIssueNumber int) error
	AddBlockedBy(projectDir, repo string, issueNumber, blockingIssueNumber int) error
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
	Title          string
	Body           string
	State          string
	Labels         []string
	Milestone      *int
	ClearMilestone bool
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

type GitHubLabelInput struct {
	Name        string
	Color       string
	Description string
}

type GitHubMilestone struct {
	Number int
	Title  string
}

type GitHubMilestoneInput struct {
	Title       string
	Description string
}

type GitHubDiscussion struct {
	Number   int
	URL      string
	Title    string
	Body     string
	Comments []GitHubDiscussionComment
}

type GitHubDiscussionComment struct {
	Body string
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

const gitHubIssueListLimit = 1000

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
	if input.Milestone != nil && input.ClearMilestone {
		return nil, fmt.Errorf("cannot set and clear a milestone in the same issue request")
	}
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
	if input.ClearMilestone {
		payload["milestone"] = nil
	} else if input.Milestone != nil {
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

func (c *cliGitHubClient) ListIssuesByLabel(projectDir, repo string, labels []string) ([]GitHubIssue, error) {
	args := []string{"issue", "list", "--repo", repo, "--state", "all", "--limit", strconv.Itoa(gitHubIssueListLimit), "--json", "number,url,title,body,state,labels,milestone"}
	for _, label := range labels {
		if strings.TrimSpace(label) == "" {
			continue
		}
		args = append(args, "--label", label)
	}
	out, err := c.run(projectDir, nil, "gh", args...)
	if err != nil {
		return nil, fmt.Errorf("gh issue list failed: %w", err)
	}
	issues, err := parseGitHubIssueList(out)
	if err != nil {
		return nil, err
	}
	if len(issues) >= gitHubIssueListLimit {
		return nil, fmt.Errorf("GitHub issue listing for labels %s reached the %d issue safety limit; rerun with a narrower Plan label or add pagination before relying on drift checks", strings.Join(labels, ","), gitHubIssueListLimit)
	}
	return issues, nil
}

func (c *cliGitHubClient) EnsureLabel(projectDir, repo string, input GitHubLabelInput) error {
	args := []string{"label", "create", input.Name, "--repo", repo, "--force"}
	if strings.TrimSpace(input.Color) != "" {
		args = append(args, "--color", strings.TrimSpace(input.Color))
	}
	if strings.TrimSpace(input.Description) != "" {
		args = append(args, "--description", strings.TrimSpace(input.Description))
	}
	if _, err := c.run(projectDir, nil, "gh", args...); err != nil {
		return fmt.Errorf("ensure GitHub label %q: %w", input.Name, err)
	}
	return nil
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

func (c *cliGitHubClient) CreateMilestone(projectDir, repo string, input GitHubMilestoneInput) (*GitHubMilestone, error) {
	payload := map[string]any{
		"title": input.Title,
	}
	if strings.TrimSpace(input.Description) != "" {
		payload["description"] = strings.TrimSpace(input.Description)
	}
	out, err := c.api(projectDir, "POST", fmt.Sprintf("repos/%s/milestones", repo), payload)
	if err != nil {
		return nil, err
	}
	var milestone struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
	}
	if err := json.Unmarshal(out, &milestone); err != nil {
		return nil, fmt.Errorf("parse created milestone: %w", err)
	}
	return &GitHubMilestone{Number: milestone.Number, Title: milestone.Title}, nil
}

func (c *cliGitHubClient) GetDiscussion(projectDir, repo string, number int) (*GitHubDiscussion, error) {
	query := `query($owner:String!, $name:String!, $number:Int!, $after:String) {
  repository(owner:$owner, name:$name) {
    discussion(number:$number) {
      number
      url
      title
      body
      comments(first:100, after:$after) {
        nodes {
          body
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }
}`
	owner, name, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}
	var (
		after    any
		item     *GitHubDiscussion
		comments []GitHubDiscussionComment
	)
	for {
		payload := map[string]any{
			"query": query,
			"variables": map[string]any{
				"owner":  owner,
				"name":   name,
				"number": number,
				"after":  after,
			},
		}
		var response struct {
			Data struct {
				Repository struct {
					Discussion *struct {
						Number   int    `json:"number"`
						URL      string `json:"url"`
						Title    string `json:"title"`
						Body     string `json:"body"`
						Comments struct {
							Nodes []struct {
								Body string `json:"body"`
							} `json:"nodes"`
							PageInfo struct {
								HasNextPage bool   `json:"hasNextPage"`
								EndCursor   string `json:"endCursor"`
							} `json:"pageInfo"`
						} `json:"comments"`
					} `json:"discussion"`
				} `json:"repository"`
			} `json:"data"`
		}
		if err := c.graphql(projectDir, payload, &response); err != nil {
			return nil, err
		}
		if response.Data.Repository.Discussion == nil {
			return nil, fmt.Errorf("discussion #%d not found in %s", number, repo)
		}
		current := response.Data.Repository.Discussion
		if item == nil {
			item = &GitHubDiscussion{
				Number: current.Number,
				URL:    current.URL,
				Title:  current.Title,
				Body:   current.Body,
			}
		}
		for _, comment := range current.Comments.Nodes {
			comments = append(comments, GitHubDiscussionComment{Body: comment.Body})
		}
		if !current.Comments.PageInfo.HasNextPage {
			break
		}
		after = current.Comments.PageInfo.EndCursor
	}
	item.Comments = comments
	return item, nil
}

func (c *cliGitHubClient) UpdateDiscussionBody(projectDir, repo string, number int, body string) (*GitHubDiscussion, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}
	query := `query($owner:String!, $name:String!, $number:Int!) {
  repository(owner:$owner, name:$name) {
    discussion(number:$number) {
      id
      number
      url
      title
      body
    }
  }
}`
	var lookup struct {
		Data struct {
			Repository struct {
				Discussion *struct {
					ID     string `json:"id"`
					Number int    `json:"number"`
					URL    string `json:"url"`
					Title  string `json:"title"`
					Body   string `json:"body"`
				} `json:"discussion"`
			} `json:"repository"`
		} `json:"data"`
	}
	payload := map[string]any{
		"query": query,
		"variables": map[string]any{
			"owner":  owner,
			"name":   name,
			"number": number,
		},
	}
	if err := c.graphql(projectDir, payload, &lookup); err != nil {
		return nil, err
	}
	if lookup.Data.Repository.Discussion == nil || strings.TrimSpace(lookup.Data.Repository.Discussion.ID) == "" {
		return nil, fmt.Errorf("discussion #%d not found in %s", number, repo)
	}
	mutation := map[string]any{
		"query": `mutation($discussionId:ID!, $body:String!) {
  updateDiscussion(input:{discussionId:$discussionId, body:$body}) {
    discussion {
      number
      url
      title
      body
    }
  }
}`,
		"variables": map[string]any{
			"discussionId": lookup.Data.Repository.Discussion.ID,
			"body":         body,
		},
	}
	var response struct {
		Data struct {
			UpdateDiscussion struct {
				Discussion struct {
					Number int    `json:"number"`
					URL    string `json:"url"`
					Title  string `json:"title"`
					Body   string `json:"body"`
				} `json:"discussion"`
			} `json:"updateDiscussion"`
		} `json:"data"`
	}
	if err := c.graphql(projectDir, mutation, &response); err != nil {
		return nil, err
	}
	return &GitHubDiscussion{
		Number: response.Data.UpdateDiscussion.Discussion.Number,
		URL:    response.Data.UpdateDiscussion.Discussion.URL,
		Title:  response.Data.UpdateDiscussion.Discussion.Title,
		Body:   response.Data.UpdateDiscussion.Discussion.Body,
	}, nil
}

func (c *cliGitHubClient) AddSubIssue(projectDir, repo string, issueNumber, subIssueNumber int) error {
	issueID, subIssueID, err := c.issueIDs(projectDir, repo, issueNumber, subIssueNumber)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"query": `mutation($issueId:ID!, $subIssueId:ID!) {
  addSubIssue(input:{issueId:$issueId, subIssueId:$subIssueId}) {
    issue { number }
    subIssue { number }
  }
}`,
		"variables": map[string]any{
			"issueId":    issueID,
			"subIssueId": subIssueID,
		},
	}
	return c.graphql(projectDir, payload, nil)
}

func (c *cliGitHubClient) AddBlockedBy(projectDir, repo string, issueNumber, blockingIssueNumber int) error {
	issueID, blockingID, err := c.issueIDs(projectDir, repo, issueNumber, blockingIssueNumber)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"query": `mutation($issueId:ID!, $blockingIssueId:ID!) {
  addBlockedBy(input:{issueId:$issueId, blockingIssueId:$blockingIssueId}) {
    issue { number }
    blockingIssue { number }
  }
}`,
		"variables": map[string]any{
			"issueId":         issueID,
			"blockingIssueId": blockingID,
		},
	}
	return c.graphql(projectDir, payload, nil)
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

func (c *cliGitHubClient) graphql(projectDir string, payload any, target any) error {
	out, err := c.api(projectDir, "POST", "graphql", payload)
	if err != nil {
		return err
	}
	if err := decodeGraphQLResponse(out, target); err != nil {
		return err
	}
	return nil
}

func (c *cliGitHubClient) issueIDs(projectDir, repo string, issueNumber, otherIssueNumber int) (string, string, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return "", "", err
	}
	payload := map[string]any{
		"query": `query($owner:String!, $name:String!, $issueNumber:Int!, $otherIssueNumber:Int!) {
  repository(owner:$owner, name:$name) {
    issue(number:$issueNumber) { id }
    otherIssue: issue(number:$otherIssueNumber) { id }
  }
}`,
		"variables": map[string]any{
			"owner":            owner,
			"name":             name,
			"issueNumber":      issueNumber,
			"otherIssueNumber": otherIssueNumber,
		},
	}
	var response struct {
		Data struct {
			Repository struct {
				Issue struct {
					ID string `json:"id"`
				} `json:"issue"`
				OtherIssue struct {
					ID string `json:"id"`
				} `json:"otherIssue"`
			} `json:"repository"`
		} `json:"data"`
	}
	if err := c.graphql(projectDir, payload, &response); err != nil {
		return "", "", err
	}
	if strings.TrimSpace(response.Data.Repository.Issue.ID) == "" || strings.TrimSpace(response.Data.Repository.OtherIssue.ID) == "" {
		return "", "", fmt.Errorf("could not resolve issue node ids for #%d and #%d", issueNumber, otherIssueNumber)
	}
	return response.Data.Repository.Issue.ID, response.Data.Repository.OtherIssue.ID, nil
}

func decodeGraphQLResponse(raw []byte, target any) error {
	var envelope struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return fmt.Errorf("parse graphql response envelope: %w", err)
	}
	if len(envelope.Errors) > 0 {
		messages := make([]string, 0, len(envelope.Errors))
		for _, item := range envelope.Errors {
			if strings.TrimSpace(item.Message) == "" {
				continue
			}
			messages = append(messages, strings.TrimSpace(item.Message))
		}
		if len(messages) == 0 {
			return fmt.Errorf("graphql request failed")
		}
		return fmt.Errorf("graphql request failed: %s", strings.Join(messages, "; "))
	}
	if target == nil {
		return nil
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("parse graphql response: %w", err)
	}
	return nil
}

func splitRepo(repo string) (string, string, error) {
	parts := strings.Split(strings.TrimSpace(repo), "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("invalid GitHub repo %q", repo)
	}
	return parts[0], parts[1], nil
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

func parseGitHubIssueList(raw []byte) ([]GitHubIssue, error) {
	type label struct {
		Name string `json:"name"`
	}
	type payload struct {
		Number    int     `json:"number"`
		URL       string  `json:"url"`
		Title     string  `json:"title"`
		Body      string  `json:"body"`
		State     string  `json:"state"`
		Labels    []label `json:"labels"`
		Milestone *struct {
			Number int    `json:"number"`
			Title  string `json:"title"`
		} `json:"milestone"`
	}
	var items []payload
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("parse issue list payload: %w", err)
	}
	out := make([]GitHubIssue, 0, len(items))
	for _, item := range items {
		issue := GitHubIssue{
			Number: item.Number,
			URL:    item.URL,
			Title:  item.Title,
			Body:   item.Body,
			State:  item.State,
		}
		for _, label := range item.Labels {
			if strings.TrimSpace(label.Name) != "" {
				issue.Labels = append(issue.Labels, label.Name)
			}
		}
		if item.Milestone != nil {
			issue.Milestone = &GitHubMilestone{
				Number: item.Milestone.Number,
				Title:  item.Milestone.Title,
			}
		}
		out = append(out, issue)
	}
	return out, nil
}
