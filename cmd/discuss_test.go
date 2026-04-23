package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"plan/internal/planning"
	"plan/internal/workspace"
)

func TestDiscussAssessCommandPrintsJSONForBrainstorm(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
	if _, err := manager.CreateBrainstorm("Discussion Assess"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.UpdateGuidedBrainstormIntake("discussion-assess", planning.GuidedBrainstormIntakeInput{
		Vision: "Assess whether a brainstorm is ready to promote.",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormRefinement("discussion-assess", planning.BrainstormRefinementInput{
		Problem:                "Promotion readiness is inconsistent.",
		UserValue:              "The user gets a deterministic gate.",
		Constraints:            "Keep the first slice JSON-only.",
		Appetite:               "Small.",
		RemainingOpenQuestions: "How should confirm be spelled?",
		CandidateApproaches:    "Assess.\nDraft promotion.",
		DecisionSnapshot:       "One spec is enough.",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormChallenge("discussion-assess", planning.BrainstormChallengeInput{
		RabbitHoles:           "Do not auto-post.",
		NoGos:                 "No custom review UI.",
		Assumptions:           "JSON output is enough for the first slice.",
		LikelyOverengineering: "Supporting every backend immediately.",
		SimplerAlternative:    "Assess then promote.",
	}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"--project", root, "discuss", "assess", "--brainstorm", "discussion-assess", "--format", "json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected discuss assess to succeed: %v\n%s", err, buf.String())
	}

	var payload planning.CollaborationAssessment
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON output, got %v\n%s", err, buf.String())
	}
	if payload.Kind != "maturity_assessment" || payload.Decision.State != planning.MaturityReadySingleSpec {
		t.Fatalf("unexpected discuss assess payload: %+v", payload)
	}
}

func TestDiscussPromoteCommandBuildsDraftJSON(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
	if _, err := manager.CreateBrainstorm("Discussion Promote"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.UpdateGuidedBrainstormIntake("discussion-promote", planning.GuidedBrainstormIntakeInput{
		Vision: "Promote a brainstorm into a reviewed draft.",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormRefinement("discussion-promote", planning.BrainstormRefinementInput{
		Problem:                "Users need a promotion draft before writes happen.",
		UserValue:              "They can review the shape before GitHub changes.",
		Constraints:            "Keep the first slice json-only.",
		Appetite:               "Small.",
		RemainingOpenQuestions: "How should apply work?",
		CandidateApproaches:    "Build promotion draft review.",
		DecisionSnapshot:       "Promote directly into one spec.",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormChallenge("discussion-promote", planning.BrainstormChallengeInput{
		RabbitHoles:           "Do not write yet.",
		NoGos:                 "No hidden writes.",
		Assumptions:           "The draft review can be consumed by an agent.",
		LikelyOverengineering: "Adding milestone creation before promotion works.",
		SimplerAlternative:    "Build the draft first.",
	}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"--project", root, "discuss", "promote", "--brainstorm", "discussion-promote", "--format", "json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected discuss promote to succeed: %v\n%s", err, buf.String())
	}

	var payload planning.PromotionDraft
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON output, got %v\n%s", err, buf.String())
	}
	if payload.Kind != "promotion_draft" || len(payload.ProposedSpecIssues) != 1 || !payload.ConfirmationRequired {
		t.Fatalf("unexpected discuss promote payload: %+v", payload)
	}
}
