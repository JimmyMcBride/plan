package planning

import (
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/workspace"
)

func TestUpdateStoryCritiqueWritesIdempotentCritiqueSection(t *testing.T) {
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
	if _, err := manager.CreateStory("billing", "Trigger export job", "Create export trigger path", []string{"Users can trigger exports"}, []string{"Run export trigger tests"}, nil); err != nil {
		t.Fatal(err)
	}

	if _, err := manager.UpdateStoryCritique("trigger-export-job", StoryCritiqueInput{
		ScopeFit:              "This story is small enough for one implementation pass.",
		VerticalSliceCheck:    "It includes the UI trigger and the initial backend handoff.",
		HiddenPrerequisites:   "External export flag must exist",
		VerificationGaps:      "No manual verification step is listed",
		RewriteRecommendation: "rewrite",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateStoryCritique("trigger-export-job", StoryCritiqueInput{
		RewriteRecommendation: "rewrite",
	}); err != nil {
		t.Fatal(err)
	}

	note, err := notes.Read(filepath.Join(root, ".plan", "stories", "trigger-export-job.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(note.Content, "## Critique"); got != 1 {
		t.Fatalf("expected one critique section, got %d:\n%s", got, note.Content)
	}
	if !strings.Contains(note.Content, "### Rewrite Recommendation\n\nrewrite") {
		t.Fatalf("expected rewrite recommendation in note:\n%s", note.Content)
	}
}
