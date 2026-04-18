package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/planning"
	"plan/internal/workspace"
)

func TestStoryListSupportsEpicAndStatusFilters(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)

	for _, title := range []string{"Billing", "Exports"} {
		if _, err := manager.CreateEpic(title, ""); err != nil {
			t.Fatal(err)
		}
		if _, err := manager.SetSpecStatus(strings.ToLower(title), "approved"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := manager.CreateStory("billing", "Implement invoices", "Create invoice generation flow", []string{"Generate invoices"}, []string{"Run billing tests"}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateStory("exports", "Ship exports", "Create export flow", []string{"Export invoices"}, []string{"Run export tests"}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateStory("ship-exports", planning.StoryChanges{Status: "in_progress"}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"--project", root, "story", "list", "--epic", "exports", "--status", "in_progress"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "Ship exports [in_progress] epic=exports spec=exports") {
		t.Fatalf("expected exports story in filtered output:\n%s", output)
	}
	if strings.Contains(output, "Implement invoices") {
		t.Fatalf("expected filters to remove billing story:\n%s", output)
	}
}

func TestStorySlicePreviewAndApplyFlow(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
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
		"Billing handoff needs stronger slicing.",
		"",
		"## Problem",
		"",
		"Stories are still created manually.",
		"",
		"## Goals",
		"",
		"- create first-pass stories",
		"",
		"## Non-Goals",
		"",
		"- tracker integration",
		"",
		"## Constraints",
		"",
		"- keep it local-first",
		"",
		"## Solution Shape",
		"",
		"Use Story Breakdown to seed slice candidates.",
		"",
		"## Flows",
		"",
		"1. Approve spec.",
		"2. Slice stories.",
		"",
		"## Data / Interfaces",
		"",
		"- slice candidate model",
		"",
		"## Risks / Open Questions",
		"",
		"- duplicate slugs",
		"",
		"## Rollout",
		"",
		"- dogfood locally",
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

	var preview bytes.Buffer
	command := newRootCmd()
	command.SetOut(&preview)
	command.SetErr(&preview)
	command.SetArgs([]string{"--project", root, "story", "slice", "billing"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected preview to succeed: %v\n%s", err, preview.String())
	}
	if !strings.Contains(preview.String(), "story_slice_preview: .plan/specs/billing.md") {
		t.Fatalf("expected preview header:\n%s", preview.String())
	}

	var apply bytes.Buffer
	command = newRootCmd()
	command.SetOut(&apply)
	command.SetErr(&apply)
	command.SetIn(strings.NewReader("y\n"))
	command.SetArgs([]string{"--project", root, "story", "slice", "billing", "--apply"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected apply to succeed: %v\n%s", err, apply.String())
	}
	if !strings.Contains(apply.String(), "created: 2") {
		t.Fatalf("expected created stories in output:\n%s", apply.String())
	}
	if _, err := notes.Read(filepath.Join(root, ".plan", "stories", "trigger-export-job.md")); err != nil {
		t.Fatalf("expected sliced story note: %v", err)
	}
}

func TestStoryCritiqueCommandWritesCritiqueAndFailsOnRewrite(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SetSpecStatus("billing", "approved"); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateStory("billing", "Trigger export job", "Create export trigger path", []string{"Users can trigger exports"}, []string{"Run export trigger tests"}, nil); err != nil {
		t.Fatal(err)
	}

	input := strings.Join([]string{
		"The story is close to the right size.",
		"",
		"It spans one vertical slice.",
		"",
		"Feature flag must already exist.",
		"",
		"No manual verification step is listed.",
		"",
		"rewrite",
		"",
	}, "\n")
	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetIn(strings.NewReader(input))
	command.SetArgs([]string{"--project", root, "story", "critique", "trigger-export-job"})
	err := command.Execute()
	if err == nil {
		t.Fatalf("expected rewrite recommendation to fail command:\n%s", buf.String())
	}
	if !strings.Contains(err.Error(), "story critique recommends rewrite") {
		t.Fatalf("unexpected critique error: %v", err)
	}

	note, err := notes.Read(filepath.Join(root, ".plan", "stories", "trigger-export-job.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(note.Content, "## Critique") || !strings.Contains(note.Content, "### Rewrite Recommendation\n\nrewrite") {
		t.Fatalf("expected critique section in story note:\n%s", note.Content)
	}
}
