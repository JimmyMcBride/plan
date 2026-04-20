package planning

import (
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/workspace"
)

func TestGuidedBrainstormSessionPersistsChainStateAndStaysStable(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)
	if _, err := manager.CreateBrainstorm("Guided Flow"); err != nil {
		t.Fatal(err)
	}

	first, err := manager.EnsureGuidedBrainstormSession("guided-flow")
	if err != nil {
		t.Fatal(err)
	}
	if first.ChainID != "brainstorm/guided-flow" {
		t.Fatalf("unexpected chain id: %+v", first)
	}
	if first.CurrentStage != "brainstorm" || first.CurrentCluster != 1 || first.CurrentClusterLabel != "vision-intake" {
		t.Fatalf("unexpected session progress: %+v", first)
	}
	if first.StageStatuses["brainstorm"] != "in_progress" {
		t.Fatalf("expected brainstorm stage to be in progress: %+v", first)
	}
	createdAt := first.CreatedAt

	updatedNote, second, err := manager.UpdateGuidedBrainstormIntake("guided-flow", GuidedBrainstormIntakeInput{
		Vision:             "Guide the user from a rough idea into a real plan without dumping templates too early.",
		SupportingMaterial: "docs/research.md\nhttps://example.com/guided-planning",
	})
	if err != nil {
		t.Fatal(err)
	}

	if second.CreatedAt != createdAt {
		t.Fatalf("expected session create time to stay stable: first=%s second=%s", createdAt, second.CreatedAt)
	}
	if second.NextAction != "Continue guided brainstorm clarification." {
		t.Fatalf("unexpected next action: %+v", second)
	}
	if !strings.Contains(second.Summary, "Vision captured.") || !strings.Contains(second.Summary, "Supporting material recorded.") {
		t.Fatalf("unexpected session summary: %+v", second)
	}

	state, err := ws.ReadGuidedSessionState()
	if err != nil {
		t.Fatal(err)
	}
	if state.LastActiveChain != "brainstorm/guided-flow" {
		t.Fatalf("unexpected last active chain: %+v", state)
	}
	if len(state.Sessions) != 1 {
		t.Fatalf("expected one guided session record: %+v", state)
	}

	raw, err := notes.Read(filepath.Join(root, ".plan", "brainstorms", "guided-flow.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got := notes.ExtractSection(raw.Content, "Vision"); got != "Guide the user from a rough idea into a real plan without dumping templates too early." {
		t.Fatalf("unexpected brainstorm vision:\n%s", got)
	}
	supporting := notes.ExtractSection(raw.Content, "Supporting Material")
	if !strings.Contains(supporting, "- docs/research.md") || !strings.Contains(supporting, "- https://example.com/guided-planning") {
		t.Fatalf("unexpected supporting material:\n%s", supporting)
	}
	if updatedNote.Path != ".plan/brainstorms/guided-flow.md" {
		t.Fatalf("unexpected updated note path: %+v", updatedNote)
	}
}

func TestSwitchAndReopenGuidedSessionState(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)
	for _, slug := range []string{"alpha", "beta"} {
		if _, err := manager.CreateBrainstorm(slug); err != nil {
			t.Fatal(err)
		}
		if _, _, err := manager.UpdateGuidedBrainstormIntake(slug, GuidedBrainstormIntakeInput{
			Vision:             "Vision for " + slug,
			SupportingMaterial: "docs/" + slug + ".md",
		}); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := manager.SwitchGuidedSession("brainstorm/alpha"); err != nil {
		t.Fatal(err)
	}
	last, err := manager.ReadLastActiveGuidedSession()
	if err != nil {
		t.Fatal(err)
	}
	if last.ChainID != "brainstorm/alpha" {
		t.Fatalf("unexpected last active session: %+v", last)
	}

	if _, err := manager.UpdateGuidedSession("brainstorm/alpha", GuidedSessionUpdateInput{
		CurrentStage: "epic",
		StageStatus:  "done",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateGuidedSession("brainstorm/alpha", GuidedSessionUpdateInput{
		CurrentStage: "spec",
		StageStatus:  "done",
	}); err != nil {
		t.Fatal(err)
	}

	updated, downstream, err := manager.ReopenGuidedSessionStage("brainstorm/alpha", "brainstorm")
	if err != nil {
		t.Fatal(err)
	}
	if len(downstream) != 3 {
		t.Fatalf("unexpected downstream stage count: %+v", downstream)
	}
	if updated.StageStatuses["epic"] != "needs_review" || updated.StageStatuses["spec"] != "needs_review" || updated.StageStatuses["stories"] != "needs_review" {
		t.Fatalf("expected downstream stages to be marked needs_review: %+v", updated)
	}
}
