package cmd

import (
	"bytes"
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/planning"
	"plan/internal/workspace"
)

func TestCheckCommandSupportsProjectEpicSpecAndStoryScopes(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
	createHealthyPlanFixture(t, root, manager)

	tests := []struct {
		name  string
		args  []string
		scope string
	}{
		{name: "project", args: []string{"check"}, scope: "check_scope: project"},
		{name: "epic", args: []string{"check", "epic", "billing"}, scope: "check_scope: epic:billing"},
		{name: "spec", args: []string{"check", "spec", "billing"}, scope: "check_scope: spec:billing"},
		{name: "story", args: []string{"check", "story", "implement-invoices"}, scope: "check_scope: story:implement-invoices"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			command := newRootCmd()
			command.SetOut(&buf)
			command.SetErr(&buf)
			command.SetArgs(append([]string{"--project", root}, tc.args...))
			if err := command.Execute(); err != nil {
				t.Fatalf("expected scope %s to pass, got error: %v\n%s", tc.name, err, buf.String())
			}
			output := buf.String()
			if !strings.Contains(output, tc.scope) {
				t.Fatalf("expected scope label in output:\n%s", output)
			}
			if !strings.Contains(output, "status: ok") {
				t.Fatalf("expected ok status in output:\n%s", output)
			}
		})
	}
}

func TestCheckCommandFailsOnBlockingFindings(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
	createHealthyPlanFixture(t, root, manager)

	storyPath := root + "/.plan/stories/implement-invoices.md"
	body := strings.Join([]string{
		"# Implement invoices",
		"",
		"Created: now",
		"",
		"## Description",
		"",
		"Create the invoice generation path for paid plans.",
		"",
		"## Acceptance Criteria",
		"",
		"- [ ] Generate invoices from line items",
		"",
		"## Verification",
		"",
	}, "\n")
	if _, err := notes.Update(storyPath, notes.UpdateInput{Body: &body}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"--project", root, "check", "story", "implement-invoices"})
	err := command.Execute()
	if err == nil {
		t.Fatalf("expected blocking findings to fail command:\n%s", buf.String())
	}
	output := buf.String()
	if !strings.Contains(output, "[error] story .plan/stories/implement-invoices.md :: Verification") {
		t.Fatalf("expected blocking story finding in output:\n%s", output)
	}
}

func TestCheckCommandShowsWarningsWithoutFailing(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
	createHealthyPlanFixture(t, root, manager)

	storyPath := root + "/.plan/stories/implement-invoices.md"
	body := strings.Join([]string{
		"# Implement invoices",
		"",
		"Created: now",
		"",
		"## Description",
		"",
		"Invoices.",
		"",
		"## Acceptance Criteria",
		"",
		"- [ ] Generate invoices from line items",
		"",
		"## Verification",
		"",
		"- Run focused billing tests",
		"",
	}, "\n")
	if _, err := notes.Update(storyPath, notes.UpdateInput{Body: &body}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"--project", root, "check", "story", "implement-invoices"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected warning-only check to pass: %v\n%s", err, buf.String())
	}
	output := buf.String()
	if !strings.Contains(output, "[warn] story .plan/stories/implement-invoices.md :: Description") {
		t.Fatalf("expected warning in output:\n%s", output)
	}
}

func createHealthyPlanFixture(t *testing.T, root string, manager *planning.Manager) {
	t.Helper()
	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}
	specBody := strings.Join([]string{
		"# Billing Spec",
		"",
		"Created: now",
		"",
		"## Why",
		"",
		"Billing drives plan monetization.",
		"",
		"## Problem",
		"",
		"Teams cannot generate invoices from the current billing data, which blocks manual reconciliation and delays payments.",
		"",
		"## Goals",
		"",
		"- generate invoices from line items",
		"- keep invoice data consistent with billing periods",
		"",
		"## Non-Goals",
		"",
		"- building tax automation",
		"- redesigning subscription management",
		"",
		"## Constraints",
		"",
		"- keep the implementation local-first",
		"- reuse the current billing schema",
		"",
		"## Solution Shape",
		"",
		"Add a billing pipeline that emits invoice records and exposes them for export.",
		"",
		"## Flows",
		"",
		"1. User closes a billing cycle.",
		"2. The system generates invoice records.",
		"",
		"## Data / Interfaces",
		"",
		"- invoice record output",
		"",
		"## Risks / Open Questions",
		"",
		"- How should failed exports retry?",
		"",
		"## Rollout",
		"",
		"- dogfood with the local workspace first",
		"",
		"## Verification",
		"",
		"- run focused billing tests",
		"- generate a sample invoice export locally",
		"",
	}, "\n")
	if _, err := manager.UpdateSpec("billing", notes.UpdateInput{
		Body: &specBody,
		Metadata: map[string]any{
			"status":         "approved",
			"target_version": "v2",
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.CreateStory(
		"billing",
		"Implement invoices",
		"Create the invoice generation path for paid plans.",
		[]string{"Generate invoices from line items"},
		[]string{"Run focused billing tests"},
		nil,
	); err != nil {
		t.Fatal(err)
	}
}
