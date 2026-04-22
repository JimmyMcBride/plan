package planning

import (
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
