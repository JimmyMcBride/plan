package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/workspace"
)

func TestBrainImportInspectCommandPrintsPlanningCandidates(t *testing.T) {
	workspacePath := brainFixturePath(t)

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"import", "brain", "inspect", "--workspace", workspacePath})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected inspect command to succeed: %v\n%s", err, buf.String())
	}

	output := buf.String()
	if !strings.Contains(output, "brain_workspace:") {
		t.Fatalf("expected workspace header in output:\n%s", output)
	}
	if !strings.Contains(output, "brainstorms:") || !strings.Contains(output, "epics:") || !strings.Contains(output, "specs:") || !strings.Contains(output, "stories:") {
		t.Fatalf("expected planning candidate groups in output:\n%s", output)
	}
	if strings.Contains(output, "context/") {
		t.Fatalf("expected inspect output to avoid brain context surfaces:\n%s", output)
	}
}

func TestBrainImportApplyCommandPrintsMappings(t *testing.T) {
	root := t.TempDir()
	if _, err := workspace.New(root).Init(); err != nil {
		t.Fatal(err)
	}
	workspacePath := brainFixturePath(t)

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{
		"--project", root,
		"import", "brain", "apply",
		"--workspace", workspacePath,
		"--epic", "planning-and-brainstorming-ux",
	})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected apply command to succeed: %v\n%s", err, buf.String())
	}

	output := buf.String()
	if !strings.Contains(output, "imported: 1") {
		t.Fatalf("expected import count in output:\n%s", output)
	}
	if !strings.Contains(output, ".brain/planning/epics/planning-and-brainstorming-ux.md -> .plan/epics/planning-and-brainstorming-ux.md") {
		t.Fatalf("expected source-to-destination mapping in output:\n%s", output)
	}
	if !strings.Contains(output, "review: inspect imported notes before execution work") {
		t.Fatalf("expected review guidance in output:\n%s", output)
	}
}

func brainFixturePath(t *testing.T) string {
	t.Helper()

	path, err := filepath.Abs(filepath.Join("..", "testdata", "brain-workspace"))
	if err != nil {
		t.Fatalf("resolve brain fixture path: %v", err)
	}
	return path
}
