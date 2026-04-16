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
	Known         []string `json:"known"`
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
	WorkspaceStatus string   `json:"workspace_status"`
	MigrationStatus string   `json:"migration_status"`
	Missing         []string `json:"missing,omitempty"`
	Broken          []string `json:"broken,omitempty"`
	Guidance        []string `json:"guidance,omitempty"`
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
	meta, err := m.ReadWorkspaceMeta()
	if err != nil {
		return nil, err
	}
	if err := validateWorkspaceMeta(meta); err != nil {
		return nil, fmt.Errorf("validate workspace metadata: %w", err)
	}
	return info, nil
}

func (m *Manager) ReadWorkspaceMeta() (*WorkspaceMeta, error) {
	info, err := m.Resolve()
	if err != nil {
		return nil, err
	}
	return readWorkspaceMetaFile(info.WorkspaceFile)
}

func readWorkspaceMetaFile(path string) (*WorkspaceMeta, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workspace metadata: %w", err)
	}
	var meta WorkspaceMeta
	if err := json.Unmarshal(raw, &meta); err != nil {
		return nil, fmt.Errorf("parse workspace metadata: %w", err)
	}
	return &meta, nil
}

func (m *Manager) ReadMigrationState() (*MigrationState, error) {
	info, err := m.Resolve()
	if err != nil {
		return nil, err
	}
	return readMigrationStateFile(info.MigrationsFile)
}

func readMigrationStateFile(path string) (*MigrationState, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read migration state: %w", err)
	}
	var state MigrationState
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("parse migration state: %w", err)
	}
	return &state, nil
}

func (m *Manager) Doctor() (*DoctorReport, error) {
	info, err := m.Resolve()
	if err != nil {
		return nil, err
	}
	report := &DoctorReport{
		ProjectDir:      info.ProjectDir,
		PlanDir:         info.PlanDir,
		WorkspaceStatus: "adoptable",
		MigrationStatus: "not_initialized",
	}
	if _, err := os.Stat(info.PlanDir); err != nil {
		if os.IsNotExist(err) {
			report.Guidance = guidanceForWorkspaceStatus(report.WorkspaceStatus)
			return report, nil
		}
		return nil, err
	}
	report.Initialized = true

	for _, surface := range info.RequiredSurfaces() {
		if _, err := os.Stat(surface.absPath); err != nil {
			if os.IsNotExist(err) {
				report.Missing = append(report.Missing, surface.Label)
				continue
			}
			return nil, err
		}
	}

	meta, err := readWorkspaceMetaFile(info.WorkspaceFile)
	if err != nil {
		if !contains(report.Missing, "workspace.json") {
			report.Broken = append(report.Broken, "workspace.json")
		}
	} else {
		report.PlanningModel = meta.PlanningModel
		report.SchemaVersion = meta.SchemaVersion
		if err := validateWorkspaceMeta(meta); err != nil {
			report.Broken = append(report.Broken, "workspace.json")
		}
	}

	state, err := readMigrationStateFile(info.MigrationsFile)
	if err != nil {
		if !contains(report.Missing, "migrations.json") {
			report.Broken = append(report.Broken, "migrations.json")
		}
	} else {
		report.MigrationStatus = state.Status
		if err := validateMigrationState(state); err != nil {
			if !contains(report.Broken, "migrations.json") {
				report.Broken = append(report.Broken, "migrations.json")
			}
		}
	}

	report.WorkspaceStatus = classifyDoctorWorkspace(report.Missing, report.Broken)
	if report.WorkspaceStatus == "partial" && report.MigrationStatus == "current" {
		report.MigrationStatus = "partial"
	}
	if report.MigrationStatus == "" {
		switch {
		case contains(report.Broken, "migrations.json"):
			report.MigrationStatus = "broken"
		case isPartialWorkspace(report.Missing):
			report.MigrationStatus = "partial"
		case contains(report.Missing, "migrations.json"):
			report.MigrationStatus = "missing"
		default:
			report.MigrationStatus = "current"
		}
	}
	report.Guidance = guidanceForWorkspaceStatus(report.WorkspaceStatus)
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
	return m.repairWorkspace(info)
}

func (m *Manager) Adopt() (*UpdateResult, error) {
	info, err := m.Resolve()
	if err != nil {
		return nil, err
	}
	return m.repairWorkspace(info)
}

func (m *Manager) repairWorkspace(info *Info) (*UpdateResult, error) {
	result := &UpdateResult{Info: info}
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
	} else {
		updated, err := reconcileWorkspaceMeta(info.WorkspaceFile, now)
		if err != nil {
			return nil, err
		}
		if updated {
			result.Updated = append(result.Updated, rel(info.ProjectDir, info.WorkspaceFile))
		}
	}
	if created, err := ensureMigrationState(info.MigrationsFile, now); err != nil {
		return nil, err
	} else if created {
		result.Created = append(result.Created, rel(info.ProjectDir, info.MigrationsFile))
	} else {
		updated, err := reconcileMigrationState(info.MigrationsFile, now)
		if err != nil {
			return nil, err
		}
		if updated {
			result.Updated = append(result.Updated, rel(info.ProjectDir, info.MigrationsFile))
		}
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
	meta := defaultWorkspaceMeta(now)
	return true, writeJSON(path, meta)
}

func reconcileWorkspaceMeta(path, now string) (bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	var meta WorkspaceMeta
	if err := json.Unmarshal(raw, &meta); err != nil {
		return true, writeJSON(path, defaultWorkspaceMeta(now))
	}
	normalized, changed, err := normalizeWorkspaceMeta(meta, now)
	if err != nil {
		return false, err
	}
	if !changed {
		return false, nil
	}
	return true, writeJSON(path, normalized)
}

func ensureMigrationState(path, now string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	state := defaultMigrationState(now)
	return true, writeJSON(path, state)
}

func reconcileMigrationState(path, now string) (bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	var state MigrationState
	if err := json.Unmarshal(raw, &state); err != nil {
		return true, writeJSON(path, defaultMigrationState(now))
	}
	normalized, changed, err := normalizeMigrationState(state, now)
	if err != nil {
		return false, err
	}
	if !changed {
		return false, nil
	}
	return true, writeJSON(path, normalized)
}

func defaultWorkspaceMeta(now string) WorkspaceMeta {
	return WorkspaceMeta{
		SchemaVersion: CurrentSchemaVersion,
		PlanningModel: PlanningModel,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func defaultMigrationState(now string) MigrationState {
	return MigrationState{
		SchemaVersion: CurrentSchemaVersion,
		Known:         []string{},
		LastRunAt:     now,
		Status:        "current",
	}
}

func normalizeWorkspaceMeta(meta WorkspaceMeta, now string) (WorkspaceMeta, bool, error) {
	if meta.SchemaVersion > CurrentSchemaVersion {
		return WorkspaceMeta{}, false, fmt.Errorf("workspace schema version %d is newer than supported version %d", meta.SchemaVersion, CurrentSchemaVersion)
	}
	if meta.PlanningModel != "" && meta.PlanningModel != PlanningModel {
		return WorkspaceMeta{}, false, fmt.Errorf("workspace planning model %q is not supported by this build", meta.PlanningModel)
	}

	normalized := meta
	changed := false
	if normalized.SchemaVersion == 0 {
		normalized.SchemaVersion = CurrentSchemaVersion
		changed = true
	}
	if normalized.PlanningModel == "" {
		normalized.PlanningModel = PlanningModel
		changed = true
	}
	if normalized.CreatedAt == "" {
		normalized.CreatedAt = now
		changed = true
	}
	if normalized.UpdatedAt == "" {
		normalized.UpdatedAt = now
		changed = true
	}
	if changed {
		normalized.UpdatedAt = now
	}
	return normalized, changed, nil
}

func normalizeMigrationState(state MigrationState, now string) (MigrationState, bool, error) {
	if state.SchemaVersion > CurrentSchemaVersion {
		return MigrationState{}, false, fmt.Errorf("migration schema version %d is newer than supported version %d", state.SchemaVersion, CurrentSchemaVersion)
	}

	normalized := state
	changed := false
	if normalized.SchemaVersion == 0 {
		normalized.SchemaVersion = CurrentSchemaVersion
		changed = true
	}
	if normalized.Known == nil {
		normalized.Known = []string{}
		changed = true
	}
	if normalized.LastRunAt == "" {
		normalized.LastRunAt = now
		changed = true
	}
	if normalized.Status == "" {
		normalized.Status = "current"
		changed = true
	}
	if changed {
		normalized.LastRunAt = now
	}
	return normalized, changed, nil
}

func validateWorkspaceMeta(meta *WorkspaceMeta) error {
	switch {
	case meta.SchemaVersion == 0:
		return fmt.Errorf("schema_version is required")
	case meta.SchemaVersion > CurrentSchemaVersion:
		return fmt.Errorf("schema_version %d is newer than supported %d", meta.SchemaVersion, CurrentSchemaVersion)
	case meta.PlanningModel == "":
		return fmt.Errorf("planning_model is required")
	case meta.PlanningModel != PlanningModel:
		return fmt.Errorf("planning_model %q is not supported", meta.PlanningModel)
	case meta.CreatedAt == "":
		return fmt.Errorf("created_at is required")
	case meta.UpdatedAt == "":
		return fmt.Errorf("updated_at is required")
	default:
		return nil
	}
}

func validateMigrationState(state *MigrationState) error {
	switch {
	case state.SchemaVersion == 0:
		return fmt.Errorf("schema_version is required")
	case state.SchemaVersion > CurrentSchemaVersion:
		return fmt.Errorf("schema_version %d is newer than supported %d", state.SchemaVersion, CurrentSchemaVersion)
	case state.LastRunAt == "":
		return fmt.Errorf("last_run_at is required")
	case state.Status == "":
		return fmt.Errorf("status is required")
	default:
		return nil
	}
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

func classifyDoctorWorkspace(missing, broken []string) string {
	switch {
	case isBrokenWorkspace(broken):
		return "broken"
	case isPartialWorkspace(missing):
		return "partial"
	case len(broken) > 0:
		return "broken"
	case len(missing) > 0:
		return "missing"
	default:
		return "current"
	}
}

func isPartialWorkspace(missing []string) bool {
	return contains(missing, ".meta/") || contains(missing, "workspace.json") || contains(missing, "migrations.json")
}

func isBrokenWorkspace(broken []string) bool {
	return contains(broken, "workspace.json") || contains(broken, "migrations.json")
}

func guidanceForWorkspaceStatus(status string) []string {
	switch status {
	case "adoptable":
		return []string{
			"Run `plan adopt --project .` to create and register the local .plan workspace for this repo.",
		}
	case "partial":
		return []string{
			"Run `plan adopt --project .` to finish adopting the partial .plan workspace.",
			"Use `plan update --project .` after adoption to normalize any remaining plan-managed files.",
		}
	case "missing":
		return []string{
			"Run `plan update --project .` to recreate missing plan-managed surfaces without overwriting user-authored notes.",
		}
	case "broken":
		return []string{
			"Run `plan update --project .` to repair broken plan-managed metadata and restore a current workspace.",
		}
	default:
		return []string{
			"Workspace is current. Use `plan update --project .` only after upgrades or when doctor reports drift.",
		}
	}
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
