package planning

import (
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/workspace"
)

func TestUpdateBrainstormChallengeWritesIdempotentChallengeSection(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	if _, err := manager.CreateBrainstorm("Billing"); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormChallenge("billing", BrainstormChallengeInput{
		RabbitHoles:           "Overbuilding the prompt loop",
		NoGos:                 "Do not add cloud sync",
		Assumptions:           "Users will tolerate one extra pass",
		LikelyOverengineering: "Packing too many challenge fields into one command",
		SimplerAlternative:    "Start with one adversarial pass",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormChallenge("billing", BrainstormChallengeInput{
		SimplerAlternative: "Start with one adversarial pass",
	}); err != nil {
		t.Fatal(err)
	}

	note, err := notes.Read(filepath.Join(root, ".plan", "brainstorms", "billing.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(note.Content, "## Challenge"); got != 1 {
		t.Fatalf("expected one challenge section, got %d:\n%s", got, note.Content)
	}
	if !strings.Contains(note.Content, "### No-Gos\n\n- Do not add cloud sync") {
		t.Fatalf("expected challenge content to persist:\n%s", note.Content)
	}
}
