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

func TestBrainstormRefineCommandPersistsClustersAndResumes(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
	if _, err := manager.CreateBrainstorm("Billing Flow"); err != nil {
		t.Fatal(err)
	}

	firstInput := strings.Join([]string{
		"Teams do not have a shaped billing planning flow before they start implementation.",
		"",
		"Agents get a clearer brief before promotion and spec work.",
		"",
		"Keep the tool local-first.",
		"Do not add new top-level artifacts.",
		"",
		"One focused shaping pass before promotion.",
		"",
	}, "\n")
	var first bytes.Buffer
	command := newRootCmd()
	command.SetOut(&first)
	command.SetErr(&first)
	command.SetIn(strings.NewReader(firstInput))
	command.SetArgs([]string{"--project", root, "brainstorm", "refine", "billing-flow"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected first refinement pass to succeed: %v\n%s", err, first.String())
	}

	secondInput := strings.Join([]string{
		"How opinionated should the refine prompts be?",
		"",
		"Add a guided refine command.",
		"Seed shaped output into spec promotion.",
		"",
		"Ship guided refinement before more power features.",
		"",
	}, "\n")
	var second bytes.Buffer
	command = newRootCmd()
	command.SetOut(&second)
	command.SetErr(&second)
	command.SetIn(strings.NewReader(secondInput))
	command.SetArgs([]string{"--project", root, "brainstorm", "refine", "billing-flow"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected resumed refinement pass to succeed: %v\n%s", err, second.String())
	}

	note, err := notes.Read(filepath.Join(root, ".plan", "brainstorms", "billing-flow.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(note.Content, "## Refinement") || !strings.Contains(note.Content, "### Decision Snapshot") {
		t.Fatalf("expected refinement section in note:\n%s", note.Content)
	}
	if !strings.Contains(note.Content, "## Constraints\n\n- Keep the tool local-first.") {
		t.Fatalf("expected constraints to be persisted after cluster save:\n%s", note.Content)
	}
	if !strings.Contains(note.Content, "### Candidate Approaches\n\n- Add a guided refine command.") {
		t.Fatalf("expected candidate approaches after resume:\n%s", note.Content)
	}
}

func TestBrainstormChallengeCommandPersistsClustersAndResumes(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
	if _, err := manager.CreateBrainstorm("Billing Flow"); err != nil {
		t.Fatal(err)
	}

	firstInput := strings.Join([]string{
		"Schema changes without rollout planning.",
		"Too many sidecar workflow objects.",
		"",
		"Do not add hosted services.",
		"Do not turn plan into a tracker clone.",
		"",
		"Users will accept one more planning pass.",
		"Prompt guidance can stay inspectable in git.",
		"",
		"",
	}, "\n")
	var first bytes.Buffer
	command := newRootCmd()
	command.SetOut(&first)
	command.SetErr(&first)
	command.SetIn(strings.NewReader(firstInput))
	command.SetArgs([]string{"--project", root, "brainstorm", "challenge", "billing-flow"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected first challenge pass to succeed: %v\n%s", err, first.String())
	}

	secondInput := strings.Join([]string{
		"If we add too many profiles too early, the product gets ceremonial.",
		"",
		"Start with one shaped prompt loop and one checklist pass.",
		"",
	}, "\n")
	var second bytes.Buffer
	command = newRootCmd()
	command.SetOut(&second)
	command.SetErr(&second)
	command.SetIn(strings.NewReader(secondInput))
	command.SetArgs([]string{"--project", root, "brainstorm", "challenge", "billing-flow"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected resumed challenge pass to succeed: %v\n%s", err, second.String())
	}

	note, err := notes.Read(filepath.Join(root, ".plan", "brainstorms", "billing-flow.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(note.Content, "## Challenge") || !strings.Contains(note.Content, "### Simpler Alternative") {
		t.Fatalf("expected challenge section in note:\n%s", note.Content)
	}
	if !strings.Contains(note.Content, "### Rabbit Holes\n\n- Schema changes without rollout planning.") {
		t.Fatalf("expected rabbit holes in challenge section:\n%s", note.Content)
	}
	if !strings.Contains(note.Content, "### Simpler Alternative\n\nStart with one shaped prompt loop and one checklist pass.") {
		t.Fatalf("expected simpler alternative after resume:\n%s", note.Content)
	}
}

func TestBrainstormStartGuidesVisionIntakeAndCreatesSession(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}

	input := strings.Join([]string{
		"I want plan to guide me from a rough vision to execution-ready work without making me fill out a template first.",
		"",
		"docs/guided-planning-notes.md",
		"https://example.com/research/guided-planning",
		"",
	}, "\n")

	var output bytes.Buffer
	command := newRootCmd()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetIn(strings.NewReader(input))
	command.SetArgs([]string{"--project", root, "brainstorm", "start", "Guided Planning"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected guided brainstorm start to succeed: %v\n%s", err, output.String())
	}

	note, err := notes.Read(filepath.Join(root, ".plan", "brainstorms", "guided-planning.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got := notes.ExtractSection(note.Content, "Vision"); got != "I want plan to guide me from a rough vision to execution-ready work without making me fill out a template first." {
		t.Fatalf("unexpected vision section:\n%s", got)
	}
	supporting := notes.ExtractSection(note.Content, "Supporting Material")
	if !strings.Contains(supporting, "- docs/guided-planning-notes.md") || !strings.Contains(supporting, "- https://example.com/research/guided-planning") {
		t.Fatalf("unexpected supporting material section:\n%s", supporting)
	}

	state, err := ws.ReadGuidedSessionState()
	if err != nil {
		t.Fatal(err)
	}
	session, ok := state.Sessions["brainstorm/guided-planning"]
	if !ok {
		t.Fatalf("expected guided session record: %+v", state)
	}
	if session.CurrentStage != "brainstorm" || session.CurrentClusterLabel != "vision-intake" {
		t.Fatalf("unexpected guided session progress: %+v", session)
	}
	if session.NextAction != "Continue guided brainstorm clarification." {
		t.Fatalf("unexpected guided session next action: %+v", session)
	}
	if !strings.Contains(output.String(), "Created brainstorm .plan/brainstorms/guided-planning.md") {
		t.Fatalf("expected creation output:\n%s", output.String())
	}
}

func TestBrainstormResumeShowsMenuStopsAndPersistsNextAction(t *testing.T) {
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

	var output bytes.Buffer
	command := newRootCmd()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetIn(strings.NewReader("3\n"))
	command.SetArgs([]string{"--project", root, "brainstorm", "resume", "guided-planning"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected guided resume stop flow to succeed: %v\n%s", err, output.String())
	}

	state, err := ws.ReadGuidedSessionState()
	if err != nil {
		t.Fatal(err)
	}
	session := state.Sessions["brainstorm/guided-planning"]
	if session.NextAction != "Resume the brainstorm when you are ready to continue shaping it." {
		t.Fatalf("unexpected next action after stop: %+v", session)
	}
	if !strings.Contains(output.String(), "1. Continue") || !strings.Contains(output.String(), "3. Stop for now") {
		t.Fatalf("expected numbered session menu:\n%s", output.String())
	}
}

func TestBrainstormResumeContinuesClusterWithReflectionAndGapGuidance(t *testing.T) {
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

	input := strings.Join([]string{
		"1",
		"Maybe everything around planning should get better.",
		"",
		"Not sure who benefits yet.",
		"",
	}, "\n")

	var output bytes.Buffer
	command := newRootCmd()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetIn(strings.NewReader(input))
	command.SetArgs([]string{"--project", root, "brainstorm", "resume", "guided-planning"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected guided resume continue flow to succeed: %v\n%s", err, output.String())
	}

	refinement, err := manager.ReadBrainstormRefinement("guided-planning")
	if err != nil {
		t.Fatal(err)
	}
	if refinement.Problem != "Maybe everything around planning should get better." {
		t.Fatalf("unexpected problem section:\n%s", refinement.Problem)
	}
	if refinement.UserValue != "Not sure who benefits yet." {
		t.Fatalf("unexpected user/value section:\n%s", refinement.UserValue)
	}

	state, err := ws.ReadGuidedSessionState()
	if err != nil {
		t.Fatal(err)
	}
	session := state.Sessions["brainstorm/guided-planning"]
	if session.CurrentClusterLabel != "clarify-constraints-appetite" {
		t.Fatalf("expected next brainstorm cluster after continue: %+v", session)
	}
	if !strings.Contains(output.String(), "Reflection:") {
		t.Fatalf("expected one-shot reflection output:\n%s", output.String())
	}
	if !strings.Contains(output.String(), "Potential gap: the problem or value is still vague.") {
		t.Fatalf("expected gap guidance output:\n%s", output.String())
	}
}
