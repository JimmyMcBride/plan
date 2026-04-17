package cmd

import (
	"bytes"
	"strings"
	"testing"

	"plan/internal/planning"
	"plan/internal/workspace"
)

func TestStoryListSupportsEpicAndStatusFilters(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)

	for _, title := range []string{"Billing", "Exports"} {
		if _, err := manager.CreateEpic(title, ""); err != nil {
			t.Fatal(err)
		}
		if _, err := manager.SetSpecStatus(strings.ToLower(title), "approved"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := manager.CreateStory("billing", "Implement invoices", "Create invoice generation flow", []string{"Generate invoices"}, []string{"Run billing tests"}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateStory("exports", "Ship exports", "Create export flow", []string{"Export invoices"}, []string{"Run export tests"}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateStory("ship-exports", planning.StoryChanges{Status: "in_progress"}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"--project", root, "story", "list", "--epic", "exports", "--status", "in_progress"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "Ship exports [in_progress] epic=exports spec=exports") {
		t.Fatalf("expected exports story in filtered output:\n%s", output)
	}
	if strings.Contains(output, "Implement invoices") {
		t.Fatalf("expected filters to remove billing story:\n%s", output)
	}
}
