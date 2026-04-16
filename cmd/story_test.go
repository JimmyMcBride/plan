package cmd

import (
	"bytes"
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/planning"
	"plan/internal/workspace"
)

func TestStoryListSupportsVersionFilter(t *testing.T) {
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
	}
	if _, err := manager.UpdateSpec("billing", notes.UpdateInput{
		Metadata: map[string]any{"status": "approved", "target_version": "v2"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateSpec("exports", notes.UpdateInput{
		Metadata: map[string]any{"status": "approved", "target_version": "v3"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateStory("billing", "Implement invoices", "Create invoice generation flow", []string{"Generate invoices"}, []string{"Run billing tests"}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateStory("exports", "Ship exports", "Create export flow", []string{"Export invoices"}, []string{"Run export tests"}, nil); err != nil {
		t.Fatal(err)
	}

	prevProjectDir := projectDir
	projectDir = root
	defer func() { projectDir = prevProjectDir }()

	var buf bytes.Buffer
	command := newStoryCommand()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"list", "--version", "v2"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "Implement invoices [todo] epic=billing spec=billing") {
		t.Fatalf("expected billing story in filtered output:\n%s", output)
	}
	if strings.Contains(output, "Ship exports") {
		t.Fatalf("expected version filter to remove exports story:\n%s", output)
	}
}
