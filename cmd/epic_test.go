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

func TestEpicShapeCommandPersistsClustersAndResumes(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}

	firstInput := strings.Join([]string{
		"One extra shaping pass before spec approval.",
		"",
		"Make epics feel like bounded bets instead of loose containers.",
		"",
		"Capture appetite, scope, and success without inventing new hierarchy.",
		"",
		"Do not rebuild roadmap math.",
		"Do not add coordination dashboards.",
		"",
	}, "\n")
	var first bytes.Buffer
	command := newRootCmd()
	command.SetOut(&first)
	command.SetErr(&first)
	command.SetIn(strings.NewReader(firstInput))
	command.SetArgs([]string{"--project", root, "epic", "shape", "billing"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected first shape pass to succeed: %v\n%s", err, first.String())
	}

	secondInput := strings.Join([]string{
		"Specs inherit a clear appetite and scope boundary for later decomposition.",
		"",
	}, "\n")
	var second bytes.Buffer
	command = newRootCmd()
	command.SetOut(&second)
	command.SetErr(&second)
	command.SetIn(strings.NewReader(secondInput))
	command.SetArgs([]string{"--project", root, "epic", "shape", "billing"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected resumed shape pass to succeed: %v\n%s", err, second.String())
	}

	note, err := notes.Read(filepath.Join(root, ".plan", "epics", "billing.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(note.Content, "## Shape") || !strings.Contains(note.Content, "### Success Signal") {
		t.Fatalf("expected shape section in epic note:\n%s", note.Content)
	}
	if !strings.Contains(note.Content, "## Outcome\n\nMake epics feel like bounded bets instead of loose containers.") {
		t.Fatalf("expected top-level outcome to mirror epic shape:\n%s", note.Content)
	}
	if !strings.Contains(note.Content, "### Out of Scope\n\n- Do not rebuild roadmap math.") {
		t.Fatalf("expected out-of-scope bullets in shape section:\n%s", note.Content)
	}
}

func TestEpicHandoffAdvancesGuidedSessionToSpec(t *testing.T) {
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

	var output bytes.Buffer
	command := newRootCmd()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetIn(strings.NewReader("y\n"))
	command.SetArgs([]string{"--project", root, "epic", "handoff", "guided-planning"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected epic handoff to succeed: %v\n%s", err, output.String())
	}

	state, err := ws.ReadGuidedSessionState()
	if err != nil {
		t.Fatal(err)
	}
	session := state.Sessions["brainstorm/guided-planning"]
	if session.CurrentStage != "spec" {
		t.Fatalf("expected session to advance to spec stage: %+v", session)
	}
	if session.StageStatuses["epic"] != "done" || session.StageStatuses["spec"] != "in_progress" {
		t.Fatalf("expected stage handoff statuses: %+v", session)
	}
	if !strings.Contains(output.String(), "Using spec .plan/specs/guided-planning.md") {
		t.Fatalf("expected spec handoff output:\n%s", output.String())
	}
}
