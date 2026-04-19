package planning

import (
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
