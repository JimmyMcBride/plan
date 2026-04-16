package cmd

import (
	"bytes"
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/planning"
	"plan/internal/workspace"
)

func TestStatusCommandPrintsEpicProgressCounts(t *testing.T) {
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
	if _, err := manager.CreateStory(
		"billing",
		"Implement invoices",
		"Create invoice generation flow",
		[]string{"Generate invoices from line items"},
		[]string{"Run focused billing tests"},
		nil,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateStory("implement-invoices", planning.StoryChanges{Status: "blocked"}); err != nil {
		t.Fatal(err)
	}

	prevProjectDir := projectDir
	projectDir = root
	defer func() { projectDir = prevProjectDir }()

	var buf bytes.Buffer
	command := newStatusCommand()
	command.SetOut(&buf)
	command.SetErr(&buf)
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "stories: 1 total, 0 done, 0 in_progress, 1 blocked") {
		t.Fatalf("expected story summary in output:\n%s", output)
	}
	if !strings.Contains(output, "ready_work: 0 ready, 1 blocked_by_dependencies") {
		t.Fatalf("expected ready-work summary in output:\n%s", output)
	}
	if !strings.Contains(output, "Billing [implementing] (0/1 done, 0 in progress, 1 blocked)") {
		t.Fatalf("expected epic progress counts in output:\n%s", output)
	}
}

func TestStatusCommandPrintsRoadmapVersionSummaries(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)

	if _, err := manager.CreateEpic("Roadmap Work", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateSpec("roadmap-work", notes.UpdateInput{
		Metadata: map[string]any{"status": "approved", "target_version": "v2"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateStory(
		"roadmap-work",
		"Ship roadmap parser",
		"Parse version sections from roadmap",
		[]string{"Version sections are parsed"},
		[]string{"go test ./internal/planning"},
		nil,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateStory("ship-roadmap-parser", planning.StoryChanges{Status: "in_progress"}); err != nil {
		t.Fatal(err)
	}

	prevProjectDir := projectDir
	projectDir = root
	defer func() { projectDir = prevProjectDir }()

	var buf bytes.Buffer
	command := newStatusCommand()
	command.SetOut(&buf)
	command.SetErr(&buf)
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "roadmap: .plan/ROADMAP.md") {
		t.Fatalf("expected roadmap path in output:\n%s", output)
	}
	if !strings.Contains(output, "ready_work: 0 ready, 0 blocked_by_dependencies") {
		t.Fatalf("expected ready-work summary in output:\n%s", output)
	}
	if !strings.Contains(output, "v2: Rigor (1 stories, 0 done, 1 in_progress, 0 blocked)") {
		t.Fatalf("expected version summary in output:\n%s", output)
	}
	if !strings.Contains(output, "Roadmap Work [implementing] (0/1 done, 1 in progress, 0 blocked)") {
		t.Fatalf("expected roadmap epic progress in output:\n%s", output)
	}
}
