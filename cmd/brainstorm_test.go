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

func TestBrainstormRefineCommandPersistsClustersAndResumes(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
	if _, err := manager.CreateBrainstorm("Billing Flow"); err != nil {
		t.Fatal(err)
	}

	firstInput := strings.Join([]string{
		"Teams do not have a shaped billing planning flow before they start implementation.",
		"",
		"Agents get a clearer brief before promotion and spec work.",
		"",
		"Keep the tool local-first.",
		"Do not add new top-level artifacts.",
		"",
		"One focused shaping pass before promotion.",
		"",
	}, "\n")
	var first bytes.Buffer
	command := newRootCmd()
	command.SetOut(&first)
	command.SetErr(&first)
	command.SetIn(strings.NewReader(firstInput))
	command.SetArgs([]string{"--project", root, "brainstorm", "refine", "billing-flow"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected first refinement pass to succeed: %v\n%s", err, first.String())
	}

	secondInput := strings.Join([]string{
		"How opinionated should the refine prompts be?",
		"",
		"Add a guided refine command.",
		"Seed shaped output into spec promotion.",
		"",
		"Ship guided refinement before more power features.",
		"",
	}, "\n")
	var second bytes.Buffer
	command = newRootCmd()
	command.SetOut(&second)
	command.SetErr(&second)
	command.SetIn(strings.NewReader(secondInput))
	command.SetArgs([]string{"--project", root, "brainstorm", "refine", "billing-flow"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected resumed refinement pass to succeed: %v\n%s", err, second.String())
	}

	note, err := notes.Read(filepath.Join(root, ".plan", "brainstorms", "billing-flow.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(note.Content, "## Refinement") || !strings.Contains(note.Content, "### Decision Snapshot") {
		t.Fatalf("expected refinement section in note:\n%s", note.Content)
	}
	if !strings.Contains(note.Content, "## Constraints\n\n- Keep the tool local-first.") {
		t.Fatalf("expected constraints to be persisted after cluster save:\n%s", note.Content)
	}
	if !strings.Contains(note.Content, "### Candidate Approaches\n\n- Add a guided refine command.") {
		t.Fatalf("expected candidate approaches after resume:\n%s", note.Content)
	}
}
