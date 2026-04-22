package cmd

import (
	"bytes"
	"os"
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

func TestSpecInitiativeCommandUpdatesMetadata(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	command := newRootCmd()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{
		"--project", root,
		"spec", "initiative", "billing",
		"--set", "guide-packet-foundation",
		"--title", "Guide Packet Foundation",
		"--summary", "Ship the first guide-packet workflow slices.",
	})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected spec initiative command to succeed: %v\n%s", err, output.String())
	}

	spec, err := notes.Read(filepath.Join(root, ".plan", "specs", "billing.md"))
	if err != nil {
		t.Fatal(err)
	}
	if spec.Metadata["initiative"] != "guide-packet-foundation" || spec.Metadata["initiative_title"] != "Guide Packet Foundation" {
		t.Fatalf("expected initiative metadata on spec: %+v", spec.Metadata)
	}
}

func TestSpecExecuteCommandStartsImplementingWithoutCreatingStories(t *testing.T) {
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
		"Billing execution needs a spec-driven workflow.",
		"",
		"## Problem",
		"",
		"Execution still depends on durable story files.",
		"",
		"## Goals",
		"",
		"- start execution from the spec",
		"",
		"## Non-Goals",
		"",
		"- rebuilding project management",
		"",
		"## Constraints",
		"",
		"- keep slices ephemeral",
		"",
		"## Solution Shape",
		"",
		"Drive execution directly from the approved spec.",
		"",
		"## Flows",
		"",
		"1. Start execution from the approved spec.",
		"2. Work through slices in order.",
		"",
		"## Data / Interfaces",
		"",
		"- spec execution plan",
		"",
		"## Risks / Open Questions",
		"",
		"- branch naming consistency",
		"",
		"## Rollout",
		"",
		"- dogfood in the repo",
		"",
		"## Verification",
		"",
		"- go test ./...",
		"",
		"## Execution Plan",
		"",
		"- Prepare billing execution",
		"  - description: validate the interfaces and rollout constraints before writing code",
		"- Implement billing execution",
		"  - description: build the main billing behavior from the spec",
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

	var output bytes.Buffer
	command := newRootCmd()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{"--project", root, "spec", "execute", "billing"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected spec execute to succeed: %v\n%s", err, output.String())
	}

	spec, err := notes.Read(filepath.Join(root, ".plan", "specs", "billing.md"))
	if err != nil {
		t.Fatal(err)
	}
	if spec.Metadata["status"] != "implementing" {
		t.Fatalf("expected execute to mark spec implementing: %+v", spec.Metadata)
	}
	if _, err := notes.Read(filepath.Join(root, ".plan", "stories", "prepare-billing-execution.md")); !os.IsNotExist(err) {
		t.Fatalf("expected execute to avoid creating story artifacts, got %v", err)
	}
	if !strings.Contains(output.String(), "spec_execution: .plan/specs/billing.md") || !strings.Contains(output.String(), "branch: feature/billing") {
		t.Fatalf("expected execution output in command result:\n%s", output.String())
	}
}

func TestSpecHandoffStartsExecutionAndAdvancesSessionToExecution(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
	if _, err := manager.CreateBrainstorm("Guided Planning"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.UpdateGuidedBrainstormIntake("guided-planning", planning.GuidedBrainstormIntakeInput{
		Vision:             "Guide a user from a rough feature idea into a shaped plan.",
		SupportingMaterial: "docs/research.md",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormRefinement("guided-planning", planning.BrainstormRefinementInput{
		Problem:                "Planning starts too artifact-first.",
		UserValue:              "Users get a collaborative planning flow.",
		Constraints:            "Keep the tool local-first.",
		Appetite:               "One focused planning session.",
		RemainingOpenQuestions: "How far should the guided loop go in v1?",
		CandidateApproaches:    "Promote at an explicit checkpoint.",
	}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.PromoteGuidedBrainstormSession("guided-planning"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.AdvanceGuidedSessionToSpec("guided-planning"); err != nil {
		t.Fatal(err)
	}

	body := strings.Join([]string{
		"# Guided Planning Spec",
		"",
		"Created: now",
		"",
		"## Why",
		"",
		"Guided planning needs a spec-driven execution handoff.",
		"",
		"## Problem",
		"",
		"Execution still depends on persistent story creation.",
		"",
		"## Goals",
		"",
		"- derive a first execution pass",
		"",
		"## Non-Goals",
		"",
		"- reinvent task tracking",
		"",
		"## Constraints",
		"",
		"- keep the output local-first",
		"",
		"## Solution Shape",
		"",
		"Use the approved spec to drive execution guidance.",
		"",
		"## Flows",
		"",
		"1. Review spec.",
		"2. Start execution.",
		"",
		"## Data / Interfaces",
		"",
		"- guided session state",
		"",
		"## Risks / Open Questions",
		"",
		"- duplicate stories",
		"",
		"## Rollout",
		"",
		"- dogfood locally",
		"",
		"## Verification",
		"",
		"- run spec execution coverage",
		"",
		"## Execution Plan",
		"",
		"- Carry forward recap state",
		"- Build the first guided execution pass",
		"",
	}, "\n")
	if _, err := manager.UpdateSpec("guided-planning", notes.UpdateInput{
		Body: &body,
		Metadata: map[string]any{
			"status": "approved",
		},
	}); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	command := newRootCmd()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetIn(strings.NewReader("y\n"))
	command.SetArgs([]string{"--project", root, "spec", "handoff", "guided-planning"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected spec handoff to succeed: %v\n%s", err, output.String())
	}

	state, err := ws.ReadGuidedSessionState()
	if err != nil {
		t.Fatal(err)
	}
	session := state.Sessions["brainstorm/guided-planning"]
	if session.CurrentStage != "execution" {
		t.Fatalf("expected session to advance to execution stage: %+v", session)
	}
	if session.StageStatuses["spec"] != "done" || session.StageStatuses["execution"] != "in_progress" {
		t.Fatalf("expected execution-stage handoff statuses: %+v", session)
	}
	spec, err := notes.Read(filepath.Join(root, ".plan", "specs", "guided-planning.md"))
	if err != nil {
		t.Fatal(err)
	}
	if spec.Metadata["status"] != "implementing" {
		t.Fatalf("expected handoff to mark spec implementing: %+v", spec.Metadata)
	}
	if _, err := notes.Read(filepath.Join(root, ".plan", "stories", "carry-forward-recap-state.md")); !os.IsNotExist(err) {
		t.Fatalf("expected handoff to avoid creating stories, got %v", err)
	}
	if !strings.Contains(output.String(), "spec_execution: .plan/specs/guided-planning.md") || !strings.Contains(output.String(), "branch: feature/guided-planning") {
		t.Fatalf("expected execution handoff output:\n%s", output.String())
	}
}
