package planning

import (
	"strings"
	"testing"

	"plan/internal/workspace"
)

func TestSetGitHubProjectIssueStatusUpdatesTrackedIssueItem(t *testing.T) {
	client := projectAutomationClient()
	client.projectItems = []GitHubProjectItemResult{{IssueNumber: 68, ItemID: "PVTI_68"}}
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	t.Cleanup(reset)
	manager := setupProjectAutomationManager(t)

	result, err := manager.SetGitHubProjectIssueStatus(GitHubProjectStatusInput{IssueNumber: 68, Status: "in-review"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != projectValueInReview || result.ItemID != "PVTI_68" {
		t.Fatalf("unexpected status result: %+v", result)
	}
	if !stubHasProjectValue(client.projectValues, 68, projectFieldStatus, projectValueInReview) {
		t.Fatalf("expected status field update: %+v", client.projectValues)
	}
}

func TestSetGitHubProjectIssueStatusRequiresProjectDecisionMetadata(t *testing.T) {
	client := projectAutomationClient()
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	t.Cleanup(reset)
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)
	state, err := ws.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	state.Repo = "JimmyMcBride/plan"
	state.Planning["status"] = workspace.GitHubPlanningRecord{Slug: "status", Kind: "spec", Title: "Status", IssueNumber: 68}
	if err := ws.WriteGitHubState(*state); err != nil {
		t.Fatal(err)
	}

	_, err = manager.SetGitHubProjectIssueStatus(GitHubProjectStatusInput{IssueNumber: 68, Status: "done"})
	if err == nil || !strings.Contains(err.Error(), "no project decision metadata") {
		t.Fatalf("expected missing metadata error, got %v", err)
	}
}

func TestCheckFlagsProjectMissingItemAndStaleField(t *testing.T) {
	client := projectAutomationClient()
	client.projectItems = []GitHubProjectItemResult{{IssueNumber: 69, ItemID: "PVTI_69"}}
	client.projectValues = []stubProjectValue{
		{ItemID: "PVTI_69", Field: projectFieldType, Value: projectValueTracking},
		{ItemID: "PVTI_69", Field: projectFieldStage, Value: projectValueApproved},
		{ItemID: "PVTI_69", Field: projectFieldReady, Value: projectValueYes},
		{ItemID: "PVTI_69", Field: projectFieldArea, Value: "workspace"},
	}
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	t.Cleanup(reset)
	manager := setupProjectAutomationManager(t)
	state, err := manager.workspace.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	state.Planning["stale"] = workspace.GitHubPlanningRecord{
		Slug:            "stale",
		Kind:            "spec",
		Title:           "Stale",
		IssueNumber:     69,
		Readiness:       string(ReadinessReady),
		MilestoneNumber: 7,
		MilestoneTitle:  "Workspace",
	}
	if err := manager.workspace.WriteGitHubState(*state); err != nil {
		t.Fatal(err)
	}

	report, err := manager.Check(CheckInput{})
	if err != nil {
		t.Fatal(err)
	}
	assertHasFinding(t, report.Findings, "github_project.missing_item", "GitHub Project")
	assertHasFinding(t, report.Findings, "github_project.stale_item_field", "GitHub Project")
	if client.getIssueCalls != 0 {
		t.Fatalf("expected project item lookup to carry issue metadata, got %d GetIssue calls", client.getIssueCalls)
	}
}

func TestReconcileRepairsSafeProjectDrift(t *testing.T) {
	client := projectAutomationClient()
	client.projectAddNilValues = true
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	t.Cleanup(reset)
	manager := setupProjectAutomationManager(t)

	result, err := manager.ReconcileGitHubStories(GitHubReconcileOptions{UpdateVisible: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.UpdatedProjectItems) != 1 || result.UpdatedProjectItems[0] != "#68" {
		t.Fatalf("expected one repaired project item: %+v", result)
	}
	if len(client.projectItems) != 1 || client.projectItems[0].IssueNumber != 68 {
		t.Fatalf("expected missing project item to be added: %+v", client.projectItems)
	}
	if !stubHasProjectValue(client.projectValues, 68, projectFieldType, projectValueSpec) ||
		!stubHasProjectValue(client.projectValues, 68, projectFieldStatus, projectValueTodo) {
		t.Fatalf("expected project item values to be repaired: %+v", client.projectValues)
	}
	if client.getIssueCalls != 0 {
		t.Fatalf("expected reconcile to reuse project item issue metadata, got %d GetIssue calls", client.getIssueCalls)
	}
}

func TestReconcileContinuesWhenExistingStatusFieldHasUnusedMissingOptions(t *testing.T) {
	client := projectAutomationClient()
	trimProjectFieldOptions(client.projects[12], projectFieldStatus, []string{projectValueTodo})
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	t.Cleanup(reset)
	manager := setupProjectAutomationManager(t)

	result, err := manager.ReconcileGitHubStories(GitHubReconcileOptions{UpdateVisible: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.UpdatedProjectItems) != 1 || result.UpdatedProjectItems[0] != "#68" {
		t.Fatalf("expected one repaired project item: %+v", result)
	}
	if !stubHasProjectValue(client.projectValues, 68, projectFieldStatus, projectValueTodo) {
		t.Fatalf("expected reconcile to set safe available status option: %+v", client.projectValues)
	}
}

func setupProjectAutomationManager(t *testing.T) *Manager {
	t.Helper()
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)
	if _, err := manager.SetSourceMode(workspace.SourceOfTruthGitHub); err != nil {
		t.Fatal(err)
	}
	state, err := ws.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	state.Repo = "JimmyMcBride/plan"
	state.RepoURL = "https://github.com/JimmyMcBride/plan"
	state.DefaultBranch = "develop"
	state.Planning["status"] = workspace.GitHubPlanningRecord{
		Slug:            "status",
		Kind:            "spec",
		Title:           "Status",
		IssueNumber:     68,
		Readiness:       string(ReadinessReady),
		MilestoneNumber: 7,
		MilestoneTitle:  "Workspace",
	}
	state.ProjectDecisions["workspace"] = workspace.GitHubProjectDecisionRecord{
		Slug:            "workspace",
		Decision:        "connect",
		InitiativeSlug:  "workspace",
		SpecCount:       1,
		MilestoneNumber: 7,
		MilestoneTitle:  "Workspace",
		ProjectOwner:    "JimmyMcBride",
		ProjectNumber:   12,
		ProjectID:       "PVT_workspace",
		ProjectURL:      "https://github.com/users/JimmyMcBride/projects/12",
		FieldIDs:        projectFieldIDsForTest(),
	}
	if err := ws.WriteGitHubState(*state); err != nil {
		t.Fatal(err)
	}
	return manager
}

func projectAutomationClient() *stubGitHubClient {
	return &stubGitHubClient{
		preflight: &GitHubRepoInfo{Repo: "JimmyMcBride/plan", RepoURL: "https://github.com/JimmyMcBride/plan", DefaultBranch: "develop"},
		context: &GitHubContext{
			Repo:          GitHubRepoInfo{Repo: "JimmyMcBride/plan", RepoURL: "https://github.com/JimmyMcBride/plan", DefaultBranch: "develop"},
			CurrentBranch: "codex/project-status",
			CurrentSHA:    "abc123",
		},
		issues: map[int]*GitHubIssue{
			68: {Number: 68, URL: "https://github.com/JimmyMcBride/plan/issues/68", Title: "Status", State: "open", Labels: []string{planIssueSpecLabel}},
			69: {Number: 69, URL: "https://github.com/JimmyMcBride/plan/issues/69", Title: "Stale", State: "open", Labels: []string{planIssueSpecLabel}},
		},
		projects: map[int]*GitHubProjectWorkspace{
			12: {
				Owner:  "JimmyMcBride",
				Number: 12,
				ID:     "PVT_workspace",
				URL:    "https://github.com/users/JimmyMcBride/projects/12",
				Title:  "Workspace",
				Fields: projectFieldsForTest(),
			},
		},
	}
}

func projectFieldsForTest() []GitHubProjectField {
	var fields []GitHubProjectField
	for _, input := range projectWorkspaceFieldInputs() {
		field := GitHubProjectField{
			ID:       "PVTF_" + slugify(input.Name),
			Name:     input.Name,
			DataType: input.DataType,
		}
		if len(input.Options) > 0 {
			field.Options = map[string]string{}
			for _, option := range input.Options {
				field.Options[option] = "PVTO_" + slugify(input.Name) + "_" + slugify(option)
			}
		}
		fields = append(fields, field)
	}
	return fields
}

func projectFieldIDsForTest() map[string]string {
	ids := map[string]string{}
	for _, field := range projectFieldsForTest() {
		ids[field.Name] = field.ID
	}
	return ids
}

func trimProjectFieldOptions(project *GitHubProjectWorkspace, fieldName string, keep []string) {
	if project == nil {
		return
	}
	allowed := map[string]bool{}
	for _, value := range keep {
		allowed[value] = true
	}
	for i := range project.Fields {
		if !strings.EqualFold(project.Fields[i].Name, fieldName) {
			continue
		}
		options := map[string]string{}
		for name, id := range project.Fields[i].Options {
			if allowed[name] {
				options[name] = id
			}
		}
		project.Fields[i].Options = options
		return
	}
}
