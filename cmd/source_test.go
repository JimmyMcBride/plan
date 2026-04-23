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
}
