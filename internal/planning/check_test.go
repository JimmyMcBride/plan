package planning

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/workspace"
)

func TestCheckSpecFindsMissingRequiredSections(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}
	body := strings.Join([]string{
		"# Billing Spec",
		"",
		"Created: now",
		"",
		"## Why",
		"",
		"Keep billing reliable.",
		"",
		"## Problem",
		"",
		"## Goals",
		"",
		"## Non-Goals",
		"",
		"## Constraints",
		"",
		"## Verification",
		"",
	}, "\n")
	if _, err := manager.UpdateSpec("billing", notes.UpdateInput{Body: &body}); err != nil {
		t.Fatal(err)
	}

	report, err := manager.Check(CheckInput{SpecSlug: "billing"})
	if err != nil {
		t.Fatal(err)
	}
	if !report.HasErrors() {
		t.Fatalf("expected missing sections to produce blocking findings: %+v", report.Findings)
	}
	assertHasFinding(t, report.Findings, "spec.missing_problem", "Problem")
	assertHasFinding(t, report.Findings, "spec.missing_goals", "Goals")
	assertHasFinding(t, report.Findings, "spec.missing_non_goals", "Non-Goals")
	assertHasFinding(t, report.Findings, "spec.missing_constraints", "Constraints")
	assertHasFinding(t, report.Findings, "spec.missing_verification", "Verification")
}

func TestCheckSpecFlagsThinRequiredSections(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}
	body := strings.Join([]string{
		"# Billing Spec",
		"",
		"Created: now",
		"",
		"## Why",
		"",
		"Keep billing reliable.",
		"",
		"## Problem",
		"",
		"Confusing invoices.",
		"",
		"## Goals",
		"",
		"- clarity",
		"",
		"## Non-Goals",
		"",
		"Not taxes.",
		"",
		"## Constraints",
		"",
		"Local only.",
		"",
		"## Verification",
		"",
		"Run tests.",
		"",
	}, "\n")
	if _, err := manager.UpdateSpec("billing", notes.UpdateInput{Body: &body}); err != nil {
		t.Fatal(err)
	}

	report, err := manager.Check(CheckInput{SpecSlug: "billing"})
	if err != nil {
		t.Fatal(err)
	}
	if report.HasErrors() {
		t.Fatalf("expected thin sections to warn without blocking: %+v", report.Findings)
	}
	assertHasFinding(t, report.Findings, "spec.thin_problem", "Problem")
	assertHasFinding(t, report.Findings, "spec.thin_goals", "Goals")
	assertHasFinding(t, report.Findings, "spec.thin_non_goals", "Non-Goals")
	assertHasFinding(t, report.Findings, "spec.thin_constraints", "Constraints")
	assertHasFinding(t, report.Findings, "spec.thin_verification", "Verification")
}

func TestCheckStoryFindsMissingExecutionSections(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

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
	body := strings.Join([]string{
		"# Implement invoices",
		"",
		"Created: now",
		"",
		"## Description",
		"",
		"## Acceptance Criteria",
		"",
		"## Verification",
		"",
	}, "\n")
	storyPath := filepath.Join(root, ".plan", "stories", "implement-invoices.md")
	if _, err := notes.Update(storyPath, notes.UpdateInput{Body: &body}); err != nil {
		t.Fatal(err)
	}

	report, err := manager.Check(CheckInput{StorySlug: "implement-invoices"})
	if err != nil {
		t.Fatal(err)
	}
	if !report.HasErrors() {
		t.Fatalf("expected missing story sections to produce blocking findings: %+v", report.Findings)
	}
	for _, finding := range report.Findings {
		if finding.ArtifactType != "story" {
			t.Fatalf("expected story scope to only report story findings: %+v", report.Findings)
		}
	}
	assertHasFinding(t, report.Findings, "story.missing_description", "Description")
	assertHasFinding(t, report.Findings, "story.missing_acceptance_criteria", "Acceptance Criteria")
	assertHasFinding(t, report.Findings, "story.missing_verification", "Verification")
}

func TestCheckStoryMatchesLifecycleExpectationsForStartedWork(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

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
	if _, err := manager.UpdateStory("implement-invoices", StoryChanges{Status: "in_progress"}); err != nil {
		t.Fatal(err)
	}
	body := strings.Join([]string{
		"# Implement invoices",
		"",
		"Created: now",
		"",
		"## Description",
		"",
		"Ship invoices.",
		"",
		"## Acceptance Criteria",
		"",
		"- [ ] Generate invoices from line items",
		"",
		"## Verification",
		"",
	}, "\n")
	storyPath := filepath.Join(root, ".plan", "stories", "implement-invoices.md")
	if _, err := notes.Update(storyPath, notes.UpdateInput{Body: &body}); err != nil {
		t.Fatal(err)
	}

	report, err := manager.Check(CheckInput{StorySlug: "implement-invoices"})
	if err != nil {
		t.Fatal(err)
	}
	assertHasFinding(t, report.Findings, "story.missing_verification", "Verification")
	assertHasFinding(t, report.Findings, "story.execution_expectations", "Acceptance Criteria / Verification")
}

func assertHasFinding(t *testing.T, findings []CheckFinding, rule, section string) {
	t.Helper()
	for _, finding := range findings {
		if finding.Rule == rule && finding.Section == section && finding.Suggestion != "" {
			return
		}
	}
	t.Fatalf("expected finding %s for %s: %+v", rule, section, findings)
}

func TestCheckStorySupportsGitHubBackedStories(t *testing.T) {
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
	if _, err := manager.CreateStory("billing", "Implement invoices", "Create invoice generation flow", []string{"Generate invoices from line items"}, []string{"Run focused billing tests"}, nil); err != nil {
		t.Fatal(err)
	}

	report, err := manager.Check(CheckInput{StorySlug: "implement-invoices"})
	if err != nil {
		t.Fatal(err)
	}
	if report.HasErrors() {
		t.Fatalf("expected GitHub-backed story check to pass: %+v", report.Findings)
	}
}

func TestCheckFlagsUntrackedPlanLabeledGitHubIssue(t *testing.T) {
	client := checkDriftClient(map[int]*GitHubIssue{
		301: {Number: 301, URL: "https://github.com/JimmyMcBride/plan/issues/301", Title: "Manual Spec", State: "open", Labels: []string{planIssueSpecLabel}},
	})
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	t.Cleanup(reset)
	manager := setupGitHubSourceModeCheck(t)

	report, err := manager.Check(CheckInput{})
	if err != nil {
		t.Fatal(err)
	}
	assertHasFinding(t, report.Findings, "github_planning.untracked_issue", "GitHub Planning")
}

func TestCheckFlagsMultiSpecInitiativeWithoutMilestone(t *testing.T) {
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return checkDriftClient(nil) })
	t.Cleanup(reset)
	manager := setupGitHubSourceModeCheck(t)
	state, err := manager.workspace.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	state.Planning["initiative"] = workspace.GitHubPlanningRecord{Slug: "initiative", Kind: "initiative", Title: "Initiative", IssueNumber: 401, IssueURL: "https://github.com/JimmyMcBride/plan/issues/401"}
	state.Planning["spec-a"] = workspace.GitHubPlanningRecord{Slug: "spec-a", Kind: "spec", Title: "Spec A", IssueNumber: 402, ParentIssueNumber: 401}
	state.Planning["spec-b"] = workspace.GitHubPlanningRecord{Slug: "spec-b", Kind: "spec", Title: "Spec B", IssueNumber: 403, ParentIssueNumber: 401}
	if err := manager.workspace.WriteGitHubState(*state); err != nil {
		t.Fatal(err)
	}

	report, err := manager.Check(CheckInput{})
	if err != nil {
		t.Fatal(err)
	}
	assertHasFinding(t, report.Findings, "github_planning.missing_multi_spec_milestone", "GitHub Planning")
}

func TestCheckFlagsMissingProjectDecisionForFiveSpecMilestone(t *testing.T) {
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return checkDriftClient(nil) })
	t.Cleanup(reset)
	manager := setupGitHubSourceModeCheck(t)
	state, err := manager.workspace.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		slug := fmt.Sprintf("spec-%d", i)
		state.Planning[slug] = workspace.GitHubPlanningRecord{
			Slug:            slug,
			Kind:            "spec",
			Title:           fmt.Sprintf("Spec %d", i),
			IssueNumber:     500 + i,
			MilestoneNumber: 7,
			MilestoneTitle:  "Readiness",
		}
	}
	if err := manager.workspace.WriteGitHubState(*state); err != nil {
		t.Fatal(err)
	}

	report, err := manager.Check(CheckInput{})
	if err != nil {
		t.Fatal(err)
	}
	assertHasFinding(t, report.Findings, "github_planning.missing_project_decision", "GitHub Planning")
}

func TestCheckFlagsLabelUsedWhereMilestoneExpected(t *testing.T) {
	client := checkDriftClient(map[int]*GitHubIssue{
		601: {Number: 601, URL: "https://github.com/JimmyMcBride/plan/issues/601", Title: "Spec", State: "open", Labels: []string{planIssueSpecLabel, "Readiness Initiative"}},
	})
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	t.Cleanup(reset)
	manager := setupGitHubSourceModeCheck(t)
	state, err := manager.workspace.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	state.Planning["readiness-initiative"] = workspace.GitHubPlanningRecord{Slug: "readiness-initiative", Kind: "initiative", Title: "Readiness Initiative", IssueNumber: 600}
	state.Planning["spec"] = workspace.GitHubPlanningRecord{Slug: "spec", Kind: "spec", Title: "Spec", IssueNumber: 601, ParentIssueNumber: 600}
	if err := manager.workspace.WriteGitHubState(*state); err != nil {
		t.Fatal(err)
	}

	report, err := manager.Check(CheckInput{})
	if err != nil {
		t.Fatal(err)
	}
	assertHasFinding(t, report.Findings, "github_planning.label_used_as_milestone", "GitHub Planning")
}

func checkDriftClient(issues map[int]*GitHubIssue) *stubGitHubClient {
	return &stubGitHubClient{
		preflight: &GitHubRepoInfo{Repo: "JimmyMcBride/plan", RepoURL: "https://github.com/JimmyMcBride/plan", DefaultBranch: "develop"},
		context: &GitHubContext{
			Repo:          GitHubRepoInfo{Repo: "JimmyMcBride/plan", RepoURL: "https://github.com/JimmyMcBride/plan", DefaultBranch: "develop"},
			CurrentBranch: "develop",
			CurrentSHA:    "abc123",
		},
		issues: issues,
	}
}

func setupGitHubSourceModeCheck(t *testing.T) *Manager {
	t.Helper()
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)
	if _, err := manager.SetSourceMode(workspace.SourceOfTruthGitHub); err != nil {
		t.Fatal(err)
	}
	return manager
}
