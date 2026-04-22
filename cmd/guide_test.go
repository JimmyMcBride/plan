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
	if !strings.Contains(err.Error(), "only support the brainstorm stage") {
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
	if !strings.Contains(err.Error(), "only json is supported in v1") {
		t.Fatalf("expected format error, got %v", err)
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
