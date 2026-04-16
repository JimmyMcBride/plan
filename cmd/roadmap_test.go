package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/workspace"
)

func TestRoadmapVersionsCommandPrintsParsedVersions(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}

	body := `# Roadmap: plan

## Overview

Overview text.

## v1: Local-First Core

Goal: Ship the trustworthy foundation.

- [ ] Core Workspace and Artifact System
- [x] Spec-Driven Planning Workflow

Summary:
- establish .plan as the workspace
- make the core loop work
`
	if err := os.WriteFile(filepath.Join(root, ".plan", "ROADMAP.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	prevProjectDir := projectDir
	projectDir = root
	defer func() { projectDir = prevProjectDir }()

	var buf bytes.Buffer
	command := newRoadmapCommand()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"versions"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, ".plan/ROADMAP.md") {
		t.Fatalf("expected roadmap path in output:\n%s", output)
	}
	if !strings.Contains(output, "v1: Local-First Core") {
		t.Fatalf("expected version heading in output:\n%s", output)
	}
	if !strings.Contains(output, "goal: Ship the trustworthy foundation.") {
		t.Fatalf("expected goal in output:\n%s", output)
	}
	if !strings.Contains(output, "  - [ ] Core Workspace and Artifact System") || !strings.Contains(output, "  - [x] Spec-Driven Planning Workflow") {
		t.Fatalf("expected roadmap epic list in output:\n%s", output)
	}
	if !strings.Contains(output, "summary:") || !strings.Contains(output, "  - establish .plan as the workspace") {
		t.Fatalf("expected roadmap summary in output:\n%s", output)
	}
}

func TestRoadmapVersionsCommandFiltersVersion(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}

	body := `# Roadmap: plan

## v1: Core

Goal: Ship core.

## v2: Rigor

Goal: Tighten planning.
`
	if err := os.WriteFile(filepath.Join(root, ".plan", "ROADMAP.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	prevProjectDir := projectDir
	projectDir = root
	defer func() { projectDir = prevProjectDir }()

	var buf bytes.Buffer
	command := newRoadmapCommand()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"versions", "--version", "v2"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if strings.Contains(output, "v1: Core") {
		t.Fatalf("expected version filter to remove v1 output:\n%s", output)
	}
	if !strings.Contains(output, "v2: Rigor") {
		t.Fatalf("expected version filter to keep v2 output:\n%s", output)
	}
}
