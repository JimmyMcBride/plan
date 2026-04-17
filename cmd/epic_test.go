package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/planning"
	"plan/internal/workspace"
)

func TestEpicShapeCommandPersistsClustersAndResumes(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}

	firstInput := strings.Join([]string{
		"One extra shaping pass before spec approval.",
		"",
		"Make epics feel like bounded bets instead of loose containers.",
		"",
		"Capture appetite, scope, and success without inventing new hierarchy.",
		"",
		"Do not rebuild roadmap math.",
		"Do not add coordination dashboards.",
		"",
	}, "\n")
	var first bytes.Buffer
	command := newRootCmd()
	command.SetOut(&first)
	command.SetErr(&first)
	command.SetIn(strings.NewReader(firstInput))
	command.SetArgs([]string{"--project", root, "epic", "shape", "billing"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected first shape pass to succeed: %v\n%s", err, first.String())
	}

	secondInput := strings.Join([]string{
		"Specs inherit a clear appetite and scope boundary for later decomposition.",
		"",
	}, "\n")
	var second bytes.Buffer
	command = newRootCmd()
	command.SetOut(&second)
	command.SetErr(&second)
	command.SetIn(strings.NewReader(secondInput))
	command.SetArgs([]string{"--project", root, "epic", "shape", "billing"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected resumed shape pass to succeed: %v\n%s", err, second.String())
	}

	note, err := notes.Read(filepath.Join(root, ".plan", "epics", "billing.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(note.Content, "## Shape") || !strings.Contains(note.Content, "### Success Signal") {
		t.Fatalf("expected shape section in epic note:\n%s", note.Content)
	}
	if !strings.Contains(note.Content, "## Outcome\n\nMake epics feel like bounded bets instead of loose containers.") {
		t.Fatalf("expected top-level outcome to mirror epic shape:\n%s", note.Content)
	}
	if !strings.Contains(note.Content, "### Out of Scope\n\n- Do not rebuild roadmap math.") {
		t.Fatalf("expected out-of-scope bullets in shape section:\n%s", note.Content)
	}
}
