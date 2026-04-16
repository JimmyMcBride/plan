package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveTargetsGlobalAndLocal(t *testing.T) {
	installer := NewInstaller("/home/tester")
	targets, err := installer.ResolveTargets(InstallRequest{
		Scope:      ScopeBoth,
		Agents:     []string{"codex", "copilot", "pi", "zed"},
		ProjectDir: "/tmp/project",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		filepath.Join("/home/tester", ".codex", "skills", "plan"):       true,
		filepath.Join("/home/tester", ".copilot", "skills", "plan"):     true,
		filepath.Join("/home/tester", ".pi", "agent", "skills", "plan"): true,
		filepath.Join("/home/tester", ".zed", "skills", "plan"):         true,
		filepath.Join("/tmp/project", ".codex", "skills", "plan"):       true,
		filepath.Join("/tmp/project", ".github", "skills", "plan"):      true,
		filepath.Join("/tmp/project", ".pi", "skills", "plan"):          true,
		filepath.Join("/tmp/project", ".zed", "skills", "plan"):         true,
	}
	if len(targets) != len(want) {
		t.Fatalf("expected %d targets, got %d", len(want), len(targets))
	}
	for _, target := range targets {
		if !want[target.Path] {
			t.Fatalf("unexpected target: %+v", target)
		}
	}
}

func TestInstallCopiesSkillBundleAndWritesManifest(t *testing.T) {
	bundleHash := registerTestBundle(t)

	home := t.TempDir()
	installer := NewInstaller(home)
	results, err := installer.Install(InstallRequest{
		Scope:  ScopeGlobal,
		Agents: []string{"codex"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	skillDir := filepath.Join(home, ".codex", "skills", "plan")
	if _, err := os.Stat(filepath.Join(skillDir, "SKILL.md")); err != nil {
		t.Fatalf("expected skill file: %v", err)
	}
	manifest, err := readManifest(skillDir)
	if err != nil {
		t.Fatalf("expected manifest: %v", err)
	}
	if manifest.BundleHash != bundleHash {
		t.Fatalf("unexpected bundle hash: %+v", manifest)
	}
}

func registerTestBundle(t *testing.T) string {
	t.Helper()

	bundleDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(bundleDir, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundleDir, "SKILL.md"), []byte("skill"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundleDir, "agents", "openai.yaml"), []byte("name: plan"), 0o644); err != nil {
		t.Fatal(err)
	}

	RegisterBundle(os.DirFS(bundleDir))
	t.Cleanup(func() {
		RegisterBundle(nil)
	})

	bundle, err := loadBundle()
	if err != nil {
		t.Fatal(err)
	}
	return bundle.Hash
}
