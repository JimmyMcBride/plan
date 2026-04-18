package planning

import (
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/workspace"
)

func TestPreviewStorySlicesDerivesCandidatesFromStoryBreakdown(t *testing.T) {
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
		"Billing work needs a stronger handoff.",
		"",
		"## Problem",
		"",
		"The spec-to-story handoff is weak.",
		"",
		"## Goals",
		"",
		"- create stories from the spec",
		"",
		"## Non-Goals",
		"",
		"- rebuild the workflow engine",
		"",
		"## Constraints",
		"",
		"- keep it local-first",
		"",
		"## Solution Shape",
		"",
		"Slice approved specs into first-pass stories.",
		"",
		"## Flows",
		"",
		"1. Approve spec.",
		"2. Preview slices.",
		"",
		"## Data / Interfaces",
		"",
		"- story slice candidate model",
		"",
		"## Risks / Open Questions",
		"",
		"- duplicate story slugs",
		"",
		"## Rollout",
		"",
		"- ship preview before apply",
		"",
		"## Verification",
		"",
		"- run story slice tests",
		"",
		"## Story Breakdown",
		"",
		"- Trigger export job",
		"  - desc: Create the trigger path from the billing UI.",
		"  - accept: Users can trigger a billing export.",
		"  - verify: Run the billing export trigger tests.",
		"- Deliver export payload",
		"",
	}, "\n")
	if _, err := manager.UpdateSpec("billing", notes.UpdateInput{
		Body: &body,
		Metadata: map[string]any{
			"status": "approved",
		},
	}); err != nil {
		t.Fatal(err)
	}

	preview, err := manager.PreviewStorySlices("billing")
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(preview.Candidates))
	}
	if preview.Candidates[0].Description != "Create the trigger path from the billing UI." {
		t.Fatalf("unexpected custom description: %+v", preview.Candidates[0])
	}
	if len(preview.Candidates[1].AcceptanceCriteria) != 1 || preview.Candidates[1].AcceptanceCriteria[0] != "Deliver export payload" {
		t.Fatalf("expected default acceptance criteria from title: %+v", preview.Candidates[1])
	}
	if len(preview.Candidates[1].Verification) != 1 || preview.Candidates[1].Verification[0] != "run story slice tests" {
		t.Fatalf("expected verification defaults from spec: %+v", preview.Candidates[1])
	}
}

func TestApplyStorySlicesCreatesStoriesAndRefreshesStoryBreakdown(t *testing.T) {
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
		"Billing export needs slicing.",
		"",
		"## Problem",
		"",
		"Approved specs still require manual story creation.",
		"",
		"## Goals",
		"",
		"- create stories from the spec",
		"",
		"## Non-Goals",
		"",
		"- tracker sync",
		"",
		"## Constraints",
		"",
		"- keep it local-first",
		"",
		"## Solution Shape",
		"",
		"Slice stories from the approved spec.",
		"",
		"## Flows",
		"",
		"1. Approve spec.",
		"2. Apply slices.",
		"",
		"## Data / Interfaces",
		"",
		"- story slice candidates",
		"",
		"## Risks / Open Questions",
		"",
		"- slug collisions",
		"",
		"## Rollout",
		"",
		"- dogfood it locally",
		"",
		"## Verification",
		"",
		"- run story slice tests",
		"",
		"## Story Breakdown",
		"",
		"- Trigger export job",
		"- Deliver export payload",
		"",
	}, "\n")
	if _, err := manager.UpdateSpec("billing", notes.UpdateInput{
		Body: &body,
		Metadata: map[string]any{
			"status": "approved",
		},
	}); err != nil {
		t.Fatal(err)
	}

	result, err := manager.ApplyStorySlices("billing")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.CreatedPaths) != 2 {
		t.Fatalf("expected 2 created stories, got %+v", result)
	}

	spec, err := notes.Read(filepath.Join(root, ".plan", "specs", "billing.md"))
	if err != nil {
		t.Fatal(err)
	}
	breakdown := notes.ExtractSection(spec.Content, "Story Breakdown")
	if !strings.Contains(breakdown, "[Trigger export job](../stories/trigger-export-job.md)") {
		t.Fatalf("expected linked story breakdown:\n%s", breakdown)
	}

	story, err := notes.Read(filepath.Join(root, ".plan", "stories", "trigger-export-job.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(story.Content, "## Acceptance Criteria") || !strings.Contains(story.Content, "## Verification") {
		t.Fatalf("expected execution-ready story:\n%s", story.Content)
	}

	rerun, err := manager.ApplyStorySlices("billing")
	if err != nil {
		t.Fatal(err)
	}
	if len(rerun.CreatedPaths) != 0 || len(rerun.SkippedPaths) != 2 {
		t.Fatalf("expected rerun to reuse existing stories: %+v", rerun)
	}
}
