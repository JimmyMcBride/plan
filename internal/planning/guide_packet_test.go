package planning

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/workspace"
)

func TestCurrentGuidePacketBuildsBrainstormPacketWithoutMutatingSessionState(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}

	githubState, err := ws.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	githubState.DefaultBranch = "develop"
	if err := ws.WriteGitHubState(*githubState); err != nil {
		t.Fatal(err)
	}

	manager := New(ws)
	if _, err := manager.CreateBrainstorm("Guide Packet Foundation"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.UpdateGuidedBrainstormIntake("guide-packet-foundation", GuidedBrainstormIntakeInput{
		Vision:             "Guide the user from idea to a live planning contract.",
		SupportingMaterial: "docs/guide-packet.md",
	}); err != nil {
		t.Fatal(err)
	}

	sessionsPath := filepath.Join(root, ".plan", ".meta", "guided_sessions.json")
	before, err := os.ReadFile(sessionsPath)
	if err != nil {
		t.Fatal(err)
	}

	packet, err := manager.CurrentGuidePacket()
	if err != nil {
		t.Fatal(err)
	}
	if packet.Kind != guidePacketKind || packet.SchemaVersion != GuidePacketSchemaVersion {
		t.Fatalf("unexpected packet identity: %+v", packet)
	}
	if packet.Workspace.IntegrationBranch != "develop" {
		t.Fatalf("expected integration branch from github state: %+v", packet.Workspace)
	}
	if packet.Session.ChainID != "brainstorm/guide-packet-foundation" {
		t.Fatalf("unexpected chain id: %+v", packet.Session)
	}
	if packet.Artifact.Path != ".plan/brainstorms/guide-packet-foundation.md" {
		t.Fatalf("unexpected artifact path: %+v", packet.Artifact)
	}
	if len(packet.Contract.Stance) == 0 || packet.Contract.QuestionStrategy.GapGuidance == "" {
		t.Fatalf("expected richer contract guidance: %+v", packet.Contract)
	}
	if len(packet.Contract.CommandHints) != 2 {
		t.Fatalf("expected command hints: %+v", packet.Contract.CommandHints)
	}
	if !strings.Contains(packet.Contract.CommandHints[1].Command, "--chain brainstorm/guide-packet-foundation") {
		t.Fatalf("expected explicit chain in command hints: %+v", packet.Contract.CommandHints)
	}
	if !strings.Contains(packet.RenderedPrompt, "Goal: ") {
		t.Fatalf("expected rendered prompt to be derived from the contract: %s", packet.RenderedPrompt)
	}

	after, err := os.ReadFile(sessionsPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("guide packet generation should not mutate guided sessions state\nbefore:\n%s\nafter:\n%s", string(before), string(after))
	}
}

func TestCurrentGuidePacketFailsWithoutActiveSession(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	_, err := manager.CurrentGuidePacket()
	if err == nil {
		t.Fatal("expected missing active session error")
	}
	if !strings.Contains(err.Error(), "no active guided session") {
		t.Fatalf("expected actionable missing-session error, got %v", err)
	}
	if !errors.Is(err, ErrNoActiveGuidedSession) {
		t.Fatalf("expected missing-session error to wrap ErrNoActiveGuidedSession, got %v", err)
	}
}

func TestGuidePacketForChainFailsWhenArtifactIsMissing(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)
	if _, err := manager.CreateBrainstorm("Guide Packet Foundation"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.UpdateGuidedBrainstormIntake("guide-packet-foundation", GuidedBrainstormIntakeInput{
		Vision:             "Guide packet planning should be runtime-driven.",
		SupportingMaterial: "docs/guide-packet.md",
	}); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(filepath.Join(root, ".plan", "brainstorms", "guide-packet-foundation.md")); err != nil {
		t.Fatal(err)
	}

	_, err := manager.GuidePacketForChain("brainstorm/guide-packet-foundation", "brainstorm", "")
	if err == nil {
		t.Fatal("expected missing artifact error")
	}
	if !strings.Contains(err.Error(), "read brainstorm artifact") {
		t.Fatalf("expected missing-artifact error, got %v", err)
	}
}

func TestCurrentGuidePacketDefaultsBlankSourceModeToLocal(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	info, err := ws.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(info.WorkspaceFile)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	payload["source_mode"] = ""
	updated, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(info.WorkspaceFile, updated, 0o644); err != nil {
		t.Fatal(err)
	}

	manager := New(ws)
	if _, err := manager.CreateBrainstorm("Guide Packet Source Mode"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.UpdateGuidedBrainstormIntake("guide-packet-source-mode", GuidedBrainstormIntakeInput{
		Vision: "Guide packets should emit a normalized source mode.",
	}); err != nil {
		t.Fatal(err)
	}

	packet, err := manager.CurrentGuidePacket()
	if err != nil {
		t.Fatal(err)
	}
	if packet.Workspace.SourceMode != "local" {
		t.Fatalf("expected local source mode fallback: %+v", packet.Workspace)
	}
	if packet.Ownership.Mode != SourceOfTruthLocal {
		t.Fatalf("expected local ownership fallback: %+v", packet.Ownership)
	}
}

func TestGuidePacketForLocalCollaborationSourceEmbedsCanonicalDraftsAndActions(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)
	createReadyCollaborationBrainstorm(t, manager, "guide-packet-collaboration")

	packet, err := manager.GuidePacketForCollaborationSource("guide-packet-collaboration", "", "promotion_review")
	if err != nil {
		t.Fatal(err)
	}
	if packet.Workspace.PlanningMode != guidePlanningModeCollab {
		t.Fatalf("expected collaboration planning mode: %+v", packet.Workspace)
	}
	if packet.Collaboration == nil || packet.Collaboration.Assessment == nil || packet.Collaboration.PromotionDraft == nil {
		t.Fatalf("expected embedded collaboration payloads: %+v", packet.Collaboration)
	}
	if packet.Collaboration.Assessment.Kind != maturityAssessmentKind {
		t.Fatalf("expected canonical assessment payload: %+v", packet.Collaboration.Assessment)
	}
	if packet.Collaboration.PromotionDraft.Kind != promotionDraftKind {
		t.Fatalf("expected canonical promotion draft payload: %+v", packet.Collaboration.PromotionDraft)
	}
	if len(packet.RenderedDrafts) != 1 || packet.RenderedDrafts[0].Kind != "spec_issue" {
		t.Fatalf("expected rendered spec draft in packet: %+v", packet.RenderedDrafts)
	}
	if len(packet.Actions) < 3 {
		t.Fatalf("expected structured review/apply actions: %+v", packet.Actions)
	}
	var refreshAction *GuidePacketAction
	for i := range packet.Actions {
		if packet.Actions[i].ID == "refresh_promotion_draft" {
			refreshAction = &packet.Actions[i]
			break
		}
	}
	if refreshAction == nil {
		t.Fatalf("expected refresh draft action in packet: %+v", packet.Actions)
	}
	if !refreshAction.Available || refreshAction.Command == "" {
		t.Fatalf("expected refresh draft action with command: %+v", *refreshAction)
	}
	foundConfirm := false
	for _, action := range packet.Actions {
		if action.RequiresConfirmation {
			foundConfirm = true
			break
		}
	}
	if !foundConfirm {
		t.Fatalf("expected explicit confirm action in packet: %+v", packet.Actions)
	}
}

func TestCollaborationSourceCommandTrimsSelectorValues(t *testing.T) {
	command, display := collaborationSourceCommand("  demo-brainstorm  ", "")
	if command != "--brainstorm demo-brainstorm" || display != "local brainstorm demo-brainstorm" {
		t.Fatalf("expected trimmed brainstorm selector, got command=%q display=%q", command, display)
	}

	command, display = collaborationSourceCommand("", "  49  ")
	if command != "--discussion 49" || display != "GitHub Discussion 49" {
		t.Fatalf("expected trimmed discussion selector, got command=%q display=%q", command, display)
	}
}

func TestGuidePacketForDiscussionSourceSupportsInitiativeDraftStage(t *testing.T) {
	client := &stubGitHubClient{
		preflight: &GitHubRepoInfo{
			Repo:          "JimmyMcBride/plan",
			RepoURL:       "https://github.com/JimmyMcBride/plan",
			DefaultBranch: "develop",
		},
		context: &GitHubContext{
			Repo: GitHubRepoInfo{
				Repo:          "JimmyMcBride/plan",
				RepoURL:       "https://github.com/JimmyMcBride/plan",
				DefaultBranch: "develop",
			},
			CurrentBranch: "develop",
			CurrentSHA:    "abc123",
		},
		discussions: map[int]*GitHubDiscussion{
			49: {
				Number: 49,
				URL:    "https://github.com/JimmyMcBride/plan/discussions/49",
				Title:  "Guide packet collaboration",
				Body: strings.Join([]string{
					"## Problem",
					"Guide packets need collaboration-stage coverage.",
					"",
					"## Goals",
					"Wrap the existing discuss contracts in runtime guidance.",
					"",
					"## Non-Goals",
					"Do not replace canonical collaboration payloads.",
					"",
					"## Constraints",
					"Keep JSON canonical and review-first.",
					"",
					"## Proposed Shape",
					"Use one initiative issue and two spec issues.",
					"",
					"## Spec Split",
					"- Guide packet v2 schema",
					"- Guide packet v2 CLI",
					"",
					"Guide packet v2 CLI depends on Guide packet v2 schema.",
				}, "\n"),
			},
		},
	}
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	defer reset()

	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	packet, err := manager.GuidePacketForCollaborationSource("", "49", "initiative_draft")
	if err != nil {
		t.Fatal(err)
	}
	if packet.Collaboration == nil || packet.Collaboration.Source.Mode != CollaborationSourceGitHubDiscussion {
		t.Fatalf("expected GitHub Discussion collaboration source: %+v", packet.Collaboration)
	}
	if len(packet.RenderedDrafts) != 1 || packet.RenderedDrafts[0].Kind != "initiative_issue" {
		t.Fatalf("expected initiative draft rendering: %+v", packet.RenderedDrafts)
	}
	if packet.Artifact.Type != "github_discussion" {
		t.Fatalf("expected discussion artifact: %+v", packet.Artifact)
	}
	if !strings.Contains(packet.RenderedPrompt, "Promotion decision: multi_spec") {
		t.Fatalf("expected rendered prompt to include collaboration draft context: %s", packet.RenderedPrompt)
	}
}

func TestGuidePacketNeedsRefinementStageRendersExplicitExceptions(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)
	createReadyCollaborationBrainstorm(t, manager, "guide-packet-needs-refinement")
	if _, err := manager.UpdateBrainstormRefinement("guide-packet-needs-refinement", BrainstormRefinementInput{
		Problem:                "Guide packet v2 needs a concrete refinement exception path.",
		UserValue:              "The user can see exactly why a spec is not ready yet.",
		Constraints:            "Keep JSON canonical.",
		Appetite:               "One collaboration slice.",
		RemainingOpenQuestions: "Needs-refinement: clarify the verification contract before execution.",
		CandidateApproaches:    "Wrap discuss contracts.\nRender stage prompts.",
		DecisionSnapshot:       "Promote directly into one spec once the gap is explicit.",
	}); err != nil {
		t.Fatal(err)
	}

	packet, err := manager.GuidePacketForCollaborationSource("guide-packet-needs-refinement", "", "needs_refinement")
	if err != nil {
		t.Fatal(err)
	}
	if packet.Collaboration == nil || packet.Collaboration.PromotionDraft == nil {
		t.Fatalf("expected promotion draft in needs-refinement packet: %+v", packet.Collaboration)
	}
	if len(packet.Collaboration.PromotionDraft.NeedsRefinementExceptions) != 1 {
		t.Fatalf("expected one explicit refinement exception: %+v", packet.Collaboration.PromotionDraft.NeedsRefinementExceptions)
	}
	if len(packet.RenderedDrafts) != 1 || packet.RenderedDrafts[0].Kind != "refinement_exception" {
		t.Fatalf("expected rendered refinement exception: %+v", packet.RenderedDrafts)
	}
	if !strings.Contains(packet.RenderedDrafts[0].Body, "## Recommended Clarification") {
		t.Fatalf("expected rendered refinement markdown body: %s", packet.RenderedDrafts[0].Body)
	}
}

func createReadyCollaborationBrainstorm(t *testing.T, manager *Manager, slug string) {
	t.Helper()
	title := strings.ReplaceAll(slug, "-", " ")
	if _, err := manager.CreateBrainstorm(title); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.UpdateGuidedBrainstormIntake(slug, GuidedBrainstormIntakeInput{
		Vision:             "Wrap collaboration shaping in runtime guide packets.",
		SupportingMaterial: "docs/using-plan.md",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormRefinement(slug, BrainstormRefinementInput{
		Problem:                "Guide packets stop at brainstorm guidance today.",
		UserValue:              "Agents need runtime help around collaboration shaping and promotion review.",
		Constraints:            "Keep JSON canonical.\nDo not replace discuss contracts.",
		Appetite:               "One bounded collaboration slice.",
		RemainingOpenQuestions: "",
		CandidateApproaches:    "Wrap maturity assessment.\nWrap promotion draft.",
		DecisionSnapshot:       "Promote directly into one spec while collaboration scope stays bounded.",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormChallenge(slug, BrainstormChallengeInput{
		RabbitHoles:           "Do not add execution-state automation.",
		NoGos:                 "No custom review UI.",
		Assumptions:           "Guide packets can stay JSON-first.",
		LikelyOverengineering: "Replacing canonical discuss payloads.",
		SimplerAlternative:    "Embed the canonical payloads and add runtime guidance around them.",
	}); err != nil {
		t.Fatal(err)
	}
}
