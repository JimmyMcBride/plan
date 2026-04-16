package planning

import (
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/workspace"
)

func TestPromoteBrainstormSeedsEpicAndSpecWithProvenance(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	brainstorm, err := manager.CreateBrainstormWithInput(BrainstormCreateInput{
		Topic:         "Auth System",
		FocusQuestion: "How do we make sign-in simpler without lowering trust?",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.AddBrainstormEntry("auth-system", "desired-outcome", "Deliver an auth flow that feels low-friction for small teams."); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.AddBrainstormEntry("auth-system", "constraints", "Keep setup local-first and avoid hosted auth dependencies."); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.AddBrainstormEntry("auth-system", "ideas", "Support passwordless sign-in"); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.AddBrainstormEntry("auth-system", "open-questions", "How should account recovery work?"); err != nil {
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
		t.Fatalf("unexpected epic brainstorm link: %v", got)
	}
	if got := bundle.Spec.Metadata["source_brainstorm"]; got != brainstorm.Path {
		t.Fatalf("unexpected spec brainstorm link: %v", got)
	}
	if !strings.Contains(bundle.Epic.Content, "Deliver an auth flow that feels low-friction for small teams.") {
		t.Fatalf("expected brainstorm outcome in epic:\n%s", bundle.Epic.Content)
	}
	if !strings.Contains(bundle.Epic.Content, "[Source Brainstorm](../brainstorms/auth-system.md)") {
		t.Fatalf("expected brainstorm link in epic resources:\n%s", bundle.Epic.Content)
	}
	if !strings.Contains(bundle.Spec.Content, "How do we make sign-in simpler without lowering trust?") {
		t.Fatalf("expected brainstorm focus question in spec problem:\n%s", bundle.Spec.Content)
	}
	if !strings.Contains(bundle.Spec.Content, "Support passwordless sign-in") {
		t.Fatalf("expected brainstorm idea in spec:\n%s", bundle.Spec.Content)
	}
	if !strings.Contains(bundle.Spec.Content, "Keep setup local-first and avoid hosted auth dependencies.") {
		t.Fatalf("expected brainstorm constraints in spec:\n%s", bundle.Spec.Content)
	}
	if !strings.Contains(bundle.Spec.Content, "How should account recovery work?") {
		t.Fatalf("expected brainstorm questions in spec:\n%s", bundle.Spec.Content)
	}
	if strings.Contains(bundle.Spec.Content, "**") {
		t.Fatalf("expected seeded spec content to avoid brainstorm timestamp noise:\n%s", bundle.Spec.Content)
	}

	promotedBrainstorm, err := notes.Read(filepath.Join(root, ".plan", "brainstorms", "auth-system.md"))
	if err != nil {
		t.Fatal(err)
	}
	if promotedBrainstorm.Metadata["status"] != "promoted" {
		t.Fatalf("expected brainstorm to be marked promoted: %+v", promotedBrainstorm.Metadata)
	}
	if promotedBrainstorm.Metadata["epic"] != "auth-system" || promotedBrainstorm.Metadata["spec"] != "auth-system" {
		t.Fatalf("expected brainstorm promotion links: %+v", promotedBrainstorm.Metadata)
	}
}

func TestCreateBrainstormWithInputSeedsFocusAndIdeas(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	note, err := manager.CreateBrainstormWithInput(BrainstormCreateInput{
		Topic:         "Release System",
		FocusQuestion: "What keeps releases safe and boring?",
		Ideas: []string{
			"Add dry-run release validation",
			"Publish checksums with each build",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	raw, err := notes.Read(filepath.Join(root, ".plan", "brainstorms", "release-system.md"))
	if err != nil {
		t.Fatal(err)
	}
	if note.Path != ".plan/brainstorms/release-system.md" {
		t.Fatalf("unexpected brainstorm path: %s", note.Path)
	}
	if got := notes.ExtractSection(raw.Content, "Focus Question"); got != "What keeps releases safe and boring?" {
		t.Fatalf("unexpected focus question:\n%s", got)
	}
	ideas := notes.ExtractSection(raw.Content, "Ideas")
	if !strings.Contains(ideas, "- Add dry-run release validation") || !strings.Contains(ideas, "- Publish checksums with each build") {
		t.Fatalf("expected seeded brainstorm ideas:\n%s", ideas)
	}
	if strings.Contains(ideas, "**") {
		t.Fatalf("expected brainstorm ideas without timestamp formatting:\n%s", ideas)
	}
}

func TestAddBrainstormEntrySupportsSectionsAndMultilineBullets(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	if _, err := manager.CreateBrainstorm("Agent Workflow"); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.AddIdea("agent-workflow", "Capture dependencies\nTrack verification upfront"); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.AddBrainstormEntry("agent-workflow", "open-questions", "How strict should approval be?\nShould specs own rollout notes?"); err != nil {
		t.Fatal(err)
	}

	raw, err := notes.Read(filepath.Join(root, ".plan", "brainstorms", "agent-workflow.md"))
	if err != nil {
		t.Fatal(err)
	}
	ideas := notes.ExtractSection(raw.Content, "Ideas")
	if !strings.Contains(ideas, "- Capture dependencies") || !strings.Contains(ideas, "- Track verification upfront") {
		t.Fatalf("expected multiline idea capture to produce bullets:\n%s", ideas)
	}
	questions := notes.ExtractSection(raw.Content, "Open Questions")
	if !strings.Contains(questions, "- How strict should approval be?") || !strings.Contains(questions, "- Should specs own rollout notes?") {
		t.Fatalf("expected open questions bullets:\n%s", questions)
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
