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
	CreateProjectWorkspace(projectDir, repo string, input GitHubProjectWorkspaceInput) (*GitHubProjectWorkspace, error)
	GetProjectWorkspace(projectDir, repo string, ref GitHubProjectReference) (*GitHubProjectWorkspace, error)
	EnsureProjectField(projectDir string, project GitHubProjectWorkspace, input GitHubProjectFieldInput) (*GitHubProjectField, error)
	AddProjectItemByIssue(projectDir, repo, projectID string, issueNumber int) (*GitHubProjectItem, error)
	SetProjectItemField(projectDir, projectID, itemID string, field GitHubProjectField, value string) error
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
	NodeID    string
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

type GitHubProjectReference struct {
	Owner  string
	Number int
	ID     string
	URL    string
}

type GitHubProjectWorkspaceInput struct {
	Owner string
	Title string
}

type GitHubProjectWorkspace struct {
	Owner  string
	Number int
	ID     string
	URL    string
	Title  string
	Fields []GitHubProjectField
}

type GitHubProjectFieldInput struct {
	Name     string
	DataType string
	Options  []string
}

type GitHubProjectField struct {
	ID       string
	Name     string
	DataType string
	Options  map[string]string
}

type GitHubProjectItem struct {
	ID          string
	IssueNumber int
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

type cliGitHubClient struct {
	issueNodeIDs map[string]string
}

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
	issue, err := parseGitHubIssue(out)
	if err != nil {
		return nil, err
	}
	c.cacheIssueNodeID(repoFromIssueAPIPath(apiPath), issue)
	return issue, nil
}

func (c *cliGitHubClient) GetIssue(projectDir, repo string, issueNumber int) (*GitHubIssue, error) {
	out, err := c.api(projectDir, "GET", fmt.Sprintf("repos/%s/issues/%d", repo, issueNumber), nil)
	if err != nil {
		return nil, err
	}
	issue, err := parseGitHubIssue(out)
	if err != nil {
		return nil, err
	}
	c.cacheIssueNodeID(repo, issue)
	return issue, nil
}

func (c *cliGitHubClient) ListIssuesByLabel(projectDir, repo string, labels []string) ([]GitHubIssue, error) {
	args := []string{"issue", "list", "--repo", repo, "--state", "all", "--limit", strconv.Itoa(gitHubIssueListLimit), "--json", "id,number,url,title,body,state,labels,milestone"}
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
	for i := range issues {
		c.cacheIssueNodeID(repo, &issues[i])
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

func (c *cliGitHubClient) CreateProjectWorkspace(projectDir, repo string, input GitHubProjectWorkspaceInput) (*GitHubProjectWorkspace, error) {
	owner := strings.TrimSpace(input.Owner)
	if owner == "" {
		repoOwner, _, err := splitRepo(repo)
		if err != nil {
			return nil, err
		}
		owner = repoOwner
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, fmt.Errorf("project title is required")
	}
	ownerID, repoID, err := c.projectOwnerAndRepositoryIDs(projectDir, repo, owner)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{
		"query": `mutation($ownerId:ID!, $repositoryId:ID!, $title:String!) {
  createProjectV2(input:{ownerId:$ownerId, repositoryId:$repositoryId, title:$title}) {
    projectV2 {
      id
      number
      url
      title
      fields(first:100) {
        nodes {
          __typename
          ... on ProjectV2Field {
            id
            name
            dataType
          }
          ... on ProjectV2SingleSelectField {
            id
            name
            dataType
            options {
              id
              name
            }
          }
        }
      }
    }
  }
}`,
		"variables": map[string]any{
			"ownerId":      ownerID,
			"repositoryId": repoID,
			"title":        title,
		},
	}
	var response struct {
		Data struct {
			CreateProjectV2 struct {
				ProjectV2 projectV2Payload `json:"projectV2"`
			} `json:"createProjectV2"`
		} `json:"data"`
	}
	if err := c.graphql(projectDir, payload, &response); err != nil {
		return nil, err
	}
	project := projectWorkspaceFromPayload(owner, response.Data.CreateProjectV2.ProjectV2)
	if strings.TrimSpace(project.ID) == "" {
		return nil, fmt.Errorf("created GitHub Project did not include an id")
	}
	return project, nil
}

func (c *cliGitHubClient) GetProjectWorkspace(projectDir, repo string, ref GitHubProjectReference) (*GitHubProjectWorkspace, error) {
	if strings.TrimSpace(ref.ID) != "" {
		return c.getProjectWorkspaceByID(projectDir, strings.TrimSpace(ref.ID), strings.TrimSpace(ref.Owner))
	}
	owner := strings.TrimSpace(ref.Owner)
	if owner == "" {
		return nil, fmt.Errorf("project owner is required to connect an existing GitHub Project")
	}
	if ref.Number <= 0 {
		return nil, fmt.Errorf("project number is required to connect an existing GitHub Project")
	}
	payload := map[string]any{
		"query": `query($owner:String!, $number:Int!) {
  repositoryOwner(login:$owner) {
    login
    ... on User {
      projectV2(number:$number) {
        id
        number
        url
        title
        fields(first:100) {
          nodes {
            __typename
            ... on ProjectV2Field {
              id
              name
              dataType
            }
            ... on ProjectV2SingleSelectField {
              id
              name
              dataType
              options {
                id
                name
              }
            }
          }
        }
      }
    }
    ... on Organization {
      projectV2(number:$number) {
        id
        number
        url
        title
        fields(first:100) {
          nodes {
            __typename
            ... on ProjectV2Field {
              id
              name
              dataType
            }
            ... on ProjectV2SingleSelectField {
              id
              name
              dataType
              options {
                id
                name
              }
            }
          }
        }
      }
    }
  }
}`,
		"variables": map[string]any{
			"owner":  owner,
			"number": ref.Number,
		},
	}
	var response struct {
		Data struct {
			RepositoryOwner *struct {
				Login   string           `json:"login"`
				Project projectV2Payload `json:"projectV2"`
			} `json:"repositoryOwner"`
		} `json:"data"`
	}
	if err := c.graphql(projectDir, payload, &response); err != nil {
		return nil, err
	}
	if response.Data.RepositoryOwner == nil {
		return nil, fmt.Errorf("GitHub Project owner %q not found", owner)
	}
	project := projectWorkspaceFromPayload(owner, response.Data.RepositoryOwner.Project)
	if strings.TrimSpace(project.ID) == "" {
		return nil, fmt.Errorf("GitHub Project %s/%d not found", owner, ref.Number)
	}
	return project, nil
}

func (c *cliGitHubClient) EnsureProjectField(projectDir string, project GitHubProjectWorkspace, input GitHubProjectFieldInput) (*GitHubProjectField, error) {
	name := strings.TrimSpace(input.Name)
	dataType := strings.TrimSpace(input.DataType)
	if name == "" || dataType == "" {
		return nil, fmt.Errorf("project field name and data type are required")
	}
	for _, field := range project.Fields {
		if !strings.EqualFold(strings.TrimSpace(field.Name), name) {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(field.DataType), dataType) {
			return nil, fmt.Errorf("GitHub Project field %q has type %q; expected %q", field.Name, field.DataType, dataType)
		}
		missing := missingProjectFieldOptions(field, input.Options)
		if len(missing) > 0 {
			return nil, fmt.Errorf("GitHub Project field %q is missing options %s; add them manually or use a new Project because Plan does not edit existing single-select options without restoring affected item selections", field.Name, strings.Join(missing, ", "))
		}
		copy := field
		copy.Options = copyStringMap(field.Options)
		return &copy, nil
	}
	variables := map[string]any{
		"projectId": project.ID,
		"name":      name,
		"dataType":  dataType,
	}
	if strings.EqualFold(dataType, "SINGLE_SELECT") {
		options := make([]map[string]string, 0, len(input.Options))
		for _, option := range input.Options {
			option = strings.TrimSpace(option)
			if option == "" {
				continue
			}
			options = append(options, map[string]string{
				"name":        option,
				"color":       "GRAY",
				"description": "",
			})
		}
		variables["singleSelectOptions"] = options
	}
	payload := map[string]any{
		"query": `mutation($projectId:ID!, $name:String!, $dataType:ProjectV2CustomFieldType!, $singleSelectOptions:[ProjectV2SingleSelectFieldOptionInput!]) {
  createProjectV2Field(input:{projectId:$projectId, name:$name, dataType:$dataType, singleSelectOptions:$singleSelectOptions}) {
    projectV2Field {
      __typename
      ... on ProjectV2Field {
        id
        name
        dataType
      }
      ... on ProjectV2SingleSelectField {
        id
        name
        dataType
        options {
          id
          name
        }
      }
    }
  }
}`,
		"variables": variables,
	}
	var response struct {
		Data struct {
			CreateProjectV2Field struct {
				ProjectV2Field projectV2FieldPayload `json:"projectV2Field"`
			} `json:"createProjectV2Field"`
		} `json:"data"`
	}
	if err := c.graphql(projectDir, payload, &response); err != nil {
		return nil, err
	}
	field := projectFieldFromPayload(response.Data.CreateProjectV2Field.ProjectV2Field)
	if strings.TrimSpace(field.ID) == "" {
		return nil, fmt.Errorf("created GitHub Project field %q did not include an id", name)
	}
	return &field, nil
}

func (c *cliGitHubClient) AddProjectItemByIssue(projectDir, repo, projectID string, issueNumber int) (*GitHubProjectItem, error) {
	issueID := c.cachedIssueNodeID(repo, issueNumber)
	if strings.TrimSpace(issueID) == "" {
		var err error
		issueID, err = c.issueID(projectDir, repo, issueNumber)
		if err != nil {
			return nil, err
		}
		c.cacheIssueNodeIDValue(repo, issueNumber, issueID)
	}
	payload := map[string]any{
		"query": `mutation($projectId:ID!, $contentId:ID!) {
  addProjectV2ItemById(input:{projectId:$projectId, contentId:$contentId}) {
    item {
      id
    }
  }
}`,
		"variables": map[string]any{
			"projectId": projectID,
			"contentId": issueID,
		},
	}
	var response struct {
		Data struct {
			AddProjectV2ItemByID struct {
				Item struct {
					ID string `json:"id"`
				} `json:"item"`
			} `json:"addProjectV2ItemById"`
		} `json:"data"`
	}
	if err := c.graphql(projectDir, payload, &response); err != nil {
		return nil, err
	}
	if strings.TrimSpace(response.Data.AddProjectV2ItemByID.Item.ID) == "" {
		return nil, fmt.Errorf("GitHub Project item for issue #%d did not include an id", issueNumber)
	}
	return &GitHubProjectItem{ID: response.Data.AddProjectV2ItemByID.Item.ID, IssueNumber: issueNumber}, nil
}

func (c *cliGitHubClient) SetProjectItemField(projectDir, projectID, itemID string, field GitHubProjectField, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	fieldValue := map[string]any{}
	switch strings.ToUpper(strings.TrimSpace(field.DataType)) {
	case "SINGLE_SELECT":
		optionID := field.Options[value]
		if strings.TrimSpace(optionID) == "" {
			return fmt.Errorf("GitHub Project field %q does not have option %q", field.Name, value)
		}
		fieldValue["singleSelectOptionId"] = optionID
	case "TEXT":
		fieldValue["text"] = value
	default:
		return fmt.Errorf("unsupported GitHub Project field type %q for field %q", field.DataType, field.Name)
	}
	payload := map[string]any{
		"query": `mutation($projectId:ID!, $itemId:ID!, $fieldId:ID!, $value:ProjectV2FieldValue!) {
  updateProjectV2ItemFieldValue(input:{projectId:$projectId, itemId:$itemId, fieldId:$fieldId, value:$value}) {
    projectV2Item {
      id
    }
  }
}`,
		"variables": map[string]any{
			"projectId": projectID,
			"itemId":    itemID,
			"fieldId":   field.ID,
			"value":     fieldValue,
		},
	}
	return c.graphql(projectDir, payload, nil)
}

type projectV2Payload struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	URL    string `json:"url"`
	Title  string `json:"title"`
	Fields struct {
		Nodes []projectV2FieldPayload `json:"nodes"`
	} `json:"fields"`
}

type projectV2FieldPayload struct {
	Typename string `json:"__typename"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	DataType string `json:"dataType"`
	Options  []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"options"`
}

func projectWorkspaceFromPayload(owner string, payload projectV2Payload) *GitHubProjectWorkspace {
	project := &GitHubProjectWorkspace{
		Owner:  strings.TrimSpace(owner),
		Number: payload.Number,
		ID:     payload.ID,
		URL:    payload.URL,
		Title:  payload.Title,
	}
	for _, node := range payload.Fields.Nodes {
		field := projectFieldFromPayload(node)
		if strings.TrimSpace(field.ID) == "" {
			continue
		}
		project.Fields = append(project.Fields, field)
	}
	return project
}

func projectFieldFromPayload(payload projectV2FieldPayload) GitHubProjectField {
	field := GitHubProjectField{
		ID:       payload.ID,
		Name:     payload.Name,
		DataType: payload.DataType,
	}
	if len(payload.Options) > 0 {
		field.Options = map[string]string{}
		for _, option := range payload.Options {
			if strings.TrimSpace(option.Name) == "" || strings.TrimSpace(option.ID) == "" {
				continue
			}
			field.Options[option.Name] = option.ID
		}
	}
	return field
}

func missingProjectFieldOptions(field GitHubProjectField, required []string) []string {
	if len(required) == 0 {
		return nil
	}
	var missing []string
	for _, option := range required {
		option = strings.TrimSpace(option)
		if option == "" {
			continue
		}
		if strings.TrimSpace(field.Options[option]) == "" {
			missing = append(missing, option)
		}
	}
	return missing
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

func (c *cliGitHubClient) projectOwnerAndRepositoryIDs(projectDir, repo, projectOwner string) (string, string, error) {
	repoOwner, repoName, err := splitRepo(repo)
	if err != nil {
		return "", "", err
	}
	payload := map[string]any{
		"query": `query($projectOwner:String!, $repoOwner:String!, $repoName:String!) {
  owner: repositoryOwner(login:$projectOwner) {
    id
  }
  repository(owner:$repoOwner, name:$repoName) {
    id
  }
}`,
		"variables": map[string]any{
			"projectOwner": projectOwner,
			"repoOwner":    repoOwner,
			"repoName":     repoName,
		},
	}
	var response struct {
		Data struct {
			Owner *struct {
				ID string `json:"id"`
			} `json:"owner"`
			Repository *struct {
				ID string `json:"id"`
			} `json:"repository"`
		} `json:"data"`
	}
	if err := c.graphql(projectDir, payload, &response); err != nil {
		return "", "", err
	}
	if response.Data.Owner == nil || strings.TrimSpace(response.Data.Owner.ID) == "" {
		return "", "", fmt.Errorf("GitHub Project owner %q not found", projectOwner)
	}
	if response.Data.Repository == nil || strings.TrimSpace(response.Data.Repository.ID) == "" {
		return "", "", fmt.Errorf("repository %q not found", repo)
	}
	return response.Data.Owner.ID, response.Data.Repository.ID, nil
}

func (c *cliGitHubClient) getProjectWorkspaceByID(projectDir, projectID, fallbackOwner string) (*GitHubProjectWorkspace, error) {
	payload := map[string]any{
		"query": `query($projectId:ID!) {
  node(id:$projectId) {
    ... on ProjectV2 {
      id
      number
      url
      title
      owner {
        ... on User {
          login
        }
        ... on Organization {
          login
        }
      }
      fields(first:100) {
        nodes {
          __typename
          ... on ProjectV2Field {
            id
            name
            dataType
          }
          ... on ProjectV2SingleSelectField {
            id
            name
            dataType
            options {
              id
              name
            }
          }
        }
      }
    }
  }
}`,
		"variables": map[string]any{
			"projectId": projectID,
		},
	}
	var response struct {
		Data struct {
			Node *struct {
				projectV2Payload
				Owner struct {
					Login string `json:"login"`
				} `json:"owner"`
			} `json:"node"`
		} `json:"data"`
	}
	if err := c.graphql(projectDir, payload, &response); err != nil {
		return nil, err
	}
	if response.Data.Node == nil || strings.TrimSpace(response.Data.Node.ID) == "" {
		return nil, fmt.Errorf("GitHub Project id %q not found", projectID)
	}
	owner := strings.TrimSpace(response.Data.Node.Owner.Login)
	if owner == "" {
		owner = fallbackOwner
	}
	return projectWorkspaceFromPayload(owner, response.Data.Node.projectV2Payload), nil
}

func (c *cliGitHubClient) issueID(projectDir, repo string, issueNumber int) (string, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return "", err
	}
	payload := map[string]any{
		"query": `query($owner:String!, $name:String!, $issueNumber:Int!) {
  repository(owner:$owner, name:$name) {
    issue(number:$issueNumber) {
      id
    }
  }
}`,
		"variables": map[string]any{
			"owner":       owner,
			"name":        name,
			"issueNumber": issueNumber,
		},
	}
	var response struct {
		Data struct {
			Repository struct {
				Issue struct {
					ID string `json:"id"`
				} `json:"issue"`
			} `json:"repository"`
		} `json:"data"`
	}
	if err := c.graphql(projectDir, payload, &response); err != nil {
		return "", err
	}
	if strings.TrimSpace(response.Data.Repository.Issue.ID) == "" {
		return "", fmt.Errorf("could not resolve issue node id for #%d", issueNumber)
	}
	return response.Data.Repository.Issue.ID, nil
}

func (c *cliGitHubClient) cacheIssueNodeID(repo string, issue *GitHubIssue) {
	if issue == nil {
		return
	}
	c.cacheIssueNodeIDValue(repo, issue.Number, issue.NodeID)
}

func (c *cliGitHubClient) cacheIssueNodeIDValue(repo string, issueNumber int, nodeID string) {
	repo = strings.TrimSpace(repo)
	nodeID = strings.TrimSpace(nodeID)
	if repo == "" || issueNumber <= 0 || nodeID == "" {
		return
	}
	if c.issueNodeIDs == nil {
		c.issueNodeIDs = map[string]string{}
	}
	c.issueNodeIDs[issueNodeCacheKey(repo, issueNumber)] = nodeID
}

func (c *cliGitHubClient) cachedIssueNodeID(repo string, issueNumber int) string {
	if c.issueNodeIDs == nil {
		return ""
	}
	return c.issueNodeIDs[issueNodeCacheKey(repo, issueNumber)]
}

func issueNodeCacheKey(repo string, issueNumber int) string {
	return fmt.Sprintf("%s#%d", strings.TrimSpace(repo), issueNumber)
}

func repoFromIssueAPIPath(apiPath string) string {
	trimmed := strings.TrimPrefix(strings.TrimSpace(apiPath), "repos/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[2] != "issues" {
		return ""
	}
	return parts[0] + "/" + parts[1]
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
		NodeID    string  `json:"node_id"`
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
		NodeID: item.NodeID,
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
		NodeID    string  `json:"id"`
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
			NodeID: item.NodeID,
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
