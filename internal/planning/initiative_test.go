package planning

import (
	"path/filepath"
	"testing"

	"plan/internal/notes"
	"plan/internal/workspace"
)

func TestListIdeasAndSpecsExposeInitiativeMetadata(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	if _, err := notes.Create(filepath.Join(root, ".plan", "ideas", "guide-packet-foundation.md"), "Guide Packet Foundation", "idea", "# Guide Packet Foundation\n", map[string]any{
		"slug":               "guide-packet-foundation",
		"initiative":         "guide-packet-foundation",
		"initiative_title":   "Guide Packet Foundation",
		"initiative_summary": "Ship the first guide-packet workflow slices.",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SetSpecInitiative("billing", InitiativeRef{
		Slug:    "guide-packet-foundation",
		Title:   "Guide Packet Foundation",
		Summary: "Ship the first guide-packet workflow slices.",
	}); err != nil {
		t.Fatal(err)
	}

	ideas, err := manager.ListIdeas()
	if err != nil {
		t.Fatal(err)
	}
	if len(ideas) != 1 || ideas[0].Initiative != "guide-packet-foundation" || ideas[0].InitiativeTitle != "Guide Packet Foundation" {
		t.Fatalf("expected initiative metadata on idea docs: %+v", ideas)
	}

	specs, err := manager.ListSpecs()
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 1 || specs[0].Initiative != "guide-packet-foundation" || specs[0].InitiativeTitle != "Guide Packet Foundation" {
		t.Fatalf("expected initiative metadata on specs: %+v", specs)
	}
}
