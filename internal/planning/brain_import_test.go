package planning

import (
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/workspace"
)

func TestInspectBrainWorkspaceOnlyReturnsPlanningArtifacts(t *testing.T) {
	preview, err := InspectBrainWorkspace("../../examples/brain")
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Brainstorms) == 0 || len(preview.Epics) == 0 || len(preview.Specs) == 0 || len(preview.Stories) == 0 {
		t.Fatalf("expected planning candidates from brain workspace: %+v", preview)
	}
	for _, item := range preview.Epics {
		if !strings.HasPrefix(item.Path, ".brain/planning/epics/") {
			t.Fatalf("expected epic candidate path under planning epics: %+v", item)
		}
	}
	for _, item := range preview.Specs {
		if !strings.HasPrefix(item.Path, ".brain/planning/specs/") {
			t.Fatalf("expected spec candidate path under planning specs: %+v", item)
		}
	}
	for _, item := range preview.Stories {
		if !strings.HasPrefix(item.Path, ".brain/planning/stories/") {
			t.Fatalf("expected story candidate path under planning stories: %+v", item)
		}
	}
	for _, item := range preview.Brainstorms {
		if !strings.HasPrefix(item.Path, ".brain/brainstorms/") {
			t.Fatalf("expected brainstorm candidate path under brain brainstorms: %+v", item)
		}
	}
}

func TestInspectBrainWorkspaceSupportsDotBrainPath(t *testing.T) {
	preview, err := InspectBrainWorkspace("../../examples/brain/.brain")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(preview.WorkspacePath, "examples/brain") {
		t.Fatalf("expected repo root path in preview: %+v", preview)
	}
}

func TestImportBrainPlanningCreatesCanonicalPlanArtifacts(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	result, err := manager.ImportBrainPlanning(BrainImportSelection{
		WorkspacePath: "../../examples/brain",
		Brainstorms:   []string{"mempalace-inspired-brain-improvements"},
		Epics:         []string{"planning-and-brainstorming-ux"},
		Specs:         []string{"planning-and-brainstorming-ux"},
		Stories:       []string{"add-session-aware-memory-distillation"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Imported) != 4 {
		t.Fatalf("expected four imported planning artifacts: %+v", result)
	}

	for _, item := range []struct {
		path     string
		noteType string
	}{
		{path: filepath.Join(root, ".plan", "brainstorms", "mempalace-inspired-brain-improvements.md"), noteType: "brainstorm"},
		{path: filepath.Join(root, ".plan", "epics", "planning-and-brainstorming-ux.md"), noteType: "epic"},
		{path: filepath.Join(root, ".plan", "specs", "planning-and-brainstorming-ux.md"), noteType: "spec"},
		{path: filepath.Join(root, ".plan", "stories", "add-session-aware-memory-distillation.md"), noteType: "story"},
	} {
		note, err := notes.Read(item.path)
		if err != nil {
			t.Fatalf("expected imported %s note: %v", item.noteType, err)
		}
		if note.Type != item.noteType {
			t.Fatalf("expected imported note type %s, got %s", item.noteType, note.Type)
		}
		if note.Metadata["imported_from"] != "brain" {
			t.Fatalf("expected import provenance metadata: %+v", note.Metadata)
		}
	}
}

func TestImportBrainPlanningAddsVisibleProvenanceAndReviewGuidance(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	if _, err := manager.ImportBrainPlanning(BrainImportSelection{
		WorkspacePath: "../../examples/brain",
		Stories:       []string{"add-session-aware-memory-distillation"},
	}); err != nil {
		t.Fatal(err)
	}

	note, err := notes.Read(filepath.Join(root, ".plan", "stories", "add-session-aware-memory-distillation.md"))
	if err != nil {
		t.Fatal(err)
	}
	if note.Metadata["review_required"] != true {
		t.Fatalf("expected review_required metadata on imported note: %+v", note.Metadata)
	}
	if !strings.Contains(note.Content, "Imported From Brain") {
		t.Fatalf("expected visible provenance link in imported note:\n%s", note.Content)
	}
	if !strings.Contains(note.Content, "Review and normalize this artifact before relying on it for deeper execution work.") {
		t.Fatalf("expected review guidance in imported note:\n%s", note.Content)
	}
}
