package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillsInstallPrintsInstalledSkillNames(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"skills", "install", "--scope", "global", "--agent", "codex"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected skills install to succeed: %v\n%s", err, buf.String())
	}

	output := buf.String()
	for _, skill := range []string{"plan", "plan-execute"} {
		if !strings.Contains(output, "codex [global] "+skill+" copy ->") {
			t.Fatalf("expected install output to include %s:\n%s", skill, output)
		}
		if _, err := os.Stat(filepath.Join(home, ".codex", "skills", skill, "SKILL.md")); err != nil {
			t.Fatalf("expected installed %s skill: %v", skill, err)
		}
	}
}
