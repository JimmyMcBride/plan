package cmd

import (
	"bytes"
	"strings"
	"testing"

	"plan/internal/planning"
	"plan/internal/workspace"
)

func TestStatusCommandPrintsSimpleEpicProgressCounts(t *testing.T) {
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
	if !strings.Contains(output, "stories: 1 total, 0 done, 0 in_progress, 1 blocked") {
		t.Fatalf("expected story summary in output:\n%s", output)
	}
	if strings.Contains(output, "ready_work:") || strings.Contains(output, "versions:") {
		t.Fatalf("expected status output to drop old power summaries:\n%s", output)
	}
	if !strings.Contains(output, "Billing [implementing] (0/1 done, 0 in progress, 1 blocked)") {
		t.Fatalf("expected epic progress counts in output:\n%s", output)
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
	if !strings.Contains(output, "ready_work: 2") {
		t.Fatalf("expected ready-work summary in output:\n%s", output)
	}
	if !strings.Contains(output, "Seed billing data issue=#101 epic=billing spec=billing") || !strings.Contains(output, "Ship exports issue=#102 epic=billing spec=billing") {
		t.Fatalf("expected multiple ready GitHub stories in output:\n%s", output)
	}
}
