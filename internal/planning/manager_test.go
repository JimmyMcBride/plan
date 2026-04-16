package planning

import (
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/workspace"
)

func TestPromoteBrainstormSeedsSpec(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	brainstorm, err := manager.CreateBrainstorm("Auth System")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.AddIdea("auth-system", "Support passwordless sign-in"); err != nil {
		t.Fatal(err)
	}

	bundle, err := manager.PromoteBrainstorm("auth-system")
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Epic.Path != ".plan/epics/auth-system.md" {
		t.Fatalf("unexpected epic path: %s", bundle.Epic.Path)
	}
	if bundle.Spec.Path != ".plan/specs/auth-system.md" {
		t.Fatalf("unexpected spec path: %s", bundle.Spec.Path)
	}
	if got := bundle.Epic.Metadata["source_brainstorm"]; got != brainstorm.Path {
		t.Fatalf("unexpected brainstorm link: %v", got)
	}
	if !strings.Contains(bundle.Spec.Content, "Support passwordless sign-in") {
		t.Fatalf("expected brainstorm idea in spec:\n%s", bundle.Spec.Content)
	}
}

func TestCreateStoryRequiresApprovedSpec(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateStory("billing", "Implement invoices", "", nil, nil, nil); err == nil {
		t.Fatal("expected draft spec to block story creation")
	}
}

func TestCreateStoryAddsSpecReferenceAndCriteria(t *testing.T) {
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
	story, err := manager.CreateStory(
		"billing",
		"Implement invoices",
		"Create invoice generation flow",
		[]string{"Generate invoices from line items"},
		[]string{"Run focused billing tests"},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if story.Path != ".plan/stories/implement-invoices.md" {
		t.Fatalf("unexpected story path: %s", story.Path)
	}

	raw, err := notes.Read(filepath.Join(root, ".plan", "stories", "implement-invoices.md"))
	if err != nil {
		t.Fatal(err)
	}
	if raw.Metadata["epic"] != "billing" || raw.Metadata["spec"] != "billing" {
		t.Fatalf("unexpected story metadata: %+v", raw.Metadata)
	}
	if !strings.Contains(raw.Content, "- [ ] Generate invoices from line items") {
		t.Fatalf("expected criterion in story:\n%s", raw.Content)
	}
	if !strings.Contains(raw.Content, "[Canonical Spec](../specs/billing.md)") {
		t.Fatalf("expected canonical spec link in story:\n%s", raw.Content)
	}
}
