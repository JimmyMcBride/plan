package skills

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveTargetsGlobalAndLocal(t *testing.T) {
	home := t.TempDir()
	projectDir := t.TempDir()
	installer := NewInstaller(home)
	targets, err := installer.ResolveTargets(InstallRequest{
		Scope:      ScopeBoth,
		Agents:     []string{"codex", "copilot", "pi", "zed"},
		ProjectDir: projectDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		filepath.Join(home, ".codex", "skills", "plan"):        true,
		filepath.Join(home, ".copilot", "skills", "plan"):      true,
		filepath.Join(home, ".pi", "agent", "skills", "plan"):  true,
		filepath.Join(home, ".zed", "skills", "plan"):          true,
		filepath.Join(projectDir, ".codex", "skills", "plan"):  true,
		filepath.Join(projectDir, ".github", "skills", "plan"): true,
		filepath.Join(projectDir, ".pi", "skills", "plan"):     true,
		filepath.Join(projectDir, ".zed", "skills", "plan"):    true,
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

func TestResolveTargetsLocalScopeUsesAbsoluteProjectDir(t *testing.T) {
	installer := NewInstaller("/home/tester")
	projectDir := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(prevWD)
	}()

	targets, err := installer.ResolveTargets(InstallRequest{
		Scope:      ScopeLocal,
		Agents:     []string{"codex"},
		ProjectDir: ".",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if !filepath.IsAbs(targets[0].Root) || !filepath.IsAbs(targets[0].Path) {
		t.Fatalf("expected absolute local target paths: %+v", targets[0])
	}
	if targets[0].Path != filepath.Join(projectDir, ".codex", "skills", "plan") {
		t.Fatalf("unexpected resolved target path: %+v", targets[0])
	}
}

func TestInspectFlagsLegacyAndStaleInstalls(t *testing.T) {
	bundleHash := registerTestBundle(t)

	home := t.TempDir()
	installer := NewInstaller(home)

	legacyDir := filepath.Join(home, ".claude", "skills", "plan")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, "SKILL.md"), []byte("legacy"), 0o644); err != nil {
		t.Fatal(err)
	}

	staleDir := filepath.Join(home, ".codex", "skills", "plan")
	if err := os.MkdirAll(staleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staleDir, "SKILL.md"), []byte("skill"), 0o644); err != nil {
		t.Fatal(err)
	}
	raw, err := json.MarshalIndent(Manifest{
		SchemaVersion: manifestSchemaVersion,
		PlanVersion:   "v0.0.1",
		PlanCommit:    "deadbeef",
		BundleHash:    "stale-hash",
		InstalledAt:   "2026-01-01T00:00:00Z",
		Agent:         "codex",
		Scope:         string(ScopeGlobal),
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(manifestPath(staleDir), raw, 0o644); err != nil {
		t.Fatal(err)
	}

	statuses, err := installer.Inspect(InstallRequest{
		Scope:  ScopeGlobal,
		Agents: []string{"codex", "claude"},
	})
	if err != nil {
		t.Fatal(err)
	}

	reasons := map[string]string{}
	for _, status := range statuses {
		if status.Path == staleDir {
			if status.Manifest == nil || status.Manifest.BundleHash == bundleHash {
				t.Fatalf("expected stale manifest to be inspected: %+v", status)
			}
		}
		reasons[status.Path] = status.Reason
	}
	if reasons[legacyDir] != "legacy_install" {
		t.Fatalf("expected legacy install reason, got %q", reasons[legacyDir])
	}
	if reasons[staleDir] != "stale_bundle" {
		t.Fatalf("expected stale bundle reason, got %q", reasons[staleDir])
	}
}

func TestInstallRepairsLegacyInstallCleanly(t *testing.T) {
	bundleHash := registerTestBundle(t)

	home := t.TempDir()
	installer := NewInstaller(home)
	skillDir := filepath.Join(home, ".codex", "skills", "plan")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "old.txt"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := installer.Install(InstallRequest{
		Scope:  ScopeGlobal,
		Agents: []string{"codex"},
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(skillDir, "old.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected stale file to be removed, got %v", err)
	}
	manifest, err := readManifest(skillDir)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.BundleHash != bundleHash {
		t.Fatalf("unexpected repaired manifest: %+v", manifest)
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
