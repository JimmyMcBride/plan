package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"plan/internal/notes"
	"plan/internal/templates"
)

const (
	CurrentSchemaVersion = 3
	PlanningModel        = "spec_first_v1"
	LegacyPlanningModel  = "epic_spec_story_v1"
)

type StoryBackend string

const (
	StoryBackendLocal  StoryBackend = "local"
	StoryBackendGitHub StoryBackend = "github"
)

type SourceOfTruthMode string

const (
	SourceOfTruthLocal  SourceOfTruthMode = "local"
	SourceOfTruthGitHub SourceOfTruthMode = "github"
	SourceOfTruthHybrid SourceOfTruthMode = "hybrid"
)

type WorkspaceMeta struct {
	SchemaVersion int               `json:"schema_version"`
	PlanningModel string            `json:"planning_model"`
	SourceMode    SourceOfTruthMode `json:"source_mode,omitempty"`
	StoryBackend  StoryBackend      `json:"story_backend"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     string            `json:"updated_at"`
}

type GitHubState struct {
	Repo             string                                 `json:"repo,omitempty"`
	RepoURL          string                                 `json:"repo_url,omitempty"`
	DefaultBranch    string                                 `json:"default_branch,omitempty"`
	LastEnabledAt    string                                 `json:"last_enabled_at,omitempty"`
	LastUpdatedAt    string                                 `json:"last_updated_at,omitempty"`
	LastReconciled   string                                 `json:"last_reconciled_at,omitempty"`
	Stories          map[string]GitHubStoryRecord           `json:"stories"`
	Planning         map[string]GitHubPlanningRecord        `json:"planning"`
	ProjectDecisions map[string]GitHubProjectDecisionRecord `json:"project_decisions,omitempty"`
}

type GitHubPlanningRecord struct {
	Slug              string   `json:"slug"`
	Kind              string   `json:"kind"`
	Title             string   `json:"title"`
	IssueNumber       int      `json:"issue_number,omitempty"`
	IssueURL          string   `json:"issue_url,omitempty"`
	RemoteState       string   `json:"remote_state,omitempty"`
	Readiness         string   `json:"readiness,omitempty"`
	OwnershipMode     string   `json:"ownership_mode,omitempty"`
	EntryMode         string   `json:"entry_mode,omitempty"`
	SourceMode        string   `json:"source_mode,omitempty"`
	DiscussionNumber  int      `json:"discussion_number,omitempty"`
	DiscussionURL     string   `json:"discussion_url,omitempty"`
	ParentIssueNumber int      `json:"parent_issue_number,omitempty"`
	MilestoneNumber   int      `json:"milestone_number,omitempty"`
	MilestoneTitle    string   `json:"milestone_title,omitempty"`
	BlockedBy         []string `json:"blocked_by,omitempty"`
	UpdatedAt         string   `json:"updated_at,omitempty"`
}

type GitHubProjectDecisionRecord struct {
	Slug             string `json:"slug"`
	Decision         string `json:"decision"`
	Reason           string `json:"reason,omitempty"`
	SpecCount        int    `json:"spec_count,omitempty"`
	MilestoneNumber  int    `json:"milestone_number,omitempty"`
	MilestoneTitle   string `json:"milestone_title,omitempty"`
	SourceMode       string `json:"source_mode,omitempty"`
	EntryMode        string `json:"entry_mode,omitempty"`
	DiscussionNumber int    `json:"discussion_number,omitempty"`
	DiscussionURL    string `json:"discussion_url,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

type GitHubStoryRecord struct {
	Slug                  string   `json:"slug"`
	Title                 string   `json:"title"`
	Epic                  string   `json:"epic"`
	Spec                  string   `json:"spec"`
	Status                string   `json:"status,omitempty"`
	Description           string   `json:"description,omitempty"`
	AcceptanceCriteria    []string `json:"acceptance_criteria,omitempty"`
	Verification          []string `json:"verification,omitempty"`
	Resources             []string `json:"resources,omitempty"`
	Dependencies          []string `json:"dependencies,omitempty"`
	AsyncNotes            []string `json:"async_notes,omitempty"`
	ScopeFit              string   `json:"scope_fit,omitempty"`
	VerticalSliceCheck    string   `json:"vertical_slice_check,omitempty"`
	HiddenPrerequisites   string   `json:"hidden_prerequisites,omitempty"`
	VerificationGaps      string   `json:"verification_gaps,omitempty"`
	RewriteRecommendation string   `json:"rewrite_recommendation,omitempty"`
	IssueNumber           int      `json:"issue_number,omitempty"`
	IssueURL              string   `json:"issue_url,omitempty"`
	RemoteState           string   `json:"remote_state,omitempty"`
	PlanningPRNumber      int      `json:"planning_pr_number,omitempty"`
	PlanningPRURL         string   `json:"planning_pr_url,omitempty"`
	PlanningPRMerged      bool     `json:"planning_pr_merged,omitempty"`
	DocRefMode            string   `json:"doc_ref_mode,omitempty"`
	DocRef                string   `json:"doc_ref,omitempty"`
	Ready                 bool     `json:"ready,omitempty"`
	BlockedReasons        []string `json:"blocked_reasons,omitempty"`
	VisibleReadyMarkerSet bool     `json:"visible_ready_marker_set,omitempty"`
	UpdatedAt             string   `json:"updated_at,omitempty"`
}

type ArchivedPath struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ArchivedSpec struct {
	Slug             string `json:"slug"`
	Path             string `json:"path"`
	Title            string `json:"title,omitempty"`
	Status           string `json:"status,omitempty"`
	SourceLegacyEpic string `json:"source_legacy_epic,omitempty"`
}

type MigrationState struct {
	SchemaVersion int            `json:"schema_version"`
	Known         []string       `json:"known"`
	LastRunAt     string         `json:"last_run_at,omitempty"`
	Status        string         `json:"status,omitempty"`
	LastOperation string         `json:"last_operation,omitempty"`
	LastCreated   []string       `json:"last_created,omitempty"`
	LastUpdated   []string       `json:"last_updated,omitempty"`
	LastArchived  []ArchivedPath `json:"last_archived,omitempty"`
	History       []MigrationRun `json:"history,omitempty"`
}

type MigrationRun struct {
	Operation string         `json:"operation"`
	At        string         `json:"at"`
	Status    string         `json:"status"`
	Created   []string       `json:"created,omitempty"`
	Updated   []string       `json:"updated,omitempty"`
	Archived  []ArchivedPath `json:"archived,omitempty"`
}

type ArchiveManifest struct {
	ArchivedAt    string         `json:"archived_at"`
	PlanningModel string         `json:"planning_model"`
	Sources       []ArchivedPath `json:"sources"`
	ActiveSpecs   []ArchivedSpec `json:"active_specs,omitempty"`
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
	IdeasDir       string
	ArchiveDir     string
	EpicsDir       string
	SpecsDir       string
	StoriesDir     string
	MetaDir        string
	WorkspaceFile  string
	MigrationsFile string
	GitHubFile     string
	SessionsFile   string
}

type InitResult struct {
	Info    *Info
	Created []string
}

type UpdateResult struct {
	Info     *Info
	Created  []string
	Updated  []string
	Archived []ArchivedPath
}

type UpdateOptions struct {
	ArchiveLegacy bool
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
		IdeasDir:       filepath.Join(planDir, "ideas"),
		ArchiveDir:     filepath.Join(planDir, "archive"),
		EpicsDir:       filepath.Join(planDir, "epics"),
		SpecsDir:       filepath.Join(planDir, "specs"),
		StoriesDir:     filepath.Join(planDir, "stories"),
		MetaDir:        metaDir,
		WorkspaceFile:  filepath.Join(metaDir, "workspace.json"),
		MigrationsFile: filepath.Join(metaDir, "migrations.json"),
		GitHubFile:     filepath.Join(metaDir, "github.json"),
		SessionsFile:   filepath.Join(metaDir, "guided_sessions.json"),
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
		i.newSurface("ideas/", i.IdeasDir, "user_authored", "dir"),
		i.newSurface("specs/", i.SpecsDir, "user_authored", "dir"),
		i.newSurface("archive/", i.ArchiveDir, "user_authored", "dir"),
	}
}

func (i *Info) ToolManagedSurfaces() []Surface {
	return []Surface{
		i.newSurface(".meta/", i.MetaDir, "tool_managed", "dir"),
		i.newSurface("workspace.json", i.WorkspaceFile, "tool_managed", "file"),
		i.newSurface("migrations.json", i.MigrationsFile, "tool_managed", "file"),
		i.newSurface("github.json", i.GitHubFile, "tool_managed", "file"),
		i.newSurface("guided_sessions.json", i.SessionsFile, "tool_managed", "file"),
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
	if created, err := ensureGitHubState(info.GitHubFile, now); err != nil {
		return nil, err
	} else if created {
		result.Created = append(result.Created, rel(info.ProjectDir, info.GitHubFile))
	}
	if created, err := ensureGuidedSessionState(info.SessionsFile, now); err != nil {
		return nil, err
	} else if created {
		result.Created = append(result.Created, rel(info.ProjectDir, info.SessionsFile))
	}
	if err := recordMigrationRun(info, "init", result.Created, nil, nil, now); err != nil {
		return nil, err
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

func (i *Info) legacyDirectoryPaths() []string {
	return []string{i.EpicsDir, i.StoriesDir}
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
	if meta.SourceMode == "" {
		meta.SourceMode = SourceOfTruthLocal
	}
	if meta.StoryBackend == "" {
		meta.StoryBackend = StoryBackendLocal
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

func (m *Manager) ReadGitHubState() (*GitHubState, error) {
	info, err := m.Resolve()
	if err != nil {
		return nil, err
	}
	return readGitHubStateFile(info.GitHubFile)
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

func readGitHubStateFile(path string) (*GitHubState, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read github state: %w", err)
	}
	var state GitHubState
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("parse github state: %w", err)
	}
	if state.Stories == nil {
		state.Stories = map[string]GitHubStoryRecord{}
	}
	if state.Planning == nil {
		state.Planning = map[string]GitHubPlanningRecord{}
	}
	if state.ProjectDecisions == nil {
		state.ProjectDecisions = map[string]GitHubProjectDecisionRecord{}
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

	githubState, err := readGitHubStateFile(info.GitHubFile)
	if err != nil {
		if !contains(report.Missing, "github.json") {
			report.Broken = append(report.Broken, "github.json")
		}
	} else if err := validateGitHubState(githubState); err != nil {
		if !contains(report.Broken, "github.json") {
			report.Broken = append(report.Broken, "github.json")
		}
	}

	sessionsState, err := readGuidedSessionStateFile(info.SessionsFile)
	if err != nil {
		if !contains(report.Missing, "guided_sessions.json") {
			report.Broken = append(report.Broken, "guided_sessions.json")
		}
	} else if err := validateGuidedSessionState(sessionsState); err != nil {
		if !contains(report.Broken, "guided_sessions.json") {
			report.Broken = append(report.Broken, "guided_sessions.json")
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
	return m.UpdateWithOptions(UpdateOptions{})
}

func (m *Manager) UpdateWithOptions(opts UpdateOptions) (*UpdateResult, error) {
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
	return m.repairWorkspace(info, "update", opts)
}

func (m *Manager) Adopt() (*UpdateResult, error) {
	info, err := m.Resolve()
	if err != nil {
		return nil, err
	}
	return m.repairWorkspace(info, "adopt", UpdateOptions{})
}

func (m *Manager) repairWorkspace(info *Info, operation string, opts UpdateOptions) (*UpdateResult, error) {
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
	if created, err := ensureGitHubState(info.GitHubFile, now); err != nil {
		return nil, err
	} else if created {
		result.Created = append(result.Created, rel(info.ProjectDir, info.GitHubFile))
	} else {
		updated, err := reconcileGitHubState(info.GitHubFile, now)
		if err != nil {
			return nil, err
		}
		if updated {
			result.Updated = append(result.Updated, rel(info.ProjectDir, info.GitHubFile))
		}
	}
	if created, err := ensureGuidedSessionState(info.SessionsFile, now); err != nil {
		return nil, err
	} else if created {
		result.Created = append(result.Created, rel(info.ProjectDir, info.SessionsFile))
	} else {
		updated, err := reconcileGuidedSessionState(info.SessionsFile, now)
		if err != nil {
			return nil, err
		}
		if updated {
			result.Updated = append(result.Updated, rel(info.ProjectDir, info.SessionsFile))
		}
	}
	if opts.ArchiveLegacy {
		archiveCreated, archiveUpdated, archived, err := archiveLegacyHierarchy(info, now)
		if err != nil {
			return nil, err
		}
		result.Created = append(result.Created, archiveCreated...)
		result.Updated = append(result.Updated, archiveUpdated...)
		result.Archived = append(result.Archived, archived...)
	}
	if err := recordMigrationRun(info, operation, result.Created, result.Updated, result.Archived, now); err != nil {
		return nil, err
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

func (m *Manager) WriteWorkspaceMeta(meta WorkspaceMeta) error {
	info, err := m.Resolve()
	if err != nil {
		return err
	}
	normalized, _, err := normalizeWorkspaceMeta(meta, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return err
	}
	return writeJSON(info.WorkspaceFile, normalized)
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

func ensureGitHubState(path, now string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	state := defaultGitHubState(now)
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

func reconcileGitHubState(path, now string) (bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	var state GitHubState
	if err := json.Unmarshal(raw, &state); err != nil {
		return true, writeJSON(path, defaultGitHubState(now))
	}
	normalized, changed, err := normalizeGitHubState(state, now)
	if err != nil {
		return false, err
	}
	if !changed {
		return false, nil
	}
	return true, writeJSON(path, normalized)
}

func archiveLegacyHierarchy(info *Info, now string) ([]string, []string, []ArchivedPath, error) {
	var created []string
	var updated []string
	var archived []ArchivedPath
	var activeSpecs []ArchivedSpec

	batchName := archiveBatchName(now)
	batchDir := filepath.Join(info.ArchiveDir, batchName)
	batchCreated := false

	for _, legacyDir := range info.legacyDirectoryPaths() {
		entries, err := os.ReadDir(legacyDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, nil, nil, err
		}
		if len(entries) == 0 {
			if err := os.Remove(legacyDir); err != nil && !os.IsNotExist(err) {
				return nil, nil, nil, err
			}
			updated = append(updated, rel(info.ProjectDir, legacyDir))
			continue
		}
		if !batchCreated {
			made, err := ensureDir(batchDir)
			if err != nil {
				return nil, nil, nil, err
			}
			if made {
				created = append(created, rel(info.ProjectDir, batchDir))
			}
			batchCreated = true
		}
		destDir := filepath.Join(batchDir, filepath.Base(legacyDir))
		if filepath.Base(legacyDir) == "epics" {
			specs, err := archivedSpecsForLegacyEpics(info, legacyDir, destDir)
			if err != nil {
				return nil, nil, nil, err
			}
			activeSpecs = append(activeSpecs, specs...)
		}
		if err := os.Rename(legacyDir, destDir); err != nil {
			return nil, nil, nil, err
		}
		archived = append(archived, ArchivedPath{
			From: rel(info.ProjectDir, legacyDir),
			To:   rel(info.ProjectDir, destDir),
		})
	}

	if !batchCreated {
		return created, updated, archived, nil
	}

	manifest := ArchiveManifest{
		ArchivedAt:    now,
		PlanningModel: PlanningModel,
		Sources:       append([]ArchivedPath(nil), archived...),
		ActiveSpecs:   activeSpecs,
	}
	manifestPath := filepath.Join(batchDir, "migration.json")
	if err := writeJSON(manifestPath, manifest); err != nil {
		return nil, nil, nil, err
	}
	created = append(created, rel(info.ProjectDir, manifestPath))

	return created, updated, archived, nil
}

func archivedSpecsForLegacyEpics(info *Info, legacyDir, destDir string) ([]ArchivedSpec, error) {
	entries, err := os.ReadDir(legacyDir)
	if err != nil {
		return nil, err
	}
	specs := make([]ArchivedSpec, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		epicPath := filepath.Join(legacyDir, entry.Name())
		epic, err := notes.Read(epicPath)
		if err != nil {
			return nil, err
		}
		specSlug := safeLegacySpecSlug(stringValue(epic.Metadata["spec"]), slugFromPath(epicPath))
		specPath := filepath.Join(info.SpecsDir, specSlug+".md")
		spec := ArchivedSpec{
			Slug:             specSlug,
			Path:             rel(info.ProjectDir, specPath),
			Status:           "draft",
			SourceLegacyEpic: rel(info.ProjectDir, filepath.Join(destDir, entry.Name())),
		}
		if current, err := notes.Read(specPath); err == nil {
			spec.Title = current.Title
			if status := stringValue(current.Metadata["status"]); status != "" {
				spec.Status = status
			}
		}
		specs = append(specs, spec)
	}
	return specs, nil
}

func safeLegacySpecSlug(raw, fallback string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return slugFromPath(fallback)
	}
	if strings.ContainsAny(raw, `/\`) {
		return slugFromPath(fallback)
	}
	base := filepath.Base(raw)
	if base == "." || base == ".." {
		return slugFromPath(fallback)
	}
	return slugFromPath(base)
}

func archiveBatchName(now string) string {
	parsed, err := time.Parse(time.RFC3339, now)
	if err != nil {
		return "legacy-" + strings.NewReplacer(":", "", "-", "", ".", "").Replace(now)
	}
	return "legacy-" + parsed.UTC().Format("20060102T150405Z")
}

func defaultWorkspaceMeta(now string) WorkspaceMeta {
	return WorkspaceMeta{
		SchemaVersion: CurrentSchemaVersion,
		PlanningModel: PlanningModel,
		SourceMode:    SourceOfTruthLocal,
		StoryBackend:  StoryBackendLocal,
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

func defaultGitHubState(now string) GitHubState {
	return GitHubState{
		LastUpdatedAt:    now,
		Stories:          map[string]GitHubStoryRecord{},
		Planning:         map[string]GitHubPlanningRecord{},
		ProjectDecisions: map[string]GitHubProjectDecisionRecord{},
	}
}

func normalizeWorkspaceMeta(meta WorkspaceMeta, now string) (WorkspaceMeta, bool, error) {
	if meta.SchemaVersion > CurrentSchemaVersion {
		return WorkspaceMeta{}, false, fmt.Errorf("workspace schema version %d is newer than supported version %d", meta.SchemaVersion, CurrentSchemaVersion)
	}
	if meta.PlanningModel != "" && meta.PlanningModel != PlanningModel && meta.PlanningModel != LegacyPlanningModel {
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
	} else if normalized.PlanningModel == LegacyPlanningModel {
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

func normalizeGitHubState(state GitHubState, now string) (GitHubState, bool, error) {
	normalized := state
	changed := false
	if normalized.LastUpdatedAt == "" {
		normalized.LastUpdatedAt = now
		changed = true
	}
	if changed {
		normalized.LastUpdatedAt = now
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
	case meta.SourceMode != "" && meta.SourceMode != SourceOfTruthLocal && meta.SourceMode != SourceOfTruthGitHub && meta.SourceMode != SourceOfTruthHybrid:
		return fmt.Errorf("source_mode %q is not supported", meta.SourceMode)
	case meta.StoryBackend != "" && meta.StoryBackend != StoryBackendLocal && meta.StoryBackend != StoryBackendGitHub:
		return fmt.Errorf("story_backend %q is not supported", meta.StoryBackend)
	case meta.CreatedAt == "":
		return fmt.Errorf("created_at is required")
	case meta.UpdatedAt == "":
		return fmt.Errorf("updated_at is required")
	default:
		return nil
	}
}

func validateGitHubState(state *GitHubState) error {
	if state == nil {
		return fmt.Errorf("github state is required")
	}
	return nil
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

func recordMigrationRun(info *Info, operation string, created, updated []string, archived []ArchivedPath, now string) error {
	if len(created) == 0 && len(updated) == 0 && len(archived) == 0 {
		return nil
	}
	state, err := readMigrationStateFile(info.MigrationsFile)
	if err != nil {
		defaultState := defaultMigrationState(now)
		state = &defaultState
	}
	created = append([]string(nil), created...)
	updated = append([]string(nil), updated...)
	archived = append([]ArchivedPath(nil), archived...)
	run := MigrationRun{
		Operation: operation,
		At:        now,
		Status:    "current",
		Created:   created,
		Updated:   updated,
		Archived:  archived,
	}
	state.LastRunAt = now
	state.Status = "current"
	state.LastOperation = operation
	state.LastCreated = created
	state.LastUpdated = updated
	state.LastArchived = archived
	state.History = append(state.History, run)
	if len(state.History) > 10 {
		state.History = append([]MigrationRun(nil), state.History[len(state.History)-10:]...)
	}
	return writeJSON(info.MigrationsFile, state)
}

func (m *Manager) WriteGitHubState(state GitHubState) error {
	info, err := m.Resolve()
	if err != nil {
		return err
	}
	normalized, _, err := normalizeGitHubState(state, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return err
	}
	return writeJSON(info.GitHubFile, normalized)
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

func slugFromPath(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
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
	return contains(missing, ".meta/") || contains(missing, "workspace.json") || contains(missing, "migrations.json") || contains(missing, "github.json") || contains(missing, "guided_sessions.json")
}

func isBrokenWorkspace(broken []string) bool {
	return contains(broken, "workspace.json") || contains(broken, "migrations.json") || contains(broken, "github.json") || contains(broken, "guided_sessions.json")
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
