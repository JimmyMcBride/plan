package cmd

import (
	"bytes"
	"testing"

	"plan/internal/workspace"
)

func TestSourceCommandShowsAndSetsMode(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}

	var showBuf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&showBuf)
	command.SetErr(&showBuf)
	command.SetArgs([]string{"--project", root, "source", "show"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected source show to succeed: %v\n%s", err, showBuf.String())
	}
	if showBuf.String() != "source_mode: local\n" {
		t.Fatalf("unexpected source show output:\n%s", showBuf.String())
	}

	var setBuf bytes.Buffer
	command = newRootCmd()
	command.SetOut(&setBuf)
	command.SetErr(&setBuf)
	command.SetArgs([]string{"--project", root, "source", "set", "hybrid"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected source set to succeed: %v\n%s", err, setBuf.String())
	}
	if setBuf.String() != "source_mode: hybrid\n" {
		t.Fatalf("unexpected source set output:\n%s", setBuf.String())
	}

	var linearBuf bytes.Buffer
	command = newRootCmd()
	command.SetOut(&linearBuf)
	command.SetErr(&linearBuf)
	command.SetArgs([]string{"--project", root, "source", "set", "linear"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected source set linear to succeed: %v\n%s", err, linearBuf.String())
	}
	if linearBuf.String() != "source_mode: linear\n" {
		t.Fatalf("unexpected source set linear output:\n%s", linearBuf.String())
	}

	var linearShow bytes.Buffer
	command = newRootCmd()
	command.SetOut(&linearShow)
	command.SetErr(&linearShow)
	command.SetArgs([]string{"--project", root, "source", "show"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected source show linear to succeed: %v\n%s", err, linearShow.String())
	}
	for _, want := range []string{
		"source_mode: linear\n",
		"linear_team: not_configured\n",
		"linear_ownership: durable planning data lives in Linear after promotion\n",
		"linear_guidance: configure .plan/.meta/linear.json with team_id or team_key before Linear promotion\n",
	} {
		if !bytes.Contains(linearShow.Bytes(), []byte(want)) {
			t.Fatalf("expected source show linear output to contain %q:\n%s", want, linearShow.String())
		}
	}
}
