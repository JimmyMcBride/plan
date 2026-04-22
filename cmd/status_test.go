package cmd

import (
	"bytes"
	"strings"
	"testing"

	"plan/internal/planning"
	"plan/internal/workspace"
)

func TestStatusCommandPrintsSpecQueueSummary(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)

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
	if _, err := manager.UpdateStory("implement-invoices", planning.StoryChanges{Status: "blocked"}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"--project", root, "status"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "specs: 1 total, 0 draft, 0 approved, 1 implementing, 0 done") {
		t.Fatalf("expected spec summary in output:\n%s", output)
	}
	if strings.Contains(output, "ready_work:") || strings.Contains(output, "versions:") {
		t.Fatalf("expected status output to stay on the spec-first queue surface:\n%s", output)
	}
	if !strings.Contains(output, "legacy_stories: 1 total, 0 done, 0 in_progress, 1 blocked") {
		t.Fatalf("expected legacy story summary in output:\n%s", output)
	}
	if !strings.Contains(output, "legacy_epics:\n  - Billing [implementing] (0/1 done, 0 in progress, 1 blocked)") {
		t.Fatalf("expected legacy epic summary in output:\n%s", output)
	}
}

func TestStatusCommandShowsMultipleReadyGitHubStories(t *testing.T) {
	client := &stubGitHubStoryClient{
		preflight: &planning.GitHubRepoInfo{
			Repo:          "JimmyMcBride/plan",
			RepoURL:       "https://github.com/JimmyMcBride/plan",
			DefaultBranch: "main",
		},
		context: &planning.GitHubContext{
			Repo: planning.GitHubRepoInfo{
				Repo:          "JimmyMcBride/plan",
				RepoURL:       "https://github.com/JimmyMcBride/plan",
				DefaultBranch: "main",
			},
			CurrentBranch: "main",
			CurrentSHA:    "abc123def456",
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
	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SetSpecStatus("billing", "approved"); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.EnableGitHubBackend(); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateStory("billing", "Seed billing data", "Seed data first", []string{"Seed data"}, []string{"Run seed checks"}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateStory("billing", "Ship exports", "Parallel story", []string{"Ship exports"}, []string{"Run export tests"}, nil); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"--project", root, "status"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "ready_specs: 1") {
		t.Fatalf("expected ready spec summary in output:\n%s", output)
	}
	if !strings.Contains(output, "Billing Spec status=approved") {
		t.Fatalf("expected approved spec in ready queue output:\n%s", output)
	}
	if !strings.Contains(output, "legacy_stories: 2 total, 0 done, 0 in_progress, 0 blocked") {
		t.Fatalf("expected legacy story summary in output:\n%s", output)
	}
}

func TestStatusCommandIgnoresArchivedLegacyHierarchy(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)

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
	if _, err := ws.UpdateWithOptions(workspace.UpdateOptions{ArchiveLegacy: true}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"--project", root, "status"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "ready_specs: 1") {
		t.Fatalf("expected ready spec summary in output:\n%s", output)
	}
	if strings.Contains(output, "legacy_stories:") || strings.Contains(output, "legacy_epics:") {
		t.Fatalf("expected archived legacy hierarchy to stay out of active status output:\n%s", output)
	}
}
