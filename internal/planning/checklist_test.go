package planning

import (
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/workspace"
)

func TestRunSpecChecklistPreservesMultipleProfiles(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}
	body := strings.Join([]string{
		"# Billing Spec",
		"",
		"Created: now",
		"",
		"## Why",
		"",
		"Billing needs a safer integration path.",
		"",
		"## Problem",
		"",
		"Billing integration work is underspecified.",
		"",
		"## Goals",
		"",
		"- ship the safer path",
		"",
		"## Non-Goals",
		"",
		"- redesign all billing UX",
		"",
		"## Constraints",
		"",
		"- stay local-first",
		"",
		"## Solution Shape",
		"",
		"Add an API export and a schema migration.",
		"",
		"## Flows",
		"",
		"1. Open the billing export screen.",
		"2. Submit the export form.",
		"",
		"## Data / Interfaces",
		"",
		"- external API contract",
		"- schema migration for export state",
		"",
		"## Risks / Open Questions",
		"",
		"- auth timeout handling",
		"- migration rollout",
		"",
		"## Rollout",
		"",
		"- ship behind a flag with rollback and monitoring",
		"",
		"## Verification",
		"",
		"- manual UI verification",
		"- validate migration data integrity",
		"",
	}, "\n")
	if _, err := manager.UpdateSpec("billing", notes.UpdateInput{Body: &body}); err != nil {
		t.Fatal(err)
	}

	if _, err := manager.RunSpecChecklist("billing", checklistProfileUIFlow); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.RunSpecChecklist("billing", checklistProfileAPIIntegration); err != nil {
		t.Fatal(err)
	}

	note, err := notes.Read(filepath.Join(root, ".plan", "specs", "billing.md"))
	if err != nil {
		t.Fatal(err)
	}
	checklist := notes.ExtractSection(note.Content, "Checklist")
	if !strings.Contains(checklist, "### ui-flow") || !strings.Contains(checklist, "### api-integration") {
		t.Fatalf("expected multiple checklist profiles to persist:\n%s", checklist)
	}
}
