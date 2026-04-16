package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"plan/internal/templates"
)

const (
	CurrentSchemaVersion = 1
	PlanningModel        = "epic_spec_story_v1"
)

type WorkspaceMeta struct {
	SchemaVersion int    `json:"schema_version"`
	PlanningModel string `json:"planning_model"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type MigrationState struct {
	SchemaVersion int      `json:"schema_version"`
	Known         []string `json:"known,omitempty"`
	LastRunAt     string   `json:"last_run_at,omitempty"`
	Status        string   `json:"status,omitempty"`
}

type Surface struct {
	Label string `json:"label"`
	Path  string `json:"path"`
	Kind  string `json:"kind"`
	Type  string `json:"type"`

	absPath string
}

type Contract struct {
	UserAuthored []Surface `json:"user_authored"`
	ToolManaged  []Surface `json:"tool_managed"`
}

type Info struct {
	ProjectDir     string
	ProjectName    string
	PlanDir        string
	ProjectFile    string
	RoadmapFile    string
	BrainstormsDir string
	EpicsDir       string
	SpecsDir       string
	StoriesDir     string
	MetaDir        string
	WorkspaceFile  string
	MigrationsFile string
}

type InitResult struct {
	Info    *Info
	Created []string
}

type UpdateResult struct {
	Info    *Info
	Created []string
	Updated []string
}

type DoctorReport struct {
	ProjectDir      string   `json:"project_dir"`
	PlanDir         string   `json:"plan_dir"`
	Initialized     bool     `json:"initialized"`
	PlanningModel   string   `json:"planning_model,omitempty"`
	SchemaVersion   int      `json:"schema_version,omitempty"`
	MigrationStatus string   `json:"migration_status"`
	Missing         []string `json:"missing,omitempty"`
}

type Manager struct {
	projectDir string
}

func New(projectDir string) *Manager {
	return &Manager{projectDir: projectDir}
}

func (m *Manager) Resolve() (*Info, error) {
	root := m.projectDir
	if root == "" {
		root = "."
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	name := filepath.Base(abs)
	planDir := filepath.Join(abs, ".plan")
	metaDir := filepath.Join(planDir, ".meta")
	return &Info{
		ProjectDir:     abs,
		ProjectName:    name,
		PlanDir:        planDir,
		ProjectFile:    filepath.Join(planDir, "PROJECT.md"),
		RoadmapFile:    filepath.Join(planDir, "ROADMAP.md"),
		BrainstormsDir: filepath.Join(planDir, "brainstorms"),
		EpicsDir:       filepath.Join(planDir, "epics"),
		SpecsDir:       filepath.Join(planDir, "specs"),
		StoriesDir:     filepath.Join(planDir, "stories"),
		MetaDir:        metaDir,
		WorkspaceFile:  filepath.Join(metaDir, "workspace.json"),
		MigrationsFile: filepath.Join(metaDir, "migrations.json"),
	}, nil
}

func (i *Info) Contract() Contract {
	return Contract{
		UserAuthored: i.UserAuthoredSurfaces(),
		ToolManaged:  i.ToolManagedSurfaces(),
	}
}

func (i *Info) UserAuthoredSurfaces() []Surface {
	return []Surface{
		i.newSurface("PROJECT.md", i.ProjectFile, "user_authored", "file"),
		i.newSurface("ROADMAP.md", i.RoadmapFile, "user_authored", "file"),
		i.newSurface("brainstorms/", i.BrainstormsDir, "user_authored", "dir"),
		i.newSurface("epics/", i.EpicsDir, "user_authored", "dir"),
		i.newSurface("specs/", i.SpecsDir, "user_authored", "dir"),
		i.newSurface("stories/", i.StoriesDir, "user_authored", "dir"),
	}
}

func (i *Info) ToolManagedSurfaces() []Surface {
	return []Surface{
		i.newSurface(".meta/", i.MetaDir, "tool_managed", "dir"),
		i.newSurface("workspace.json", i.WorkspaceFile, "tool_managed", "file"),
		i.newSurface("migrations.json", i.MigrationsFile, "tool_managed", "file"),
	}
}

func (i *Info) RequiredSurfaces() []Surface {
	surfaces := make([]Surface, 0, len(i.UserAuthoredSurfaces())+len(i.ToolManagedSurfaces()))
	surfaces = append(surfaces, i.UserAuthoredSurfaces()...)
	surfaces = append(surfaces, i.ToolManagedSurfaces()...)
	return surfaces
}

func (i *Info) newSurface(label, path, kind, surfaceType string) Surface {
	return Surface{
		Label:   label,
		Path:    rel(i.ProjectDir, path),
		Kind:    kind,
		Type:    surfaceType,
		absPath: path,
	}
}

func (m *Manager) Init() (*InitResult, error) {
	info, err := m.Resolve()
	if err != nil {
		return nil, err
	}

	result := &InitResult{Info: info}
	for _, dir := range append([]string{info.PlanDir}, info.directoryPaths()...) {
		created, err := ensureDir(dir)
		if err != nil {
			return nil, err
		}
		if created {
			result.Created = append(result.Created, rel(info.ProjectDir, dir))
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if created, err := ensureTemplateFile(info.ProjectFile, "PROJECT.md", map[string]any{
		"ProjectName": info.ProjectName,
		"Now":         now,
	}); err != nil {
		return nil, err
	} else if created {
		result.Created = append(result.Created, rel(info.ProjectDir, info.ProjectFile))
	}
	if created, err := ensureTemplateFile(info.RoadmapFile, "ROADMAP.md", map[string]any{
		"ProjectName": info.ProjectName,
		"Now":         now,
	}); err != nil {
		return nil, err
	} else if created {
		result.Created = append(result.Created, rel(info.ProjectDir, info.RoadmapFile))
	}
	if created, err := ensureWorkspaceMeta(info.WorkspaceFile, now); err != nil {
		return nil, err
	} else if created {
		result.Created = append(result.Created, rel(info.ProjectDir, info.WorkspaceFile))
	}
	if created, err := ensureMigrationState(info.MigrationsFile, now); err != nil {
		return nil, err
	} else if created {
		result.Created = append(result.Created, rel(info.ProjectDir, info.MigrationsFile))
	}

	return result, nil
}

func (i *Info) directoryPaths() []string {
	var dirs []string
	for _, surface := range i.RequiredSurfaces() {
		if surface.Type == "dir" {
			dirs = append(dirs, surface.absPath)
		}
	}
	return dirs
}

func (m *Manager) EnsureInitialized() (*Info, error) {
	info, err := m.Resolve()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(info.PlanDir); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("plan workspace not initialized at %s (run `plan init --project .`)", info.ProjectDir)
		}
		return nil, err
	}
	if _, err := m.ReadWorkspaceMeta(); err != nil {
		return nil, err
	}
	return info, nil
}

func (m *Manager) ReadWorkspaceMeta() (*WorkspaceMeta, error) {
	info, err := m.Resolve()
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(info.WorkspaceFile)
	if err != nil {
		return nil, fmt.Errorf("read workspace metadata: %w", err)
	}
	var meta WorkspaceMeta
	if err := json.Unmarshal(raw, &meta); err != nil {
		return nil, fmt.Errorf("parse workspace metadata: %w", err)
	}
	return &meta, nil
}

func (m *Manager) Doctor() (*DoctorReport, error) {
	info, err := m.Resolve()
	if err != nil {
		return nil, err
	}
	report := &DoctorReport{
		ProjectDir:      info.ProjectDir,
		PlanDir:         info.PlanDir,
		MigrationStatus: "not_initialized",
	}
	if _, err := os.Stat(info.PlanDir); err != nil {
		if os.IsNotExist(err) {
			return report, nil
		}
		return nil, err
	}
	report.Initialized = true

	for _, path := range []struct {
		label string
		path  string
	}{
		{"PROJECT.md", info.ProjectFile},
		{"ROADMAP.md", info.RoadmapFile},
		{"brainstorms/", info.BrainstormsDir},
		{"epics/", info.EpicsDir},
		{"specs/", info.SpecsDir},
		{"stories/", info.StoriesDir},
		{"workspace.json", info.WorkspaceFile},
		{"migrations.json", info.MigrationsFile},
	} {
		if _, err := os.Stat(path.path); err != nil {
			if os.IsNotExist(err) {
				report.Missing = append(report.Missing, path.label)
				continue
			}
			return nil, err
		}
	}

	meta, err := m.ReadWorkspaceMeta()
	if err != nil {
		report.MigrationStatus = "broken"
		return report, nil
	}
	report.PlanningModel = meta.PlanningModel
	report.SchemaVersion = meta.SchemaVersion

	switch {
	case len(report.Missing) > 0:
		report.MigrationStatus = "missing"
	default:
		report.MigrationStatus = "current"
	}
	return report, nil
}

func (m *Manager) Update() (*UpdateResult, error) {
	info, err := m.Resolve()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(info.PlanDir); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("plan workspace not initialized at %s (run `plan init --project .`)", info.ProjectDir)
		}
		return nil, err
	}

	result := &UpdateResult{Info: info}
	for _, dir := range []string{
		info.BrainstormsDir,
		info.EpicsDir,
		info.SpecsDir,
		info.StoriesDir,
		info.MetaDir,
	} {
		created, err := ensureDir(dir)
		if err != nil {
			return nil, err
		}
		if created {
			result.Created = append(result.Created, rel(info.ProjectDir, dir))
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if created, err := ensureTemplateFile(info.ProjectFile, "PROJECT.md", map[string]any{
		"ProjectName": info.ProjectName,
		"Now":         now,
	}); err != nil {
		return nil, err
	} else if created {
		result.Created = append(result.Created, rel(info.ProjectDir, info.ProjectFile))
	}
	if created, err := ensureTemplateFile(info.RoadmapFile, "ROADMAP.md", map[string]any{
		"ProjectName": info.ProjectName,
		"Now":         now,
	}); err != nil {
		return nil, err
	} else if created {
		result.Created = append(result.Created, rel(info.ProjectDir, info.RoadmapFile))
	}
	if created, err := ensureWorkspaceMeta(info.WorkspaceFile, now); err != nil {
		return nil, err
	} else if created {
		result.Created = append(result.Created, rel(info.ProjectDir, info.WorkspaceFile))
	} else {
		if err := touchWorkspaceMeta(info.WorkspaceFile, now); err != nil {
			return nil, err
		}
		result.Updated = append(result.Updated, rel(info.ProjectDir, info.WorkspaceFile))
	}
	if created, err := ensureMigrationState(info.MigrationsFile, now); err != nil {
		return nil, err
	} else if created {
		result.Created = append(result.Created, rel(info.ProjectDir, info.MigrationsFile))
	} else {
		if err := touchMigrationState(info.MigrationsFile, now); err != nil {
			return nil, err
		}
		result.Updated = append(result.Updated, rel(info.ProjectDir, info.MigrationsFile))
	}

	return result, nil
}

func ensureDir(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return false, err
	}
	return true, nil
}

func ensureTemplateFile(path, templateName string, data any) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	body, err := templates.Render(templateName, data)
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func ensureWorkspaceMeta(path, now string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	meta := WorkspaceMeta{
		SchemaVersion: CurrentSchemaVersion,
		PlanningModel: PlanningModel,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	return true, writeJSON(path, meta)
}

func touchWorkspaceMeta(path, now string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var meta WorkspaceMeta
	if err := json.Unmarshal(raw, &meta); err != nil {
		return err
	}
	if meta.SchemaVersion == 0 {
		meta.SchemaVersion = CurrentSchemaVersion
	}
	if meta.PlanningModel == "" {
		meta.PlanningModel = PlanningModel
	}
	if meta.CreatedAt == "" {
		meta.CreatedAt = now
	}
	meta.UpdatedAt = now
	return writeJSON(path, meta)
}

func ensureMigrationState(path, now string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	state := MigrationState{
		SchemaVersion: CurrentSchemaVersion,
		Known:         []string{},
		LastRunAt:     now,
		Status:        "current",
	}
	return true, writeJSON(path, state)
}

func touchMigrationState(path, now string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var state MigrationState
	if err := json.Unmarshal(raw, &state); err != nil {
		return err
	}
	if state.SchemaVersion == 0 {
		state.SchemaVersion = CurrentSchemaVersion
	}
	if state.Known == nil {
		state.Known = []string{}
	}
	state.LastRunAt = now
	state.Status = "current"
	return writeJSON(path, state)
}

func writeJSON(path string, v any) error {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o644)
}

func rel(root, path string) string {
	r, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(r)
}
