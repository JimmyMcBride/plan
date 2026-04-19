package planning

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/workspace"
)

type stubGitHubClient struct {
	preflight    *GitHubRepoInfo
	preflightErr error
	context      *GitHubContext
	issues       map[int]*GitHubIssue
	nextIssue    int
	lastCreate   GitHubIssueInput
	lastUpdate   GitHubIssueInput
}

func (s *stubGitHubClient) Preflight(projectDir string) (*GitHubRepoInfo, error) {
	if s.preflightErr != nil {
		return nil, s.preflightErr
	}
	return s.preflight, nil
}

func (s *stubGitHubClient) CurrentContext(projectDir string) (*GitHubContext, error) {
	if s.context == nil {
		panic("unexpected CurrentContext call")
	}
	return s.context, nil
}

func (s *stubGitHubClient) CreateIssue(projectDir, repo string, input GitHubIssueInput) (*GitHubIssue, error) {
	s.lastCreate = input
	if s.issues == nil {
		s.issues = map[int]*GitHubIssue{}
	}
	if s.nextIssue == 0 {
		s.nextIssue = 101
	}
	issue := &GitHubIssue{
		Number: s.nextIssue,
		URL:    fmt.Sprintf("https://github.com/%s/issues/%d", repo, s.nextIssue),
		Title:  input.Title,
		Body:   input.Body,
		State:  "open",
		Labels: append([]string(nil), input.Labels...),
	}
	if strings.TrimSpace(input.State) != "" {
		issue.State = input.State
	}
	s.issues[issue.Number] = issue
	s.nextIssue++
	return issue, nil
}

func (s *stubGitHubClient) UpdateIssue(projectDir, repo string, issueNumber int, input GitHubIssueInput) (*GitHubIssue, error) {
	s.lastUpdate = input
	issue, ok := s.issues[issueNumber]
	if !ok {
		panic("unexpected UpdateIssue call")
	}
	issue.Title = input.Title
	issue.Body = input.Body
	if strings.TrimSpace(input.State) != "" {
		issue.State = input.State
	}
	issue.Labels = append([]string(nil), input.Labels...)
	return issue, nil
}

func (s *stubGitHubClient) GetIssue(projectDir, repo string, issueNumber int) (*GitHubIssue, error) {
	issue, ok := s.issues[issueNumber]
	if !ok {
		panic("unexpected GetIssue call")
	}
	copy := *issue
	copy.Labels = append([]string(nil), issue.Labels...)
	return &copy, nil
}

func TestEnableGitHubBackendPersistsRepoConfigAfterPreflight(t *testing.T) {
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient {
		return &stubGitHubClient{
			preflight: &GitHubRepoInfo{
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
	manager := New(ws)

	result, err := manager.EnableGitHubBackend()
	if err != nil {
		t.Fatal(err)
	}
	if result.Backend != "github" || result.Repo != "JimmyMcBride/plan" || result.DefaultBranch != "main" {
		t.Fatalf("unexpected enable result: %+v", result)
	}

	meta, err := ws.ReadWorkspaceMeta()
	if err != nil {
		t.Fatal(err)
	}
	if meta.StoryBackend != workspace.StoryBackendGitHub {
		t.Fatalf("expected GitHub story backend: %+v", meta)
	}

	state, err := ws.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	if state.Repo != "JimmyMcBride/plan" || state.DefaultBranch != "main" {
		t.Fatalf("unexpected github state: %+v", state)
	}
	if state.LastEnabledAt == "" || state.LastUpdatedAt == "" {
		t.Fatalf("expected github state timestamps: %+v", state)
	}
}

func TestEnableGitHubBackendFailsWhenLocalStoriesAlreadyExist(t *testing.T) {
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient {
		return &stubGitHubClient{
			preflight: &GitHubRepoInfo{
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
	if err := os.WriteFile(filepath.Join(root, ".plan", "stories", "existing.md"), []byte("# Existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	manager := New(ws)
	if _, err := manager.EnableGitHubBackend(); err == nil {
		t.Fatal("expected local stories to block GitHub enablement")
	} else if !strings.Contains(err.Error(), "local story notes still exist") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnableGitHubBackendPropagatesPreflightFailures(t *testing.T) {
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient {
		return &stubGitHubClient{preflightErr: errors.New("gh auth status failed; run `gh auth login` before enabling GitHub mode")}
	})
	t.Cleanup(reset)

	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}

	manager := New(ws)
	if _, err := manager.EnableGitHubBackend(); err == nil {
		t.Fatal("expected preflight error")
	} else if !strings.Contains(err.Error(), "gh auth status failed") {
		t.Fatalf("unexpected preflight error: %v", err)
	}
}

func TestCreateStoryUsesGitHubIssueStorageWhenBackendEnabled(t *testing.T) {
	client := &stubGitHubClient{
		preflight: &GitHubRepoInfo{
			Repo:          "JimmyMcBride/plan",
			RepoURL:       "https://github.com/JimmyMcBride/plan",
			DefaultBranch: "main",
		},
		context: &GitHubContext{
			Repo: GitHubRepoInfo{
				Repo:          "JimmyMcBride/plan",
				RepoURL:       "https://github.com/JimmyMcBride/plan",
				DefaultBranch: "main",
			},
			CurrentBranch: "main",
			CurrentSHA:    "abc123def456",
		},
	}
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	t.Cleanup(reset)

	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)
	if _, err := manager.EnableGitHubBackend(); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SetSpecStatus("billing", "approved"); err != nil {
		t.Fatal(err)
	}

	story, err := manager.CreateStory(
		"billing",
		"Implement invoices",
		"Create invoice generation flow",
		[]string{"Generate invoices from line items"},
		[]string{"Run focused billing tests"},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if story.Path != "https://github.com/JimmyMcBride/plan/issues/101" {
		t.Fatalf("unexpected GitHub story path: %s", story.Path)
	}
	if _, err := os.Stat(filepath.Join(root, ".plan", "stories", "implement-invoices.md")); !os.IsNotExist(err) {
		t.Fatalf("expected no local story note in GitHub mode, got err=%v", err)
	}

	state, err := ws.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	record, ok := state.Stories["implement-invoices"]
	if !ok {
		t.Fatalf("expected GitHub story record to be stored: %+v", state.Stories)
	}
	if record.IssueNumber != 101 || record.Status != "todo" {
		t.Fatalf("unexpected issue-backed story record: %+v", record)
	}
	if !strings.Contains(client.lastCreate.Body, "## Planning Links") || !strings.Contains(client.lastCreate.Body, "## Dependencies") {
		t.Fatalf("expected issue contract in created body:\n%s", client.lastCreate.Body)
	}

	items, err := manager.ListStories("", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Backend != "github" || items[0].IssueNumber != 101 {
		t.Fatalf("unexpected GitHub story list: %+v", items)
	}
}

func TestUpdateStoryMutatesIssueBackedRecordWithoutLocalMarkdown(t *testing.T) {
	client := &stubGitHubClient{
		preflight: &GitHubRepoInfo{
			Repo:          "JimmyMcBride/plan",
			RepoURL:       "https://github.com/JimmyMcBride/plan",
			DefaultBranch: "main",
		},
		context: &GitHubContext{
			Repo: GitHubRepoInfo{
				Repo:          "JimmyMcBride/plan",
				RepoURL:       "https://github.com/JimmyMcBride/plan",
				DefaultBranch: "main",
			},
			CurrentBranch: "main",
			CurrentSHA:    "abc123def456",
		},
	}
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	t.Cleanup(reset)

	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)
	if _, err := manager.EnableGitHubBackend(); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SetSpecStatus("billing", "approved"); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateStory(
		"billing",
		"Implement invoices",
		"Create invoice generation flow",
		[]string{"Generate invoices from line items"},
		[]string{"Run focused billing tests"},
		nil,
	); err != nil {
		t.Fatal(err)
	}

	updated, err := manager.UpdateStory("implement-invoices", StoryChanges{
		Status:          "in_progress",
		AddResources:    []string{"Issue owner: billing"},
		SetBlockers:     []string{"seed-billing-data"},
		AddVerification: []string{"Check GitHub issue body"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Metadata["status"] != "blocked" && updated.Metadata["status"] != "in_progress" {
		t.Fatalf("expected GitHub story note metadata to reflect updated status: %+v", updated.Metadata)
	}
	if _, err := os.Stat(filepath.Join(root, ".plan", "stories", "implement-invoices.md")); !os.IsNotExist(err) {
		t.Fatalf("expected no local story note in GitHub mode, got err=%v", err)
	}

	state, err := ws.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	record := state.Stories["implement-invoices"]
	if record.Status != "in_progress" {
		t.Fatalf("expected stored GitHub story status to update: %+v", record)
	}
	if len(record.Dependencies) != 1 || record.Dependencies[0] != "seed-billing-data" {
		t.Fatalf("expected dependencies to persist in record: %+v", record)
	}
	if !strings.Contains(client.lastUpdate.Body, "seed-billing-data") {
		t.Fatalf("expected dependency contract in updated issue body:\n%s", client.lastUpdate.Body)
	}
}
