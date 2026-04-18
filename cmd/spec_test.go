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

func TestSpecAnalyzeCommandWritesAnalysisAndFailsOnBlockingFindings(t *testing.T) {
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
		"The API should feel simpler for billing setup.",
		"",
		"## Problem",
		"",
		"Invoice generation fails too often.",
		"",
		"## Goals",
		"",
		"- generate invoices consistently",
		"",
		"## Non-Goals",
		"",
		"- redesigning every billing screen",
		"",
		"## Constraints",
		"",
		"",
		"## Solution Shape",
		"",
		"Add a worker and schema migration for invoice generation.",
		"",
		"## Flows",
		"",
		"1. Close billing cycle.",
		"2. Queue invoice generation.",
		"",
		"## Data / Interfaces",
		"",
		"- database schema update",
		"- external API export",
		"",
		"## Risks / Open Questions",
		"",
		"",
		"## Rollout",
		"",
		"- ship behind a flag",
		"",
		"## Verification",
		"",
		"",
	}, "\n")
	if _, err := manager.UpdateSpec("billing", notes.UpdateInput{Body: &body}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"--project", root, "spec", "analyze", "billing"})
	err := command.Execute()
	if err == nil {
		t.Fatalf("expected blocking analysis findings:\n%s", buf.String())
	}

	output := buf.String()
	if !strings.Contains(output, "spec_analysis: .plan/specs/billing.md") {
		t.Fatalf("expected analysis header in output:\n%s", output)
	}
	if !strings.Contains(output, "Missing Constraints:") || !strings.Contains(output, "Recommended Revisions:") {
		t.Fatalf("expected grouped analysis categories in output:\n%s", output)
	}

	note, err := notes.Read(filepath.Join(root, ".plan", "specs", "billing.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(note.Content, "## Analysis") || !strings.Contains(note.Content, "### Hidden Dependencies") {
		t.Fatalf("expected analysis section in spec note:\n%s", note.Content)
	}
}

func TestSpecChecklistCommandWritesChecklistAndFailsOnBlockingFindings(t *testing.T) {
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
		"The billing export should integrate with a third-party API.",
		"",
		"## Problem",
		"",
		"Billing exports are inconsistent.",
		"",
		"## Goals",
		"",
		"- ship export support",
		"",
		"## Non-Goals",
		"",
		"- redesign every billing screen",
		"",
		"## Constraints",
		"",
		"- keep the rollout low-risk",
		"",
		"## Solution Shape",
		"",
		"Add export support for billing.",
		"",
		"## Flows",
		"",
		"1. Trigger export.",
		"",
		"## Data / Interfaces",
		"",
		"- internal billing types",
		"",
		"## Risks / Open Questions",
		"",
		"- what should the export format be",
		"",
		"## Rollout",
		"",
		"- ship it carefully",
		"",
		"## Verification",
		"",
		"- run billing tests",
		"",
	}, "\n")
	if _, err := manager.UpdateSpec("billing", notes.UpdateInput{Body: &body}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"--project", root, "spec", "checklist", "billing", "--profile", "api-integration"})
	err := command.Execute()
	if err == nil {
		t.Fatalf("expected blocking checklist findings:\n%s", buf.String())
	}

	output := buf.String()
	if !strings.Contains(output, "spec_checklist: .plan/specs/billing.md") || !strings.Contains(output, "profile: api-integration") {
		t.Fatalf("expected checklist header in output:\n%s", output)
	}
	if !strings.Contains(output, "[error] Data / Interfaces:") {
		t.Fatalf("expected blocking checklist item in output:\n%s", output)
	}

	note, err := notes.Read(filepath.Join(root, ".plan", "specs", "billing.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(note.Content, "## Checklist") || !strings.Contains(note.Content, "### api-integration") {
		t.Fatalf("expected checklist section in spec note:\n%s", note.Content)
	}
}
