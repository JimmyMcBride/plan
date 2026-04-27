package skills

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveTargetsGlobalAndLocal(t *testing.T) {
	registerTestBundles(t)

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
		filepath.Join(home, ".codex", "skills", "plan"):                true,
		filepath.Join(home, ".codex", "skills", "plan-execute"):        true,
		filepath.Join(home, ".copilot", "skills", "plan"):              true,
		filepath.Join(home, ".copilot", "skills", "plan-execute"):      true,
		filepath.Join(home, ".pi", "agent", "skills", "plan"):          true,
		filepath.Join(home, ".pi", "agent", "skills", "plan-execute"):  true,
		filepath.Join(home, ".zed", "skills", "plan"):                  true,
		filepath.Join(home, ".zed", "skills", "plan-execute"):          true,
		filepath.Join(projectDir, ".codex", "skills", "plan"):          true,
		filepath.Join(projectDir, ".codex", "skills", "plan-execute"):  true,
		filepath.Join(projectDir, ".github", "skills", "plan"):         true,
		filepath.Join(projectDir, ".github", "skills", "plan-execute"): true,
		filepath.Join(projectDir, ".pi", "skills", "plan"):             true,
		filepath.Join(projectDir, ".pi", "skills", "plan-execute"):     true,
		filepath.Join(projectDir, ".zed", "skills", "plan"):            true,
		filepath.Join(projectDir, ".zed", "skills", "plan-execute"):    true,
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
	bundleHashes := registerTestBundles(t)

	home := t.TempDir()
	installer := NewInstaller(home)
	results, err := installer.Install(InstallRequest{
		Scope:  ScopeGlobal,
		Agents: []string{"codex"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, skill := range []string{"plan", "plan-execute"} {
		skillDir := filepath.Join(home, ".codex", "skills", skill)
		if _, err := os.Stat(filepath.Join(skillDir, "SKILL.md")); err != nil {
			t.Fatalf("expected %s skill file: %v", skill, err)
		}
		manifest, err := readManifest(skillDir)
		if err != nil {
			t.Fatalf("expected %s manifest: %v", skill, err)
		}
		if manifest.Skill != skill || manifest.BundleHash != bundleHashes[skill] {
			t.Fatalf("unexpected %s manifest: %+v", skill, manifest)
		}
	}
}

func TestResolveTargetsLocalScopeUsesAbsoluteProjectDir(t *testing.T) {
	registerTestBundles(t)

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
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
	want := map[string]bool{
		filepath.Join(projectDir, ".codex", "skills", "plan"):         true,
		filepath.Join(projectDir, ".codex", "skills", "plan-execute"): true,
	}
	for _, target := range targets {
		if !filepath.IsAbs(target.Root) || !filepath.IsAbs(target.Path) {
			t.Fatalf("expected absolute local target paths: %+v", target)
		}
		if !want[target.Path] {
			t.Fatalf("unexpected resolved target path: %+v", target)
		}
	}
}

func TestInspectFlagsLegacyAndStaleInstalls(t *testing.T) {
	bundleHashes := registerTestBundles(t)

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
			if status.Manifest == nil || status.Manifest.BundleHash == bundleHashes["plan"] {
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
	bundleHashes := registerTestBundles(t)

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
	if manifest.Skill != "plan" || manifest.BundleHash != bundleHashes["plan"] {
		t.Fatalf("unexpected repaired manifest: %+v", manifest)
	}
	if _, err := os.Stat(filepath.Join(home, ".codex", "skills", "plan-execute", "SKILL.md")); err != nil {
		t.Fatalf("expected plan-execute to install alongside plan: %v", err)
	}
}

func TestSourceTreeBundlesInstallPlanAndExecuteSkills(t *testing.T) {
	RegisterBundles(nil)

	home := t.TempDir()
	installer := NewInstaller(home)
	if _, err := installer.Install(InstallRequest{
		Scope:  ScopeGlobal,
		Agents: []string{"codex"},
	}); err != nil {
		t.Fatal(err)
	}

	skillDir := filepath.Join(home, ".codex", "skills", "plan")
	for _, rel := range []string{
		"SKILL.md",
		filepath.Join("agents", "openai.yaml"),
		filepath.Join("agents", "gpt-style.yaml"),
		filepath.Join("agents", "reasoning.yaml"),
	} {
		if _, err := os.Stat(filepath.Join(skillDir, rel)); err != nil {
			t.Fatalf("expected installed skill file %s: %v", rel, err)
		}
	}
	if _, err := os.Stat(filepath.Join(home, ".codex", "skills", "plan-execute", "SKILL.md")); err != nil {
		t.Fatalf("expected installed plan-execute skill file: %v", err)
	}
}

func registerTestBundles(t *testing.T) map[string]string {
	t.Helper()

	root := t.TempDir()
	bundleFSes := map[string]fs.FS{}
	for _, skill := range []string{"plan", "plan-execute"} {
		bundleDir := filepath.Join(root, skill)
		if err := os.MkdirAll(filepath.Join(bundleDir, "agents"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(bundleDir, "SKILL.md"), []byte("skill: "+skill), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(bundleDir, "agents", "openai.yaml"), []byte("name: "+skill), 0o644); err != nil {
			t.Fatal(err)
		}
		bundleFSes[skill] = os.DirFS(bundleDir)
	}

	RegisterBundles(bundleFSes)
	t.Cleanup(func() {
		RegisterBundles(nil)
	})

	loaded, err := loadBundles()
	if err != nil {
		t.Fatal(err)
	}
	hashes := map[string]string{}
	for _, bundle := range loaded {
		hashes[bundle.Name] = bundle.Hash
	}
	return hashes
}
