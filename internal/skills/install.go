package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Scope string

const (
	ScopeGlobal Scope = "global"
	ScopeLocal  Scope = "local"
	ScopeBoth   Scope = "both"
)

type Installer struct {
	Home string
}

type InstallRequest struct {
	Scope      Scope
	Agents     []string
	ProjectDir string
}

type InstallResult struct {
	Agent  string `json:"agent"`
	Skill  string `json:"skill"`
	Scope  string `json:"scope"`
	Root   string `json:"root"`
	Path   string `json:"path"`
	Method string `json:"method"`
}

type Target struct {
	Agent string `json:"agent"`
	Skill string `json:"skill"`
	Scope string `json:"scope"`
	Root  string `json:"root"`
	Path  string `json:"path"`
}

type TargetStatus struct {
	Target
	Installed   bool      `json:"installed"`
	NeedsRepair bool      `json:"needs_repair"`
	Reason      string    `json:"reason,omitempty"`
	Manifest    *Manifest `json:"manifest,omitempty"`
}

func NewInstaller(home string) *Installer {
	if home == "" {
		if userHome, err := os.UserHomeDir(); err == nil {
			home = userHome
		}
	}
	return &Installer{Home: home}
}

func (i *Installer) Install(req InstallRequest) ([]InstallResult, error) {
	bundles, err := loadBundles()
	if err != nil {
		return nil, err
	}
	bundlesByName := map[string]skillBundle{}
	for _, bundle := range bundles {
		bundlesByName[bundle.Name] = bundle
	}

	targets, err := i.ResolveTargets(req)
	if err != nil {
		return nil, err
	}

	results := make([]InstallResult, 0, len(targets))
	for _, target := range targets {
		bundle, ok := bundlesByName[target.Skill]
		if !ok {
			return nil, fmt.Errorf("missing bundled skill %s", target.Skill)
		}
		if err := os.MkdirAll(target.Root, 0o755); err != nil {
			return nil, fmt.Errorf("create skill root %s: %w", target.Root, err)
		}
		if err := installPath(bundle, target); err != nil {
			return nil, err
		}
		results = append(results, InstallResult{
			Agent:  target.Agent,
			Skill:  target.Skill,
			Scope:  target.Scope,
			Root:   target.Root,
			Path:   target.Path,
			Method: "copy",
		})
	}
	return results, nil
}

func (i *Installer) ResolveTargets(req InstallRequest) ([]Target, error) {
	bundles, err := loadBundles()
	if err != nil {
		return nil, err
	}
	skillNames := make([]string, 0, len(bundles))
	for _, bundle := range bundles {
		skillNames = append(skillNames, bundle.Name)
	}
	sort.Strings(skillNames)

	scope := req.Scope
	if scope == "" {
		scope = ScopeGlobal
	}
	if scope != ScopeGlobal && scope != ScopeLocal && scope != ScopeBoth {
		return nil, fmt.Errorf("unsupported scope: %s", scope)
	}

	agents := normalizeAgents(req.Agents)
	if len(agents) == 0 {
		agents = knownAgents()
	}

	var targets []Target
	if scope == ScopeLocal || scope == ScopeBoth {
		projectDir := req.ProjectDir
		if projectDir == "" {
			projectDir = "."
		}
		resolvedProjectDir, err := resolveProjectDir(projectDir, i.Home)
		if err != nil {
			return nil, err
		}
		for _, agent := range agents {
			root := knownLocalSkillRoot(resolvedProjectDir, agent)
			for _, skill := range skillNames {
				targets = append(targets, Target{
					Agent: agent,
					Skill: skill,
					Scope: string(ScopeLocal),
					Root:  root,
					Path:  filepath.Join(root, skill),
				})
			}
		}
	}
	if scope == ScopeGlobal || scope == ScopeBoth {
		for _, agent := range agents {
			root := knownGlobalSkillRoot(i.Home, agent)
			for _, skill := range skillNames {
				targets = append(targets, Target{
					Agent: agent,
					Skill: skill,
					Scope: string(ScopeGlobal),
					Root:  root,
					Path:  filepath.Join(root, skill),
				})
			}
		}
	}
	return dedupeTargets(targets), nil
}

func (i *Installer) Inspect(req InstallRequest) ([]TargetStatus, error) {
	bundles, err := loadBundles()
	if err != nil {
		return nil, err
	}
	hashesByName := map[string]string{}
	for _, bundle := range bundles {
		hashesByName[bundle.Name] = bundle.Hash
	}

	targets, err := i.ResolveTargets(req)
	if err != nil {
		return nil, err
	}

	statuses := make([]TargetStatus, 0, len(targets))
	for _, target := range targets {
		bundleHash, ok := hashesByName[target.Skill]
		if !ok {
			return nil, fmt.Errorf("missing bundled skill %s", target.Skill)
		}
		status, err := inspectTarget(target, bundleHash)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}

func installPath(bundle skillBundle, target Target) error {
	if err := os.RemoveAll(target.Path); err != nil {
		return fmt.Errorf("clear target %s: %w", target.Path, err)
	}
	if err := copyBundle(bundle, target.Path); err != nil {
		return fmt.Errorf("copy skill to %s: %w", target.Path, err)
	}
	if err := writeManifest(target.Path, target.Skill, target.Agent, target.Scope, bundle.Hash); err != nil {
		return fmt.Errorf("write skill manifest %s: %w", manifestPath(target.Path), err)
	}
	return nil
}

func inspectTarget(target Target, bundleHash string) (TargetStatus, error) {
	status := TargetStatus{Target: target}

	info, err := os.Lstat(target.Path)
	if os.IsNotExist(err) {
		return status, nil
	}
	if err != nil {
		return status, fmt.Errorf("inspect target %s: %w", target.Path, err)
	}

	status.Installed = true
	if info.Mode()&os.ModeSymlink != 0 {
		status.NeedsRepair = true
		status.Reason = "legacy_symlink"
		return status, nil
	}
	if !info.IsDir() {
		status.NeedsRepair = true
		status.Reason = "invalid_target"
		return status, nil
	}
	if _, err := os.Stat(filepath.Join(target.Path, "SKILL.md")); err != nil {
		status.NeedsRepair = true
		status.Reason = "missing_skill_file"
		return status, nil
	}

	manifest, err := readManifest(target.Path)
	if os.IsNotExist(err) {
		status.NeedsRepair = true
		status.Reason = "legacy_install"
		return status, nil
	}
	if err != nil {
		status.NeedsRepair = true
		status.Reason = "invalid_manifest"
		return status, nil
	}
	status.Manifest = manifest
	if manifest.BundleHash != bundleHash {
		status.NeedsRepair = true
		status.Reason = "stale_bundle"
	}
	return status, nil
}

func knownAgents() []string {
	return []string{"codex", "claude", "copilot", "openclaw", "pi", "ai"}
}

func knownGlobalSkillRoot(home, agent string) string {
	switch agent {
	case "codex":
		return filepath.Join(home, ".codex", "skills")
	case "claude":
		return filepath.Join(home, ".claude", "skills")
	case "copilot":
		return filepath.Join(home, ".copilot", "skills")
	case "openclaw":
		return filepath.Join(home, ".openclaw", "skills")
	case "pi":
		return filepath.Join(home, ".pi", "agent", "skills")
	case "ai":
		return filepath.Join(home, ".ai", "skills")
	default:
		return filepath.Join(home, "."+agent, "skills")
	}
}

func knownLocalSkillRoot(projectDir, agent string) string {
	switch agent {
	case "copilot":
		return filepath.Join(projectDir, ".github", "skills")
	case "pi":
		return filepath.Join(projectDir, ".pi", "skills")
	default:
		return filepath.Join(projectDir, "."+agent, "skills")
	}
}

func normalizeAgents(agents []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, agent := range agents {
		agent = canonicalAgentName(agent)
		if agent == "" {
			continue
		}
		if _, ok := seen[agent]; ok {
			continue
		}
		seen[agent] = struct{}{}
		out = append(out, agent)
	}
	sort.Strings(out)
	return out
}

func canonicalAgentName(agent string) string {
	agent = strings.TrimSpace(strings.ToLower(agent))
	switch agent {
	case "github-copilot", "copilot-cli", "copilot-chat":
		return "copilot"
	case "pi.dev", "pi-dev":
		return "pi"
	default:
		return agent
	}
}

func dedupeTargets(targets []Target) []Target {
	seen := map[string]struct{}{}
	var out []Target
	for _, target := range targets {
		if _, ok := seen[target.Path]; ok {
			continue
		}
		seen[target.Path] = struct{}{}
		out = append(out, target)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Scope == out[j].Scope {
			if out[i].Agent == out[j].Agent {
				return out[i].Skill < out[j].Skill
			}
			return out[i].Agent < out[j].Agent
		}
		return out[i].Scope < out[j].Scope
	})
	return out
}

func expandHome(path, home string) string {
	if path == "" || path[0] != '~' {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~/"))
}

func resolveProjectDir(path, home string) (string, error) {
	path = filepath.Clean(expandHome(path, home))
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve project dir %s: %w", path, err)
	}
	return abs, nil
}
