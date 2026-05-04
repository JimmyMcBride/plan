package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"plan/internal/planning"
	"plan/internal/workspace"
)

type stubGitHubEnableClient struct {
	preflight *planning.GitHubRepoInfo
	err       error
}

func (s *stubGitHubEnableClient) Preflight(projectDir string) (*planning.GitHubRepoInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.preflight, nil
}

func (s *stubGitHubEnableClient) CurrentContext(projectDir string) (*planning.GitHubContext, error) {
	panic("unexpected CurrentContext call")
}

func (s *stubGitHubEnableClient) CreateIssue(projectDir, repo string, input planning.GitHubIssueInput) (*planning.GitHubIssue, error) {
	panic("unexpected CreateIssue call")
}

func (s *stubGitHubEnableClient) UpdateIssue(projectDir, repo string, issueNumber int, input planning.GitHubIssueInput) (*planning.GitHubIssue, error) {
	panic("unexpected UpdateIssue call")
}

func (s *stubGitHubEnableClient) GetIssue(projectDir, repo string, issueNumber int) (*planning.GitHubIssue, error) {
	panic("unexpected GetIssue call")
}

func (s *stubGitHubEnableClient) ListIssuesByLabel(projectDir, repo string, labels []string) ([]planning.GitHubIssue, error) {
	panic("unexpected ListIssuesByLabel call")
}

func (s *stubGitHubEnableClient) EnsureLabel(projectDir, repo string, input planning.GitHubLabelInput) error {
	panic("unexpected EnsureLabel call")
}

func (s *stubGitHubEnableClient) FindMilestone(projectDir, repo, title string) (*planning.GitHubMilestone, error) {
	panic("unexpected FindMilestone call")
}

func (s *stubGitHubEnableClient) CreateMilestone(projectDir, repo string, input planning.GitHubMilestoneInput) (*planning.GitHubMilestone, error) {
	panic("unexpected CreateMilestone call")
}

func (s *stubGitHubEnableClient) GetDiscussion(projectDir, repo string, number int) (*planning.GitHubDiscussion, error) {
	panic("unexpected GetDiscussion call")
}

func (s *stubGitHubEnableClient) UpdateDiscussionBody(projectDir, repo string, number int, body string) (*planning.GitHubDiscussion, error) {
	panic("unexpected UpdateDiscussionBody call")
}

func (s *stubGitHubEnableClient) AddSubIssue(projectDir, repo string, issueNumber, subIssueNumber int) error {
	panic("unexpected AddSubIssue call")
}

func (s *stubGitHubEnableClient) AddBlockedBy(projectDir, repo string, issueNumber, blockingIssueNumber int) error {
	panic("unexpected AddBlockedBy call")
}

func (s *stubGitHubEnableClient) CreateProjectWorkspace(projectDir, repo string, input planning.GitHubProjectWorkspaceInput) (*planning.GitHubProjectWorkspace, error) {
	panic("unexpected CreateProjectWorkspace call")
}

func (s *stubGitHubEnableClient) GetProjectWorkspace(projectDir, repo string, ref planning.GitHubProjectReference) (*planning.GitHubProjectWorkspace, error) {
	panic("unexpected GetProjectWorkspace call")
}

func (s *stubGitHubEnableClient) EnsureProjectField(projectDir string, project planning.GitHubProjectWorkspace, input planning.GitHubProjectFieldInput) (*planning.GitHubProjectField, error) {
	panic("unexpected EnsureProjectField call")
}

func (s *stubGitHubEnableClient) AddProjectItemByIssue(projectDir, repo, projectID string, issueNumber int) (*planning.GitHubProjectItem, error) {
	panic("unexpected AddProjectItemByIssue call")
}

func (s *stubGitHubEnableClient) GetProjectItemByIssue(projectDir, repo, projectID string, issueNumber int) (*planning.GitHubProjectItem, error) {
	panic("unexpected GetProjectItemByIssue call")
}

func (s *stubGitHubEnableClient) SetProjectItemField(projectDir, projectID, itemID string, field planning.GitHubProjectField, value string) error {
	panic("unexpected SetProjectItemField call")
}

func TestGitHubEnableCommandPrintsBackendSummary(t *testing.T) {
	reset := planning.SetGitHubClientFactoryForTesting(func() planning.GitHubClient {
		return &stubGitHubEnableClient{
			preflight: &planning.GitHubRepoInfo{
				Repo:          "JimmyMcBride/plan",
				RepoURL:       "https://github.com/JimmyMcBride/plan",
				DefaultBranch: "main",
			},
		}
	})
	t.Cleanup(reset)

	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"--project", root, "github", "enable"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected github enable to succeed: %v\n%s", err, buf.String())
	}

	output := buf.String()
	if !containsLine(output, "github_backend: github") || !containsLine(output, "repo: JimmyMcBride/plan") || !containsLine(output, "default_branch: main") {
		t.Fatalf("unexpected github enable output:\n%s", output)
	}
}

func TestGitHubAdoptCommandPrintsJSON(t *testing.T) {
	client := &stubGitHubStoryClient{
		preflight: &planning.GitHubRepoInfo{
			Repo:          "JimmyMcBride/plan",
			RepoURL:       "https://github.com/JimmyMcBride/plan",
			DefaultBranch: "develop",
		},
		context: &planning.GitHubContext{
			Repo: planning.GitHubRepoInfo{
				Repo:          "JimmyMcBride/plan",
				RepoURL:       "https://github.com/JimmyMcBride/plan",
				DefaultBranch: "develop",
			},
			CurrentBranch: "develop",
			CurrentSHA:    "abc123",
		},
		issues: map[int]*planning.GitHubIssue{
			201: {Number: 201, URL: "https://github.com/JimmyMcBride/plan/issues/201", Title: "Adopt Command", State: "open"},
			202: {Number: 202, URL: "https://github.com/JimmyMcBride/plan/issues/202", Title: "Adopt command schema", State: "open"},
			203: {Number: 203, URL: "https://github.com/JimmyMcBride/plan/issues/203", Title: "Adopt command CLI", State: "open"},
		},
		discussions: map[int]*planning.GitHubDiscussion{
			90: {
				Number: 90,
				URL:    "https://github.com/JimmyMcBride/plan/discussions/90",
				Title:  "Adopt Command",
				Body: strings.Join([]string{
					"## Problem",
					"Existing issues need Plan metadata adoption.",
					"",
					"## Goals",
					"Adopt the issue set.",
					"",
					"## Non-Goals",
					"Do not create unrelated issues.",
					"",
					"## Constraints",
					"Keep issue order explicit.",
					"",
					"## Proposed Shape",
					"Use two spec issues.",
					"",
					"## Spec Split",
					"- Adopt command schema",
					"- Adopt command CLI",
				}, "\n"),
			},
		},
	}
	reset := planning.SetGitHubClientFactoryForTesting(func() planning.GitHubClient { return client })
	t.Cleanup(reset)

	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"--project", root, "github", "adopt", "--discussion", "90", "--issues", "201,202,203", "--format", "json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected github adopt to succeed: %v\n%s", err, buf.String())
	}

	var payload planning.GitHubAdoptResult
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON output, got %v\n%s", err, buf.String())
	}
	if payload.Initiative == nil || len(payload.Specs) != 2 {
		t.Fatalf("unexpected adopt payload: %+v", payload)
	}
}

func TestGitHubProjectStatusCommandMovesIssueCard(t *testing.T) {
	client := &stubGitHubStoryClient{
		preflight: &planning.GitHubRepoInfo{Repo: "JimmyMcBride/plan", RepoURL: "https://github.com/JimmyMcBride/plan", DefaultBranch: "develop"},
		context: &planning.GitHubContext{
			Repo:          planning.GitHubRepoInfo{Repo: "JimmyMcBride/plan", RepoURL: "https://github.com/JimmyMcBride/plan", DefaultBranch: "develop"},
			CurrentBranch: "codex/status",
			CurrentSHA:    "abc123",
		},
		issues: map[int]*planning.GitHubIssue{
			68: {Number: 68, URL: "https://github.com/JimmyMcBride/plan/issues/68", Title: "Status", State: "open"},
		},
		projects: map[int]*planning.GitHubProjectWorkspace{
			12: {
				Owner:  "JimmyMcBride",
				Number: 12,
				ID:     "PVT_workspace",
				URL:    "https://github.com/users/JimmyMcBride/projects/12",
				Title:  "Workspace",
				Fields: []planning.GitHubProjectField{
					{ID: "PVTF_status", Name: "Status", DataType: "SINGLE_SELECT", Options: map[string]string{"Todo": "todo", "In Progress": "progress", "In Review": "review", "Done": "done"}},
				},
			},
		},
		projectItems: map[int]*planning.GitHubProjectItem{
			68: {ID: "PVTI_68", IssueNumber: 68, ProjectID: "PVT_workspace", Values: map[string]string{"Status": "Todo"}},
		},
	}
	reset := planning.SetGitHubClientFactoryForTesting(func() planning.GitHubClient { return client })
	t.Cleanup(reset)

	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
	if _, err := manager.SetSourceMode(workspace.SourceOfTruthGitHub); err != nil {
		t.Fatal(err)
	}
	state, err := ws.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	state.Repo = "JimmyMcBride/plan"
	state.Planning["status"] = workspace.GitHubPlanningRecord{
		Slug:            "status",
		Kind:            "spec",
		Title:           "Status",
		IssueNumber:     68,
		Readiness:       "ready",
		MilestoneNumber: 7,
		MilestoneTitle:  "Workspace",
	}
	state.ProjectDecisions["workspace"] = workspace.GitHubProjectDecisionRecord{
		Slug:            "workspace",
		Decision:        "connect",
		MilestoneNumber: 7,
		MilestoneTitle:  "Workspace",
		ProjectOwner:    "JimmyMcBride",
		ProjectNumber:   12,
		ProjectID:       "PVT_workspace",
		ProjectURL:      "https://github.com/users/JimmyMcBride/projects/12",
		FieldIDs: map[string]string{
			"Type":   "PVTF_type",
			"Stage":  "PVTF_stage",
			"Ready":  "PVTF_ready",
			"Status": "PVTF_status",
			"Area":   "PVTF_area",
		},
	}
	if err := ws.WriteGitHubState(*state); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"--project", root, "github", "project", "status", "--issue", "68", "--set", "in-progress"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected project status command to succeed: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "status: In Progress") {
		t.Fatalf("unexpected output:\n%s", buf.String())
	}
	if len(client.projectValues) != 1 || client.projectValues[0] != "Status=In Progress" {
		t.Fatalf("expected status update, got %+v", client.projectValues)
	}
}

func containsLine(output, expected string) bool {
	for _, line := range bytes.Split([]byte(output), []byte("\n")) {
		if string(line) == expected {
			return true
		}
	}
	return false
}
