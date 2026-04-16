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

func TestDoctorReportsMissingWorkspaceSurfaces(t *testing.T) {
	root := t.TempDir()
	manager := New(root)
	if _, err := manager.Init(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, ".plan", "ROADMAP.md")); err != nil {
		t.Fatal(err)
	}

	report, err := manager.Doctor()
	if err != nil {
		t.Fatal(err)
	}
	if report.WorkspaceStatus != "missing" {
		t.Fatalf("unexpected workspace status: %+v", report)
	}
	if !contains(report.Missing, "ROADMAP.md") {
		t.Fatalf("expected missing roadmap in report: %+v", report)
	}
	if report.MigrationStatus != "current" {
		t.Fatalf("unexpected migration status: %+v", report)
	}
}

func TestDoctorReportsBrokenToolManagedState(t *testing.T) {
	root := t.TempDir()
	manager := New(root)
	if _, err := manager.Init(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".plan", ".meta", "workspace.json"), []byte("{broken"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := manager.Doctor()
	if err != nil {
		t.Fatal(err)
	}
	if report.WorkspaceStatus != "broken" {
		t.Fatalf("unexpected workspace status: %+v", report)
	}
	if !contains(report.Broken, "workspace.json") {
		t.Fatalf("expected broken workspace metadata in report: %+v", report)
	}
}

func TestUpdateRepairsToolManagedStateWithoutTouchingUserNotes(t *testing.T) {
	root := t.TempDir()
	manager := New(root)
	if _, err := manager.Init(); err != nil {
		t.Fatal(err)
	}

	projectPath := filepath.Join(root, ".plan", "PROJECT.md")
	const customProject = "# custom project\n"
	if err := os.WriteFile(projectPath, []byte(customProject), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".plan", ".meta", "workspace.json"), []byte("{broken"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, ".plan", ".meta", "migrations.json")); err != nil {
		t.Fatal(err)
	}

	result, err := manager.Update()
	if err != nil {
		t.Fatal(err)
	}
	if !contains(result.Updated, ".plan/.meta/workspace.json") {
		t.Fatalf("expected workspace metadata repair: %+v", result)
	}
	if !contains(result.Created, ".plan/.meta/migrations.json") {
		t.Fatalf("expected migration state recreation: %+v", result)
	}

	raw, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != customProject {
		t.Fatalf("project note was modified:\n%s", raw)
	}

	report, err := manager.Doctor()
	if err != nil {
		t.Fatal(err)
	}
	if report.WorkspaceStatus != "current" || report.MigrationStatus != "current" {
		t.Fatalf("expected repaired workspace to be current: %+v", report)
	}
}

func TestUpdateIsIdempotentWhenWorkspaceIsCurrent(t *testing.T) {
	root := t.TempDir()
	manager := New(root)
	if _, err := manager.Init(); err != nil {
		t.Fatal(err)
	}

	first, err := manager.Update()
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Created) != 0 || len(first.Updated) != 0 {
		t.Fatalf("expected no changes for current workspace: %+v", first)
	}

	second, err := manager.Update()
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Created) != 0 || len(second.Updated) != 0 {
		t.Fatalf("expected second update to stay idempotent: %+v", second)
	}
}
