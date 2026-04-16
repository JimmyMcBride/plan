package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesPlanWorkspace(t *testing.T) {
	root := t.TempDir()
	manager := New(root)

	result, err := manager.Init()
	if err != nil {
		t.Fatal(err)
	}
	if result.Info == nil {
		t.Fatal("expected workspace info")
	}

	for _, path := range []string{
		filepath.Join(root, ".plan", "brainstorms"),
		filepath.Join(root, ".plan", "epics"),
		filepath.Join(root, ".plan", "specs"),
		filepath.Join(root, ".plan", "stories"),
		filepath.Join(root, ".plan", ".meta"),
		filepath.Join(root, ".plan", "PROJECT.md"),
		filepath.Join(root, ".plan", "ROADMAP.md"),
		filepath.Join(root, ".plan", ".meta", "workspace.json"),
		filepath.Join(root, ".plan", ".meta", "migrations.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s: %v", path, err)
		}
	}
}

func TestWorkspaceContractSeparatesUserAuthoredAndToolManagedSurfaces(t *testing.T) {
	root := t.TempDir()
	manager := New(root)

	info, err := manager.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	contract := info.Contract()

	if len(contract.UserAuthored) != 6 {
		t.Fatalf("unexpected user-authored surface count: %d", len(contract.UserAuthored))
	}
	if len(contract.ToolManaged) != 3 {
		t.Fatalf("unexpected tool-managed surface count: %d", len(contract.ToolManaged))
	}

	for _, surface := range contract.UserAuthored {
		if surface.Kind != "user_authored" {
			t.Fatalf("unexpected user-authored surface kind: %+v", surface)
		}
		if !strings.HasPrefix(surface.Path, ".plan/") {
			t.Fatalf("unexpected user-authored surface path: %+v", surface)
		}
	}

	for _, surface := range contract.ToolManaged {
		if surface.Kind != "tool_managed" {
			t.Fatalf("unexpected tool-managed surface kind: %+v", surface)
		}
		if !strings.HasPrefix(surface.Path, ".plan/") {
			t.Fatalf("unexpected tool-managed surface path: %+v", surface)
		}
	}
}

func TestDoctorReportsCurrentAfterInit(t *testing.T) {
	root := t.TempDir()
	manager := New(root)
	if _, err := manager.Init(); err != nil {
		t.Fatal(err)
	}

	report, err := manager.Doctor()
	if err != nil {
		t.Fatal(err)
	}
	if !report.Initialized {
		t.Fatal("expected initialized workspace")
	}
	if report.MigrationStatus != "current" {
		t.Fatalf("unexpected migration status: %+v", report)
	}
	if report.PlanningModel != PlanningModel {
		t.Fatalf("unexpected planning model: %+v", report)
	}
}
