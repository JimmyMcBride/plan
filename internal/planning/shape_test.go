package planning

import (
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/workspace"
)

func TestUpdateEpicShapeWritesShapeSectionAndMirrorsSummary(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateEpicShape("billing", EpicShapeInput{
		Appetite:      "One extra shaping pass before spec approval.",
		Outcome:       "Make epics feel like bounded bets.",
		ScopeBoundary: "Capture appetite and out-of-scope decisions for each epic.",
		OutOfScope:    "Do not build tracker sync\nDo not add hosted services",
		SuccessSignal: "Specs inherit clearer boundaries without extra artifact types.",
	}); err != nil {
		t.Fatal(err)
	}

	note, err := notes.Read(filepath.Join(root, ".plan", "epics", "billing.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(note.Content, "## Shape") || !strings.Contains(note.Content, "### Appetite") {
		t.Fatalf("expected shape section:\n%s", note.Content)
	}
	if !strings.Contains(note.Content, "## Outcome\n\nMake epics feel like bounded bets.") {
		t.Fatalf("expected mirrored top-level outcome:\n%s", note.Content)
	}
	if !strings.Contains(note.Content, "Not in scope:\n\n- Do not build tracker sync") {
		t.Fatalf("expected mirrored out-of-scope summary:\n%s", note.Content)
	}
}
