package cmd

import (
	"bytes"
	"strings"
	"testing"

	"plan/internal/planning"
	"plan/internal/workspace"
)

func TestReadyCommandPrintsReadyAndBlockedStories(t *testing.T) {
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
	for _, input := range []struct {
		title string
		body  string
	}{
		{title: "Implement invoices", body: "Create invoice generation flow"},
		{title: "Ship exports", body: "Create export flow"},
		{title: "Manual blocker", body: "Needs external review"},
	} {
		if _, err := manager.CreateStory(
			"billing",
			input.title,
			input.body,
			[]string{"Acceptance for " + input.title},
			[]string{"Verify " + input.title},
			nil,
		); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := manager.UpdateStory("ship-exports", planning.StoryChanges{SetBlockers: []string{"implement-invoices"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateStory("manual-blocker", planning.StoryChanges{Status: "blocked"}); err != nil {
		t.Fatal(err)
	}

	prevProjectDir := projectDir
	projectDir = root
	defer func() { projectDir = prevProjectDir }()

	var buf bytes.Buffer
	command := newReadyCommand()
	command.SetOut(&buf)
	command.SetErr(&buf)
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "ready: 1") || !strings.Contains(output, "Implement invoices [todo] epic=billing") {
		t.Fatalf("expected ready story in output:\n%s", output)
	}
	if !strings.Contains(output, "blocked: 2") {
		t.Fatalf("expected blocked summary in output:\n%s", output)
	}
	if !strings.Contains(output, "blocked by implement-invoices [todo]") {
		t.Fatalf("expected dependency blocker reason in output:\n%s", output)
	}
	if !strings.Contains(output, "story status is blocked") {
		t.Fatalf("expected manual blocked reason in output:\n%s", output)
	}
}
