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
	milestones   map[string]*GitHubMilestone
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
	if input.Milestone != nil {
		for _, milestone := range s.milestones {
			if milestone.Number == *input.Milestone {
				copy := *milestone
				issue.Milestone = &copy
				break
			}
		}
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
	if input.Labels != nil {
		issue.Labels = append([]string(nil), input.Labels...)
	}
	if input.Milestone != nil {
		issue.Milestone = nil
		for _, milestone := range s.milestones {
			if milestone.Number == *input.Milestone {
				copy := *milestone
				issue.Milestone = &copy
				break
			}
		}
	}
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

func (s *stubGitHubClient) FindMilestone(projectDir, repo, title string) (*GitHubMilestone, error) {
	if s.milestones == nil {
		return nil, nil
	}
	milestone, ok := s.milestones[title]
	if !ok {
		return nil, nil
	}
	copy := *milestone
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
	if err := os.MkdirAll(filepath.Join(root, ".plan", "stories"), 0o755); err != nil {
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

func TestCreateGitHubStoryMapsSpecInitiativeToMilestoneWhenPresent(t *testing.T) {
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
		milestones: map[string]*GitHubMilestone{
			"Guide Packet Foundation": {Number: 12, Title: "Guide Packet Foundation"},
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
	if _, err := manager.SetSpecInitiative("billing", InitiativeRef{
		Slug:  "guide-packet-foundation",
		Title: "Guide Packet Foundation",
	}); err != nil {
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
	if client.lastCreate.Milestone == nil || *client.lastCreate.Milestone != 12 {
		t.Fatalf("expected GitHub issue create to include initiative milestone: %+v", client.lastCreate)
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

func TestGitHubIssueContractIncludesMetadataAndPlanningSections(t *testing.T) {
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
	manager := newGitHubStoryManager(t, root)
	if _, err := manager.EnableGitHubBackend(); err != nil {
		t.Fatal(err)
	}
	seedApprovedGitHubEpic(t, manager)
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

	body := client.lastCreate.Body
	if !strings.Contains(body, planIssueBlockStart) || !strings.Contains(body, "## Planning Links") || !strings.Contains(body, "## Dependencies") {
		t.Fatalf("expected plan-managed issue sections:\n%s", body)
	}
	meta := parseGitHubStoryMetadata(body)
	if meta["slug"] != "implement-invoices" || meta["epic"] != "billing" || meta["spec"] != "billing" {
		t.Fatalf("unexpected issue metadata block: %+v\n%s", meta, body)
	}
}

func TestCreateGitHubStoryUsesShaLinksAndPlanningPRBeforeMerge(t *testing.T) {
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
			CurrentBranch: "feat/planning-branch",
			CurrentSHA:    "abc123def456",
			PlanningPR: &GitHubPullRequest{
				Number:   42,
				URL:      "https://github.com/JimmyMcBride/plan/pull/42",
				State:    "OPEN",
				HeadRef:  "feat/planning-branch",
				BaseRef:  "main",
				IsMerged: false,
			},
		},
	}
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	t.Cleanup(reset)

	root := t.TempDir()
	manager := newGitHubStoryManager(t, root)
	if _, err := manager.EnableGitHubBackend(); err != nil {
		t.Fatal(err)
	}
	seedApprovedGitHubEpic(t, manager)
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

	body := client.lastCreate.Body
	if !strings.Contains(body, "/blob/abc123def456/.plan/specs/billing.md") {
		t.Fatalf("expected SHA permalink in pre-merge issue body:\n%s", body)
	}
	if strings.Contains(body, "/blob/feat/planning-branch/") {
		t.Fatalf("expected branch-name links to stay out of issue body:\n%s", body)
	}
	if !strings.Contains(body, "Planning PR: [#42](https://github.com/JimmyMcBride/plan/pull/42)") {
		t.Fatalf("expected planning PR link in issue body:\n%s", body)
	}
}

func TestCreateGitHubStoryRequiresPlanningPROffDefaultBranch(t *testing.T) {
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
			CurrentBranch: "feat/planning-branch",
			CurrentSHA:    "abc123def456",
		},
	}
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	t.Cleanup(reset)

	root := t.TempDir()
	manager := newGitHubStoryManager(t, root)
	if _, err := manager.EnableGitHubBackend(); err != nil {
		t.Fatal(err)
	}
	seedApprovedGitHubEpic(t, manager)
	if _, err := manager.CreateStory(
		"billing",
		"Implement invoices",
		"Create invoice generation flow",
		[]string{"Generate invoices from line items"},
		[]string{"Run focused billing tests"},
		nil,
	); err == nil {
		t.Fatal("expected planning branch without PR to be rejected")
	} else if !strings.Contains(err.Error(), "has no planning PR") {
		t.Fatalf("unexpected planning-branch error: %v", err)
	}
}

func TestReconcileGitHubStoriesPromotesMainLinksAndPreservesUserText(t *testing.T) {
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
			CurrentBranch: "feat/planning-branch",
			CurrentSHA:    "abc123def456",
			PlanningPR: &GitHubPullRequest{
				Number:   42,
				URL:      "https://github.com/JimmyMcBride/plan/pull/42",
				State:    "OPEN",
				HeadRef:  "feat/planning-branch",
				BaseRef:  "main",
				IsMerged: false,
			},
		},
	}
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	t.Cleanup(reset)

	root := t.TempDir()
	manager := newGitHubStoryManager(t, root)
	if _, err := manager.EnableGitHubBackend(); err != nil {
		t.Fatal(err)
	}
	seedApprovedGitHubEpic(t, manager)
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

	client.issues[101].Body += "\n\nUser note outside managed block.\n"
	client.context.CurrentBranch = "main"
	client.context.PlanningPR = nil

	result, err := manager.ReconcileGitHubStories(GitHubReconcileOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.UpdatedIssues) != 1 {
		t.Fatalf("expected one reconciled issue: %+v", result)
	}
	issue := client.issues[101]
	if !strings.Contains(issue.Body, "/blob/main/.plan/specs/billing.md") {
		t.Fatalf("expected canonical main link after reconcile:\n%s", issue.Body)
	}
	if strings.Contains(issue.Body, "/blob/abc123def456/") {
		t.Fatalf("expected SHA links to be removed after reconcile:\n%s", issue.Body)
	}
	if !strings.Contains(issue.Body, "User note outside managed block.") {
		t.Fatalf("expected user text outside managed block to survive reconcile:\n%s", issue.Body)
	}

	state, err := manager.workspace.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	record := state.Stories["implement-invoices"]
	if !record.PlanningPRMerged || record.DocRefMode != "main" || record.DocRef != "main" {
		t.Fatalf("expected record to reconcile to default branch links: %+v", record)
	}
}

func TestReconcileGitHubStoriesNoOpDoesNotRewriteStateFile(t *testing.T) {
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
	manager := newGitHubStoryManager(t, root)
	if _, err := manager.EnableGitHubBackend(); err != nil {
		t.Fatal(err)
	}
	seedApprovedGitHubEpic(t, manager)
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

	if _, err := manager.ReconcileGitHubStories(GitHubReconcileOptions{}); err != nil {
		t.Fatal(err)
	}
	info, err := manager.workspace.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(info.GitHubFile)
	if err != nil {
		t.Fatal(err)
	}

	result, err := manager.ReconcileGitHubStories(GitHubReconcileOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.UpdatedIssues) != 0 {
		t.Fatalf("expected no issue updates on no-op reconcile: %+v", result)
	}

	after, err := os.ReadFile(info.GitHubFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("expected no-op reconcile to leave github state unchanged")
	}
}

func TestReconcileGitHubStoriesRefreshesRepoDefaultBranchInState(t *testing.T) {
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
	manager := newGitHubStoryManager(t, root)
	if _, err := manager.EnableGitHubBackend(); err != nil {
		t.Fatal(err)
	}
	seedApprovedGitHubEpic(t, manager)
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

	client.context.Repo.DefaultBranch = "develop"
	client.context.CurrentBranch = "develop"

	result, err := manager.ReconcileGitHubStories(GitHubReconcileOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.DefaultBranch != "develop" {
		t.Fatalf("expected reconcile result to report develop default branch: %+v", result)
	}

	state, err := manager.workspace.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	if state.DefaultBranch != "develop" {
		t.Fatalf("expected github state default branch to refresh to develop: %+v", state)
	}
	record := state.Stories["implement-invoices"]
	if record.DocRef != "develop" || record.DocRefMode != "main" {
		t.Fatalf("expected record to promote links to develop: %+v", record)
	}
	if !strings.Contains(client.issues[101].Body, "/blob/develop/.plan/specs/billing.md") {
		t.Fatalf("expected issue body to point at develop after reconcile:\n%s", client.issues[101].Body)
	}
}

func TestGitHubStoryReadinessDerivesFromDependencies(t *testing.T) {
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
	manager := newGitHubStoryManager(t, root)
	if _, err := manager.EnableGitHubBackend(); err != nil {
		t.Fatal(err)
	}
	seedApprovedGitHubEpic(t, manager)
	if _, err := manager.CreateStory("billing", "Seed billing data", "Seed data first", []string{"Seed data"}, []string{"Run seed checks"}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateStory("billing", "Implement invoices", "Create invoice generation flow", []string{"Generate invoices"}, []string{"Run billing tests"}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateStory("implement-invoices", StoryChanges{SetBlockers: []string{"seed-billing-data"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateStory("billing", "Ship exports", "Parallel story", []string{"Ship exports"}, []string{"Run export tests"}, nil); err != nil {
		t.Fatal(err)
	}

	stories, err := manager.ListStories("", "")
	if err != nil {
		t.Fatal(err)
	}
	bySlug := map[string]StoryInfo{}
	for _, story := range stories {
		bySlug[story.Slug] = story
	}
	if !bySlug["seed-billing-data"].Ready || !bySlug["ship-exports"].Ready {
		t.Fatalf("expected independent stories to be ready: %+v", stories)
	}
	if bySlug["implement-invoices"].Status != "blocked" || !bySlug["implement-invoices"].BlockedByDeps {
		t.Fatalf("expected dependent story to stay blocked: %+v", bySlug["implement-invoices"])
	}

	if _, err := manager.UpdateStory("seed-billing-data", StoryChanges{Status: "done"}); err != nil {
		t.Fatal(err)
	}
	stories, err = manager.ListStories("", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, story := range stories {
		bySlug[story.Slug] = story
	}
	if bySlug["implement-invoices"].Status != "todo" || !bySlug["implement-invoices"].Ready {
		t.Fatalf("expected dependent story to become ready once dependency closes: %+v", bySlug["implement-invoices"])
	}
}

func TestReconcileUpdateVisibleAppliesDerivedLabels(t *testing.T) {
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
	manager := newGitHubStoryManager(t, root)
	if _, err := manager.EnableGitHubBackend(); err != nil {
		t.Fatal(err)
	}
	seedApprovedGitHubEpic(t, manager)
	if _, err := manager.CreateStory("billing", "Seed billing data", "Seed data first", []string{"Seed data"}, []string{"Run seed checks"}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateStory("billing", "Implement invoices", "Create invoice generation flow", []string{"Generate invoices"}, []string{"Run billing tests"}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateStory("implement-invoices", StoryChanges{SetBlockers: []string{"seed-billing-data"}}); err != nil {
		t.Fatal(err)
	}

	if _, err := manager.ReconcileGitHubStories(GitHubReconcileOptions{UpdateVisible: true}); err != nil {
		t.Fatal(err)
	}
	if !containsString(client.issues[101].Labels, planIssueReadyLabel) {
		t.Fatalf("expected ready label on independent issue: %+v", client.issues[101].Labels)
	}
	if !containsString(client.issues[102].Labels, planIssueBlockedLabel) {
		t.Fatalf("expected blocked label on dependent issue: %+v", client.issues[102].Labels)
	}
}

func TestReconcileUpdateVisibleClearsDerivedLabelsAndBecomesNoOp(t *testing.T) {
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
	manager := newGitHubStoryManager(t, root)
	if _, err := manager.EnableGitHubBackend(); err != nil {
		t.Fatal(err)
	}
	seedApprovedGitHubEpic(t, manager)
	if _, err := manager.CreateStory("billing", "Implement invoices", "Create invoice generation flow", []string{"Generate invoices"}, []string{"Run billing tests"}, nil); err != nil {
		t.Fatal(err)
	}

	if _, err := manager.ReconcileGitHubStories(GitHubReconcileOptions{UpdateVisible: true}); err != nil {
		t.Fatal(err)
	}
	if !containsString(client.issues[101].Labels, planIssueReadyLabel) {
		t.Fatalf("expected ready label after first reconcile: %+v", client.issues[101].Labels)
	}

	if _, err := manager.UpdateStory("implement-invoices", StoryChanges{Status: "done"}); err != nil {
		t.Fatal(err)
	}
	if !containsString(client.issues[101].Labels, planIssueReadyLabel) {
		t.Fatalf("expected done issue to still carry stale ready label before cleanup: %+v", client.issues[101].Labels)
	}

	result, err := manager.ReconcileGitHubStories(GitHubReconcileOptions{UpdateVisible: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.UpdatedIssues) != 1 {
		t.Fatalf("expected reconcile to clear stale derived labels once: %+v", result)
	}
	if containsString(client.issues[101].Labels, planIssueReadyLabel) || containsString(client.issues[101].Labels, planIssueBlockedLabel) {
		t.Fatalf("expected derived labels to clear on done issue: %+v", client.issues[101].Labels)
	}

	info, err := manager.workspace.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(info.GitHubFile)
	if err != nil {
		t.Fatal(err)
	}

	result, err = manager.ReconcileGitHubStories(GitHubReconcileOptions{UpdateVisible: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.UpdatedIssues) != 0 {
		t.Fatalf("expected second reconcile to be a no-op after label cleanup: %+v", result)
	}
	after, err := os.ReadFile(info.GitHubFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("expected second reconcile to leave github state unchanged")
	}
}

func newGitHubStoryManager(t *testing.T, root string) *Manager {
	t.Helper()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	return New(ws)
}

func seedApprovedGitHubEpic(t *testing.T, manager *Manager) {
	t.Helper()
	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SetSpecStatus("billing", "approved"); err != nil {
		t.Fatal(err)
	}
}
