package planning

import (
	"strings"
	"testing"
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
