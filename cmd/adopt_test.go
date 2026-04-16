package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/workspace"
)

func TestAdoptCommandCreatesManagedWorkspace(t *testing.T) {
	root := t.TempDir()
	readmePath := filepath.Join(root, "README.md")
	const readme = "# existing repo\n"
	if err := os.WriteFile(readmePath, []byte(readme), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"--project", root, "adopt"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected adopt command to succeed: %v\n%s", err, buf.String())
	}

	output := buf.String()
	if !strings.Contains(output, "Adopted plan workspace") {
		t.Fatalf("expected adopt output:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(root, ".plan", ".meta", "workspace.json")); err != nil {
		t.Fatalf("expected workspace metadata after adopt: %v", err)
	}
	raw, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != readme {
		t.Fatalf("expected adopt to leave repo files alone:\n%s", raw)
	}

	report, err := workspace.New(root).Doctor()
	if err != nil {
		t.Fatal(err)
	}
	if report.WorkspaceStatus != "current" {
		t.Fatalf("expected adopted workspace to be current: %+v", report)
	}
}
