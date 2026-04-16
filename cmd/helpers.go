package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"plan/internal/planning"
	"plan/internal/workspace"
)

func workspaceManager() *workspace.Manager {
	return workspace.New(projectDir)
}

func planningManager() *planning.Manager {
	return planning.New(workspaceManager())
}

func readBody(stdin io.Reader, body string, useStdin bool) (string, error) {
	switch {
	case useStdin:
		raw, err := io.ReadAll(stdin)
		if err != nil {
			return "", err
		}
		return string(raw), nil
	default:
		return body, nil
	}
}

func editTextInEditor(initial, editor string) (string, error) {
	cmdName := strings.TrimSpace(editor)
	if cmdName == "" {
		cmdName = strings.TrimSpace(os.Getenv("VISUAL"))
	}
	if cmdName == "" {
		cmdName = strings.TrimSpace(os.Getenv("EDITOR"))
	}
	if cmdName == "" {
		return "", errors.New("no editor configured (set $EDITOR or use --editor)")
	}

	tempDir := os.TempDir()
	file, err := os.CreateTemp(tempDir, "plan-edit-*.md")
	if err != nil {
		return "", err
	}
	path := file.Name()
	defer os.Remove(path)
	defer file.Close()

	if _, err := file.WriteString(initial); err != nil {
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}

	parts := strings.Fields(cmdName)
	bin := parts[0]
	args := append(parts[1:], path)
	cmd := exec.Command(bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor %s: %w", bin, err)
	}

	raw, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
