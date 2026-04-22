package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/notes"
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
		filepath.Join(root, ".plan", "ideas"),
		filepath.Join(root, ".plan", "archive"),
		filepath.Join(root, ".plan", "specs"),
		filepath.Join(root, ".plan", ".meta"),
		filepath.Join(root, ".plan", "PROJECT.md"),
		filepath.Join(root, ".plan", "ROADMAP.md"),
		filepath.Join(root, ".plan", ".meta", "workspace.json"),
		filepath.Join(root, ".plan", ".meta", "migrations.json"),
		filepath.Join(root, ".plan", ".meta", "github.json"),
		filepath.Join(root, ".plan", ".meta", "guided_sessions.json"),
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
	if len(contract.ToolManaged) != 5 {
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

func TestDoctorReportsAdoptableBeforeInit(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# repo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manager := New(root)

	report, err := manager.Doctor()
	if err != nil {
		t.Fatal(err)
	}
	if report.Initialized {
		t.Fatalf("expected uninitialized workspace report: %+v", report)
	}
	if report.WorkspaceStatus != "adoptable" || report.MigrationStatus != "not_initialized" {
		t.Fatalf("unexpected adoptable report state: %+v", report)
	}
	if !contains(report.Guidance, "Run `plan adopt --project .` to create and register the local .plan workspace for this repo.") {
		t.Fatalf("expected adopt guidance: %+v", report)
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
	if report.SchemaVersion != CurrentSchemaVersion {
		t.Fatalf("unexpected schema version: %+v", report)
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
	if !contains(report.Guidance, "Run `plan update --project .` to recreate missing plan-managed surfaces without overwriting user-authored notes.") {
		t.Fatalf("expected missing-workspace guidance: %+v", report)
	}
}

func TestDoctorReportsPartialWhenToolManagedSurfacesAreMissing(t *testing.T) {
	root := t.TempDir()
	manager := New(root)
	if _, err := manager.Init(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, ".plan", ".meta", "workspace.json")); err != nil {
		t.Fatal(err)
	}

	report, err := manager.Doctor()
	if err != nil {
		t.Fatal(err)
	}
	if report.WorkspaceStatus != "partial" {
		t.Fatalf("unexpected workspace status: %+v", report)
	}
	if report.MigrationStatus != "partial" {
		t.Fatalf("unexpected migration status for partial workspace: %+v", report)
	}
	if !contains(report.Missing, "workspace.json") {
		t.Fatalf("expected missing workspace metadata in report: %+v", report)
	}
	if !contains(report.Guidance, "Run `plan adopt --project .` to finish adopting the partial .plan workspace.") {
		t.Fatalf("expected partial-workspace guidance: %+v", report)
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
	if !contains(report.Guidance, "Run `plan update --project .` to repair broken plan-managed metadata and restore a current workspace.") {
		t.Fatalf("expected broken-workspace guidance: %+v", report)
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
	if err := os.Remove(filepath.Join(root, ".plan", ".meta", "github.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, ".plan", ".meta", "guided_sessions.json")); err != nil {
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
	if !contains(result.Created, ".plan/.meta/github.json") {
		t.Fatalf("expected GitHub state recreation: %+v", result)
	}
	if !contains(result.Created, ".plan/.meta/guided_sessions.json") {
		t.Fatalf("expected guided session state recreation: %+v", result)
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

func TestAdoptCreatesWorkspaceWithoutTouchingRepoFiles(t *testing.T) {
	root := t.TempDir()
	manager := New(root)

	readmePath := filepath.Join(root, "README.md")
	const readme = "# existing repo\n"
	if err := os.WriteFile(readmePath, []byte(readme), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := manager.Adopt()
	if err != nil {
		t.Fatal(err)
	}
	if !contains(result.Created, ".plan") || !contains(result.Created, ".plan/PROJECT.md") || !contains(result.Created, ".plan/.meta/workspace.json") {
		t.Fatalf("expected adopt to create plan workspace surfaces: %+v", result)
	}

	raw, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != readme {
		t.Fatalf("expected non-plan repo files to stay untouched:\n%s", raw)
	}

	report, err := manager.Doctor()
	if err != nil {
		t.Fatal(err)
	}
	if report.WorkspaceStatus != "current" {
		t.Fatalf("expected adopted workspace to be current: %+v", report)
	}
}

func TestAdoptRepairsPartialWorkspace(t *testing.T) {
	root := t.TempDir()
	manager := New(root)
	if err := os.MkdirAll(filepath.Join(root, ".plan"), 0o755); err != nil {
		t.Fatal(err)
	}
	projectPath := filepath.Join(root, ".plan", "PROJECT.md")
	const customProject = "# preserved project\n"
	if err := os.WriteFile(projectPath, []byte(customProject), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := manager.Adopt()
	if err != nil {
		t.Fatal(err)
	}
	if !contains(result.Created, ".plan/.meta/workspace.json") {
		t.Fatalf("expected adopt to fill missing managed surfaces: %+v", result)
	}
	raw, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != customProject {
		t.Fatalf("expected adopt to preserve existing plan notes:\n%s", raw)
	}

	report, err := manager.Doctor()
	if err != nil {
		t.Fatal(err)
	}
	if report.WorkspaceStatus != "current" || report.MigrationStatus != "current" {
		t.Fatalf("expected adopted partial workspace to become current: %+v", report)
	}
}

func TestMigrationStateRecordsMeaningfulRepairDetails(t *testing.T) {
	root := t.TempDir()
	manager := New(root)

	if _, err := manager.Adopt(); err != nil {
		t.Fatal(err)
	}
	state, err := manager.ReadMigrationState()
	if err != nil {
		t.Fatal(err)
	}
	if state.LastOperation != "adopt" || len(state.History) != 1 {
		t.Fatalf("expected adopt run to be recorded: %+v", state)
	}
	if !contains(state.LastCreated, ".plan/PROJECT.md") {
		t.Fatalf("expected adopt details to include created workspace files: %+v", state)
	}

	if err := os.Remove(filepath.Join(root, ".plan", "ROADMAP.md")); err != nil {
		t.Fatal(err)
	}
	result, err := manager.Update()
	if err != nil {
		t.Fatal(err)
	}
	if !contains(result.Created, ".plan/ROADMAP.md") {
		t.Fatalf("expected update to repair roadmap: %+v", result)
	}

	state, err = manager.ReadMigrationState()
	if err != nil {
		t.Fatal(err)
	}
	if state.LastOperation != "update" || len(state.History) != 2 {
		t.Fatalf("expected update repair to append migration history: %+v", state)
	}
	last := state.History[len(state.History)-1]
	if last.Operation != "update" || !contains(last.Created, ".plan/ROADMAP.md") {
		t.Fatalf("expected update run details in migration history: %+v", last)
	}
}

func TestRepeatedAdoptAndUpdateRemainIdempotent(t *testing.T) {
	root := t.TempDir()
	manager := New(root)

	if _, err := manager.Adopt(); err != nil {
		t.Fatal(err)
	}
	state, err := manager.ReadMigrationState()
	if err != nil {
		t.Fatal(err)
	}
	initialHistory := len(state.History)

	adoptAgain, err := manager.Adopt()
	if err != nil {
		t.Fatal(err)
	}
	if len(adoptAgain.Created) != 0 || len(adoptAgain.Updated) != 0 {
		t.Fatalf("expected repeated adopt to stay idempotent: %+v", adoptAgain)
	}
	updateAgain, err := manager.Update()
	if err != nil {
		t.Fatal(err)
	}
	if len(updateAgain.Created) != 0 || len(updateAgain.Updated) != 0 {
		t.Fatalf("expected repeated update to stay idempotent: %+v", updateAgain)
	}

	state, err = manager.ReadMigrationState()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.History) != initialHistory {
		t.Fatalf("expected no-op adopt/update to avoid mutation churn: %+v", state)
	}
}

func TestUpdateRecreatesMissingScaffoldingWithoutOverwritingExistingNotes(t *testing.T) {
	root := t.TempDir()
	manager := New(root)
	if _, err := manager.Init(); err != nil {
		t.Fatal(err)
	}

	projectPath := filepath.Join(root, ".plan", "PROJECT.md")
	const customProject = "# kept project note\n"
	if err := os.WriteFile(projectPath, []byte(customProject), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, ".plan", "ROADMAP.md")); err != nil {
		t.Fatal(err)
	}

	result, err := manager.Update()
	if err != nil {
		t.Fatal(err)
	}
	if !contains(result.Created, ".plan/ROADMAP.md") {
		t.Fatalf("expected roadmap recreation: %+v", result)
	}

	raw, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != customProject {
		t.Fatalf("project note was overwritten:\n%s", raw)
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

func TestUpdateMigratesLegacyWorkspaceMetadataToSpecFirstModel(t *testing.T) {
	root := t.TempDir()
	manager := New(root)
	if _, err := manager.Init(); err != nil {
		t.Fatal(err)
	}
	legacy := `{
  "schema_version": 1,
  "planning_model": "epic_spec_story_v1",
  "story_backend": "local",
  "created_at": "2026-04-01T00:00:00Z",
  "updated_at": "2026-04-01T00:00:00Z"
}
`
	if err := os.WriteFile(filepath.Join(root, ".plan", ".meta", "workspace.json"), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := manager.Update()
	if err != nil {
		t.Fatal(err)
	}
	if !contains(result.Updated, ".plan/.meta/workspace.json") {
		t.Fatalf("expected workspace metadata migration: %+v", result)
	}

	meta, err := manager.ReadWorkspaceMeta()
	if err != nil {
		t.Fatal(err)
	}
	if meta.PlanningModel != PlanningModel || meta.SchemaVersion != 1 {
		t.Fatalf("expected legacy workspace metadata to normalize to spec-first model: %+v", meta)
	}
}

func TestUpdateWithArchiveLegacyMovesLegacyHierarchyAndPreservesActiveSpecs(t *testing.T) {
	root := t.TempDir()
	manager := New(root)
	if _, err := manager.Init(); err != nil {
		t.Fatal(err)
	}

	epicsDir := filepath.Join(root, ".plan", "epics")
	storiesDir := filepath.Join(root, ".plan", "stories")
	if err := os.MkdirAll(epicsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(storiesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	specPath := filepath.Join(root, ".plan", "specs", "billing.md")
	if _, err := notes.Create(specPath, "Billing Spec", "spec", "# Billing Spec\n", map[string]any{
		"slug":   "billing",
		"status": "approved",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := notes.Create(filepath.Join(epicsDir, "billing.md"), "Billing", "epic", "# Billing\n", map[string]any{
		"slug": "billing",
		"spec": "billing",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := notes.Create(filepath.Join(storiesDir, "billing-ui.md"), "Billing UI", "story", "# Billing UI\n", map[string]any{
		"slug":   "billing-ui",
		"status": "todo",
		"epic":   "billing",
		"spec":   "billing",
	}); err != nil {
		t.Fatal(err)
	}

	result, err := manager.UpdateWithOptions(UpdateOptions{ArchiveLegacy: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Archived) != 2 {
		t.Fatalf("expected both legacy directories to be archived: %+v", result)
	}
	if _, err := os.Stat(epicsDir); !os.IsNotExist(err) {
		t.Fatalf("expected legacy epics dir to be removed, got %v", err)
	}
	if _, err := os.Stat(storiesDir); !os.IsNotExist(err) {
		t.Fatalf("expected legacy stories dir to be removed, got %v", err)
	}
	if _, err := os.Stat(specPath); err != nil {
		t.Fatalf("expected active spec to remain in place: %v", err)
	}

	var batchDir string
	for _, move := range result.Archived {
		if strings.HasSuffix(move.To, "/epics") {
			batchDir = filepath.Dir(filepath.Join(root, filepath.FromSlash(move.To)))
			break
		}
	}
	if batchDir == "" {
		t.Fatalf("expected archived epics destination in result: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(batchDir, "epics", "billing.md")); err != nil {
		t.Fatalf("expected archived epic to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(batchDir, "stories", "billing-ui.md")); err != nil {
		t.Fatalf("expected archived story to exist: %v", err)
	}

	manifestPath := filepath.Join(batchDir, "migration.json")
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest ArchiveManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.ActiveSpecs) != 1 || manifest.ActiveSpecs[0].Path != ".plan/specs/billing.md" {
		t.Fatalf("expected manifest to describe preserved active spec: %+v", manifest)
	}

	state, err := manager.ReadMigrationState()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.LastArchived) != 2 {
		t.Fatalf("expected migration state to record archived moves: %+v", state)
	}
	last := state.History[len(state.History)-1]
	if len(last.Archived) != 2 {
		t.Fatalf("expected migration history to record archived moves: %+v", last)
	}
}

func TestEnsureInitializedRejectsInvalidWorkspaceMetadata(t *testing.T) {
	root := t.TempDir()
	manager := New(root)
	if _, err := manager.Init(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".plan", ".meta", "workspace.json"), []byte(`{"schema_version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := manager.EnsureInitialized(); err == nil {
		t.Fatal("expected invalid workspace metadata to fail initialization check")
	}
}
