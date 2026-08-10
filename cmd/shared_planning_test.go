package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"plan/internal/planning"
	"plan/internal/workspace"
)

const pinnedBrainPlanningVersion = "github.com/JimmyMcBride/brain v0.1.15-0.20260810065410-c2c71279030f"

func TestBrainPlanningDependencyIsPinnedWithoutReplace(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	if !strings.Contains(content, pinnedBrainPlanningVersion) {
		t.Fatalf("Brain dependency is not pinned to merged revision: %s", content)
	}
	if strings.Contains(content, "replace github.com/JimmyMcBride/brain") {
		t.Fatal("Brain dependency must resolve without a local replace directive")
	}
}

func TestSharedPlanningDispatchesOnlyCompatibleLocalWorkspace(t *testing.T) {
	root := t.TempDir()
	manager := workspace.New(root)
	if _, err := manager.Init(); err != nil {
		t.Fatal(err)
	}
	projectDir = root
	_, shared, err := sharedPlanningService(t.Context())
	if err != nil || !shared {
		t.Fatalf("expected local workspace to use shared Planning: shared=%v err=%v", shared, err)
	}
	meta, err := manager.ReadWorkspaceMeta()
	if err != nil {
		t.Fatal(err)
	}
	meta.SourceMode = workspace.SourceOfTruthGitHub
	if err := manager.WriteWorkspaceMeta(*meta); err != nil {
		t.Fatal(err)
	}
	_, shared, err = sharedPlanningService(t.Context())
	if err != nil || shared {
		t.Fatalf("expected GitHub workspace to remain standalone: shared=%v err=%v", shared, err)
	}
}

func TestSharedPlanningWarningIsInteractiveOncePerProcess(t *testing.T) {
	root := t.TempDir()
	manager := workspace.New(root)
	if _, err := manager.Init(); err != nil {
		t.Fatal(err)
	}
	resetMigrationWarningTestState(t, func(io.Writer) bool { return true })

	var stdout, stderr bytes.Buffer
	status := newRootCmd()
	status.SetOut(&stdout)
	status.SetErr(&stderr)
	status.SetArgs([]string{"--project", root, "status"})
	if err := status.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := stderr.String(); got != "Warning: this local command now uses Brain Planning; prefer `brain plan status`.\n" {
		t.Fatalf("unexpected migration warning: %q", got)
	}

	stdout.Reset()
	stderr.Reset()
	roadmap := newRootCmd()
	roadmap.SetOut(&stdout)
	roadmap.SetErr(&stderr)
	roadmap.SetArgs([]string{"--project", root, "roadmap", "show"})
	if err := roadmap.Execute(); err != nil {
		t.Fatal(err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("warning repeated in one process: %q", stderr.String())
	}
}

func TestSharedPlanningWarningSuppressedForMachineAndNonInteractiveOutput(t *testing.T) {
	root := t.TempDir()
	manager := workspace.New(root)
	if _, err := manager.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := planning.New(manager).CreateBrainstorm("Machine Output"); err != nil {
		t.Fatal(err)
	}
	if _, err := planning.New(manager).EnsureGuidedBrainstormSession("machine-output"); err != nil {
		t.Fatal(err)
	}

	resetMigrationWarningTestState(t, func(io.Writer) bool { return true })
	var stdout, stderr bytes.Buffer
	guide := newRootCmd()
	guide.SetOut(&stdout)
	guide.SetErr(&stderr)
	guide.SetArgs([]string{"--project", root, "guide", "current", "--format", "json"})
	if err := guide.Execute(); err != nil {
		t.Fatal(err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("JSON output received migration warning: %q", stderr.String())
	}

	resetMigrationWarningTestState(t, func(io.Writer) bool { return false })
	stdout.Reset()
	stderr.Reset()
	status := newRootCmd()
	status.SetOut(&stdout)
	status.SetErr(&stderr)
	status.SetArgs([]string{"--project", root, "status"})
	if err := status.Execute(); err != nil {
		t.Fatal(err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("non-interactive output received migration warning: %q", stderr.String())
	}
}

func resetMigrationWarningTestState(t *testing.T, detector func(io.Writer) bool) {
	t.Helper()
	priorDetector := isInteractiveWriter
	isInteractiveWriter = detector
	migrationWarningOnce = sync.Once{}
	t.Cleanup(func() {
		isInteractiveWriter = priorDetector
		migrationWarningOnce = sync.Once{}
	})
}
