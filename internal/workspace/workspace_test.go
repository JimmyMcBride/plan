package workspace

import (
	"os"
	"path/filepath"
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
