package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/planning"
	"plan/internal/workspace"
)

func TestGuideCurrentCommandEmitsJSONForActiveSession(t *testing.T) {
	root := t.TempDir()
	setupGuidePacketFixture(t, root)

	var output bytes.Buffer
	command := newRootCmd()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{"--project", root, "guide", "current", "--format", "json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected guide current to succeed: %v", err)
	}

	var packet map[string]any
	if err := json.Unmarshal(output.Bytes(), &packet); err != nil {
		t.Fatalf("expected valid JSON output: %v\n%s", err, output.String())
	}
	if packet["kind"] != "guide_packet" {
		t.Fatalf("expected guide packet kind, got %#v", packet["kind"])
	}
	session, ok := packet["session"].(map[string]any)
	if !ok {
		t.Fatalf("expected packet.session to be an object, got %#v", packet["session"])
	}
	if session["chain_id"] != "brainstorm/guide-packet-foundation" {
		t.Fatalf("unexpected chain id in packet: %#v", session)
	}
	mode, ok := packet["mode"].(map[string]any)
	if !ok {
		t.Fatalf("expected packet.mode to be an object, got %#v", packet["mode"])
	}
	if mode["stage"] != "brainstorm" {
		t.Fatalf("expected brainstorm stage: %#v", mode)
	}
}

func TestGuideCurrentCommandReturnsActionableErrorWithoutActiveSession(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	command := newRootCmd()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{"--project", root, "guide", "current", "--format", "json"})
	err := command.Execute()
	if err == nil {
		t.Fatal("expected guide current to fail without an active session")
	}
	if !strings.Contains(err.Error(), "no active guided session") {
		t.Fatalf("expected actionable missing-session error, got %v", err)
	}
}

func TestGuideShowCommandEmitsJSONForExplicitChainAndCheckpoint(t *testing.T) {
	root := t.TempDir()
	setupGuidePacketFixture(t, root)

	var output bytes.Buffer
	command := newRootCmd()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{
		"--project", root,
		"guide", "show",
		"--chain", "brainstorm/guide-packet-foundation",
		"--stage", "brainstorm",
		"--checkpoint", "clarify-open-approaches",
		"--format", "json",
	})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected guide show to succeed: %v", err)
	}

	var packet map[string]any
	if err := json.Unmarshal(output.Bytes(), &packet); err != nil {
		t.Fatalf("expected valid JSON output: %v\n%s", err, output.String())
	}
	mode, ok := packet["mode"].(map[string]any)
	if !ok {
		t.Fatalf("expected packet.mode to be an object, got %#v", packet["mode"])
	}
	if mode["checkpoint"] != "clarify-open-approaches" {
		t.Fatalf("expected explicit checkpoint override, got %#v", mode)
	}
	session, ok := packet["session"].(map[string]any)
	if !ok {
		t.Fatalf("expected packet.session to be an object, got %#v", packet["session"])
	}
	if session["current_cluster_label"] != "clarify-open-approaches" {
		t.Fatalf("expected session checkpoint to match preview override, got %#v", session)
	}
}

func TestGuideShowCommandRejectsUnsupportedStage(t *testing.T) {
	root := t.TempDir()
	setupGuidePacketFixture(t, root)

	command := newRootCmd()
	command.SetArgs([]string{
		"--project", root,
		"guide", "show",
		"--chain", "brainstorm/guide-packet-foundation",
		"--stage", "execution",
		"--format", "json",
	})
	err := command.Execute()
	if err == nil {
		t.Fatal("expected guide show to fail for unsupported stage")
	}
	if !strings.Contains(err.Error(), "guided session chain packets only support the brainstorm stage") {
		t.Fatalf("expected unsupported-stage error, got %v", err)
	}
}

func TestGuideShowCommandFailsForUnknownChain(t *testing.T) {
	root := t.TempDir()
	setupGuidePacketFixture(t, root)

	command := newRootCmd()
	command.SetArgs([]string{
		"--project", root,
		"guide", "show",
		"--chain", "brainstorm/missing",
		"--format", "json",
	})
	err := command.Execute()
	if err == nil {
		t.Fatal("expected guide show to fail for an unknown chain")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected unknown-chain error, got %v", err)
	}
}

func TestGuideShowCommandRejectsUnsupportedCheckpoint(t *testing.T) {
	root := t.TempDir()
	setupGuidePacketFixture(t, root)

	command := newRootCmd()
	command.SetArgs([]string{
		"--project", root,
		"guide", "show",
		"--chain", "brainstorm/guide-packet-foundation",
		"--stage", "brainstorm",
		"--checkpoint", "unknown",
		"--format", "json",
	})
	err := command.Execute()
	if err == nil {
		t.Fatal("expected guide show to fail for unsupported checkpoint")
	}
	if !strings.Contains(err.Error(), "unsupported brainstorm checkpoint") {
		t.Fatalf("expected unsupported-checkpoint error, got %v", err)
	}
}

func TestGuideCurrentCommandRejectsNonJSONFormat(t *testing.T) {
	root := t.TempDir()
	setupGuidePacketFixture(t, root)

	command := newRootCmd()
	command.SetArgs([]string{"--project", root, "guide", "current", "--format", "md"})
	err := command.Execute()
	if err == nil {
		t.Fatal("expected guide current to reject non-json formats")
	}
	if !strings.Contains(err.Error(), "only json is supported") {
		t.Fatalf("expected format error, got %v", err)
	}
}

func TestGuideShowCommandEmitsCollaborationPacketForBrainstormSource(t *testing.T) {
	root := t.TempDir()
	setupCollaborationGuideBrainstormFixture(t, root)

	var output bytes.Buffer
	command := newRootCmd()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{
		"--project", root,
		"guide", "show",
		"--brainstorm", "guide-packet-collaboration",
		"--stage", "promotion_review",
		"--format", "json",
	})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected collaboration guide show to succeed: %v", err)
	}

	var packet map[string]any
	if err := json.Unmarshal(output.Bytes(), &packet); err != nil {
		t.Fatalf("expected valid JSON output: %v\n%s", err, output.String())
	}
	if packet["kind"] != "guide_packet" {
		t.Fatalf("expected guide packet kind, got %#v", packet["kind"])
	}
	mode := packet["mode"].(map[string]any)
	if mode["stage"] != "promotion_review" {
		t.Fatalf("expected collaboration stage in packet: %#v", mode)
	}
	collaboration, ok := packet["collaboration"].(map[string]any)
	if !ok {
		t.Fatalf("expected collaboration payload in packet: %#v", packet["collaboration"])
	}
	if _, ok := collaboration["promotion_draft"]; !ok {
		t.Fatalf("expected embedded promotion draft: %#v", collaboration)
	}
}

func TestGuideShowCommandEmitsCollaborationPacketForDiscussionSource(t *testing.T) {
	root := t.TempDir()
	setupCollaborationGuideDiscussionFixture(t, root)

	var output bytes.Buffer
	command := newRootCmd()
	command.SetOut(&output)
	command.SetErr(&output)
	command.SetArgs([]string{
		"--project", root,
		"guide", "show",
		"--discussion", "49",
		"--stage", "initiative_draft",
		"--format", "json",
	})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected discussion guide show to succeed: %v", err)
	}

	var packet map[string]any
	if err := json.Unmarshal(output.Bytes(), &packet); err != nil {
		t.Fatalf("expected valid JSON output: %v\n%s", err, output.String())
	}
	artifact := packet["artifact"].(map[string]any)
	if artifact["type"] != "github_discussion" {
		t.Fatalf("expected discussion artifact, got %#v", artifact)
	}
	rendered, ok := packet["rendered_drafts"].([]any)
	if !ok || len(rendered) != 1 {
		t.Fatalf("expected rendered initiative draft, got %#v", packet["rendered_drafts"])
	}
}

func TestGuideShowCommandRejectsMixedChainAndCollaborationSourceFlags(t *testing.T) {
	root := t.TempDir()
	setupGuidePacketFixture(t, root)

	command := newRootCmd()
	command.SetArgs([]string{
		"--project", root,
		"guide", "show",
		"--chain", "brainstorm/guide-packet-foundation",
		"--brainstorm", "guide-packet-foundation",
		"--stage", "brainstorm",
		"--format", "json",
	})
	err := command.Execute()
	if err == nil {
		t.Fatal("expected guide show to reject mixed chain and collaboration flags")
	}
	if !strings.Contains(err.Error(), "choose either --chain or one collaboration source flag") {
		t.Fatalf("expected mixed-source error, got %v", err)
	}
}

func TestGuideShowCommandRejectsCheckpointForCollaborationSource(t *testing.T) {
	root := t.TempDir()
	setupCollaborationGuideBrainstormFixture(t, root)

	command := newRootCmd()
	command.SetArgs([]string{
		"--project", root,
		"guide", "show",
		"--brainstorm", "guide-packet-collaboration",
		"--stage", "promotion_review",
		"--checkpoint", "clarify-open-approaches",
		"--format", "json",
	})
	err := command.Execute()
	if err == nil {
		t.Fatal("expected guide show to reject --checkpoint for collaboration sources")
	}
	if !strings.Contains(err.Error(), "--checkpoint only applies to --chain guide previews") {
		t.Fatalf("expected collaboration checkpoint error, got %v", err)
	}
}

func setupGuidePacketFixture(t *testing.T, root string) {
	t.Helper()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	state, err := ws.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	state.DefaultBranch = "develop"
	if err := ws.WriteGitHubState(*state); err != nil {
		t.Fatal(err)
	}

	manager := planning.New(ws)
	if _, err := manager.CreateBrainstorm("Guide Packet Foundation"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.UpdateGuidedBrainstormIntake("guide-packet-foundation", planning.GuidedBrainstormIntakeInput{
		Vision:             "Give agents live guide packets instead of static stage prose.",
		SupportingMaterial: "docs/guide-packet.md",
	}); err != nil {
		t.Fatal(err)
	}

	sessionsPath := filepath.Join(root, ".plan", ".meta", "guided_sessions.json")
	if _, err := os.Stat(sessionsPath); err != nil {
		t.Fatalf("expected guided sessions state to exist: %v", err)
	}
}

func setupCollaborationGuideBrainstormFixture(t *testing.T, root string) {
	t.Helper()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
	if _, err := manager.CreateBrainstorm("Guide Packet Collaboration"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.UpdateGuidedBrainstormIntake("guide-packet-collaboration", planning.GuidedBrainstormIntakeInput{
		Vision: "Wrap collaboration shaping in live guide packets.",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormRefinement("guide-packet-collaboration", planning.BrainstormRefinementInput{
		Problem:             "Collaboration guidance is missing from guide packets.",
		UserValue:           "Agents can review promotion-ready payloads without guessing.",
		Constraints:         "Keep JSON canonical.\nDo not replace discuss payloads.",
		Appetite:            "One bounded collaboration slice.",
		CandidateApproaches: "Wrap assessment.\nWrap promotion draft.",
		DecisionSnapshot:    "Promote directly into one spec.",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormChallenge("guide-packet-collaboration", planning.BrainstormChallengeInput{
		NoGos:              "No custom review UI.",
		SimplerAlternative: "Embed canonical collaboration payloads in the guide packet.",
	}); err != nil {
		t.Fatal(err)
	}
}

func setupCollaborationGuideDiscussionFixture(t *testing.T, root string) {
	t.Helper()
	client := &stubGuideGitHubClient{
		preflight: &planning.GitHubRepoInfo{
			Repo:          "JimmyMcBride/plan",
			RepoURL:       "https://github.com/JimmyMcBride/plan",
			DefaultBranch: "develop",
		},
		context: &planning.GitHubContext{
			Repo: planning.GitHubRepoInfo{
				Repo:          "JimmyMcBride/plan",
				RepoURL:       "https://github.com/JimmyMcBride/plan",
				DefaultBranch: "develop",
			},
			CurrentBranch: "develop",
			CurrentSHA:    "abc123",
		},
		discussions: map[int]*planning.GitHubDiscussion{
			49: {
				Number: 49,
				URL:    "https://github.com/JimmyMcBride/plan/discussions/49",
				Title:  "Guide packet collaboration",
				Body: strings.Join([]string{
					"## Problem",
					"Guide packets need collaboration-stage coverage.",
					"",
					"## Goals",
					"Review promotion drafts in runtime packets.",
					"",
					"## Non-Goals",
					"Do not replace canonical discuss contracts.",
					"",
					"## Constraints",
					"Keep JSON canonical.",
					"",
					"## Proposed Shape",
					"Use one initiative issue plus two spec issues.",
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
	reset := planning.SetGitHubClientFactoryForTesting(func() planning.GitHubClient { return client })
	t.Cleanup(reset)

	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
}

type stubGuideGitHubClient struct {
	preflight   *planning.GitHubRepoInfo
	context     *planning.GitHubContext
	discussions map[int]*planning.GitHubDiscussion
}

func (s *stubGuideGitHubClient) Preflight(projectDir string) (*planning.GitHubRepoInfo, error) {
	return s.preflight, nil
}

func (s *stubGuideGitHubClient) CurrentContext(projectDir string) (*planning.GitHubContext, error) {
	return s.context, nil
}

func (s *stubGuideGitHubClient) CreateIssue(projectDir, repo string, input planning.GitHubIssueInput) (*planning.GitHubIssue, error) {
	return nil, nil
}

func (s *stubGuideGitHubClient) UpdateIssue(projectDir, repo string, issueNumber int, input planning.GitHubIssueInput) (*planning.GitHubIssue, error) {
	return nil, nil
}

func (s *stubGuideGitHubClient) GetIssue(projectDir, repo string, issueNumber int) (*planning.GitHubIssue, error) {
	return nil, nil
}

func (s *stubGuideGitHubClient) ListIssuesByLabel(projectDir, repo string, labels []string) ([]planning.GitHubIssue, error) {
	return nil, nil
}

func (s *stubGuideGitHubClient) EnsureLabel(projectDir, repo string, input planning.GitHubLabelInput) error {
	return nil
}

func (s *stubGuideGitHubClient) FindMilestone(projectDir, repo, title string) (*planning.GitHubMilestone, error) {
	return nil, nil
}

func (s *stubGuideGitHubClient) CreateMilestone(projectDir, repo string, input planning.GitHubMilestoneInput) (*planning.GitHubMilestone, error) {
	return nil, nil
}

func (s *stubGuideGitHubClient) GetDiscussion(projectDir, repo string, number int) (*planning.GitHubDiscussion, error) {
	return s.discussions[number], nil
}

func (s *stubGuideGitHubClient) UpdateDiscussionBody(projectDir, repo string, number int, body string) (*planning.GitHubDiscussion, error) {
	s.discussions[number].Body = body
	return s.discussions[number], nil
}

func (s *stubGuideGitHubClient) AddSubIssue(projectDir, repo string, issueNumber, subIssueNumber int) error {
	return nil
}

func (s *stubGuideGitHubClient) AddBlockedBy(projectDir, repo string, issueNumber, blockingIssueNumber int) error {
	return nil
}
