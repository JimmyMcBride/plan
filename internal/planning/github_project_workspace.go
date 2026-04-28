package planning

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"plan/internal/workspace"
)

const (
	projectFieldType   = "Type"
	projectFieldStage  = "Stage"
	projectFieldReady  = "Ready"
	projectFieldStatus = "Status"
	projectFieldArea   = "Area"

	projectValueTracking = "Tracking"
	projectValueSpec     = "Spec"
	projectValueApproved = "Approved"
	projectValueYes      = "Yes"
	projectValueNo       = "No"
	projectValueTodo     = "Todo"
)

type GitHubProjectWorkspaceResult struct {
	Decision              string                    `json:"decision"`
	ProjectOwner          string                    `json:"project_owner"`
	ProjectNumber         int                       `json:"project_number"`
	ProjectID             string                    `json:"project_id"`
	ProjectURL            string                    `json:"project_url"`
	FieldIDs              map[string]string         `json:"field_ids"`
	Items                 []GitHubProjectItemResult `json:"items"`
	SavedViewInstructions []string                  `json:"saved_view_instructions"`
}

type GitHubProjectItemResult struct {
	IssueNumber int               `json:"issue_number"`
	Kind        string            `json:"kind"`
	ItemID      string            `json:"item_id"`
	Values      map[string]string `json:"values"`
}

type preparedProjectWorkspace struct {
	project *GitHubProjectWorkspace
	fields  map[string]GitHubProjectField
	result  *GitHubProjectWorkspaceResult
	area    string
}

func (m *Manager) prepareProjectWorkspace(projectDir, repo string, draft *PromotionDraft, record workspace.GitHubProjectDecisionRecord) (*preparedProjectWorkspace, workspace.GitHubProjectDecisionRecord, error) {
	decision := normalizeProjectDecision(record.Decision)
	if decision == "" || decision == projectDecisionSkip {
		return nil, record, nil
	}
	ref := GitHubProjectReference{
		Owner:  record.ProjectOwner,
		Number: record.ProjectNumber,
		ID:     record.ProjectID,
		URL:    record.ProjectURL,
	}
	if ref.Owner == "" && decision == projectDecisionCreate {
		owner, _, err := splitRepo(repo)
		if err != nil {
			return nil, record, err
		}
		ref.Owner = owner
	}
	title := record.MilestoneTitle
	if strings.TrimSpace(title) == "" && draft != nil && draft.ProposedInitiativeIssue != nil {
		title = draft.ProposedInitiativeIssue.Title
	}
	if strings.TrimSpace(title) == "" && draft != nil && len(draft.ProposedSpecIssues) > 0 {
		title = draft.ProposedSpecIssues[0].Title
	}
	if strings.TrimSpace(title) == "" {
		title = record.Slug
	}

	var (
		project *GitHubProjectWorkspace
		err     error
	)
	switch decision {
	case projectDecisionCreate:
		project, err = m.github.CreateProjectWorkspace(projectDir, repo, GitHubProjectWorkspaceInput{
			Owner: ref.Owner,
			Title: title,
		})
	case projectDecisionConnect:
		project, err = m.github.GetProjectWorkspace(projectDir, repo, ref)
	default:
		return nil, record, fmt.Errorf("unsupported project decision %q", record.Decision)
	}
	if err != nil {
		return nil, record, err
	}
	fields := map[string]GitHubProjectField{}
	fieldIDs := map[string]string{}
	for _, fieldInput := range projectWorkspaceFieldInputs() {
		field, err := m.github.EnsureProjectField(projectDir, *project, fieldInput)
		if err != nil {
			return nil, record, err
		}
		fields[fieldInput.Name] = *field
		fieldIDs[fieldInput.Name] = field.ID
		project.Fields = upsertProjectField(project.Fields, *field)
	}
	record.ProjectOwner = project.Owner
	record.ProjectNumber = project.Number
	record.ProjectID = project.ID
	record.ProjectURL = project.URL
	record.FieldIDs = fieldIDs
	record.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return &preparedProjectWorkspace{
		project: project,
		fields:  fields,
		area:    defaultString(record.InitiativeSlug, record.Slug),
		result: &GitHubProjectWorkspaceResult{
			Decision:              decision,
			ProjectOwner:          project.Owner,
			ProjectNumber:         project.Number,
			ProjectID:             project.ID,
			ProjectURL:            project.URL,
			FieldIDs:              fieldIDs,
			SavedViewInstructions: projectSavedViewInstructions(project.URL),
		},
	}, record, nil
}

func (m *Manager) populateProjectWorkspaceItems(projectDir, repo string, prepared *preparedProjectWorkspace, initiativeIssue *GitHubIssue, specIssuesBySlug map[string]*GitHubIssue, specDrafts []PromotionIssueDraft) error {
	if prepared == nil || prepared.project == nil {
		return nil
	}
	area := defaultString(prepared.area, slugify(prepared.project.Title))
	if initiativeIssue != nil {
		values := map[string]string{
			projectFieldType:  projectValueTracking,
			projectFieldStage: projectValueApproved,
			projectFieldReady: projectValueNo,
			projectFieldArea:  area,
		}
		item, err := m.addProjectIssueItem(projectDir, repo, prepared, initiativeIssue.Number, "initiative", values)
		if err != nil {
			return err
		}
		prepared.result.Items = append(prepared.result.Items, *item)
	}
	for _, draft := range specDrafts {
		issue := specIssuesBySlug[draft.Slug]
		if issue == nil {
			return fmt.Errorf("spec issue for %q was not available for project provisioning", draft.Title)
		}
		ready := projectValueYes
		if draft.Readiness != ReadinessReady || len(draft.BlockedBy) > 0 {
			ready = projectValueNo
		}
		values := map[string]string{
			projectFieldType:   projectValueSpec,
			projectFieldStage:  projectValueApproved,
			projectFieldReady:  ready,
			projectFieldStatus: projectValueTodo,
			projectFieldArea:   area,
		}
		item, err := m.addProjectIssueItem(projectDir, repo, prepared, issue.Number, "spec", values)
		if err != nil {
			return err
		}
		prepared.result.Items = append(prepared.result.Items, *item)
	}
	return nil
}

func (m *Manager) addProjectIssueItem(projectDir, repo string, prepared *preparedProjectWorkspace, issueNumber int, kind string, values map[string]string) (*GitHubProjectItemResult, error) {
	item, err := m.github.AddProjectItemByIssue(projectDir, repo, prepared.project.ID, issueNumber)
	if err != nil {
		return nil, err
	}
	for _, fieldInput := range projectWorkspaceFieldInputs() {
		value := values[fieldInput.Name]
		if strings.TrimSpace(value) == "" {
			continue
		}
		field, ok := prepared.fields[fieldInput.Name]
		if !ok {
			return nil, fmt.Errorf("project field %q was not prepared", fieldInput.Name)
		}
		if err := m.github.SetProjectItemField(projectDir, prepared.project.ID, item.ID, field, value); err != nil {
			return nil, err
		}
	}
	return &GitHubProjectItemResult{
		IssueNumber: issueNumber,
		Kind:        kind,
		ItemID:      item.ID,
		Values:      values,
	}, nil
}

func projectWorkspaceFieldInputs() []GitHubProjectFieldInput {
	return []GitHubProjectFieldInput{
		{Name: projectFieldType, DataType: "SINGLE_SELECT", Options: []string{projectValueTracking, projectValueSpec}},
		{Name: projectFieldStage, DataType: "SINGLE_SELECT", Options: []string{projectValueApproved}},
		{Name: projectFieldReady, DataType: "SINGLE_SELECT", Options: []string{projectValueYes, projectValueNo}},
		{Name: projectFieldStatus, DataType: "SINGLE_SELECT", Options: []string{projectValueTodo}},
		{Name: projectFieldArea, DataType: "TEXT"},
	}
}

func upsertProjectField(fields []GitHubProjectField, field GitHubProjectField) []GitHubProjectField {
	for i := range fields {
		if strings.EqualFold(fields[i].Name, field.Name) {
			fields[i] = field
			return fields
		}
	}
	return append(fields, field)
}

func projectSavedViewInstructions(projectURL string) []string {
	prefix := "Open the provisioned GitHub Project"
	if strings.TrimSpace(projectURL) != "" {
		prefix = "Open " + strings.TrimSpace(projectURL)
	}
	return []string{
		prefix + " and create a Workspace table view for all initiative items.",
		"Create an Execution board filtered to Ready:Yes or execution statuses.",
		"Create an Ideas / Tracking table filtered to Type:Tracking.",
	}
}

func normalizeProjectReference(owner string, number int, projectID, projectURL string) (GitHubProjectReference, error) {
	ref := GitHubProjectReference{
		Owner:  strings.TrimSpace(owner),
		Number: number,
		ID:     strings.TrimSpace(projectID),
		URL:    strings.TrimSpace(projectURL),
	}
	if ref.URL != "" {
		parsed, err := parseGitHubProjectURL(ref.URL)
		if err != nil {
			return GitHubProjectReference{}, err
		}
		if ref.Owner == "" {
			ref.Owner = parsed.Owner
		} else if !strings.EqualFold(ref.Owner, parsed.Owner) {
			return GitHubProjectReference{}, fmt.Errorf("project owner %q does not match project URL owner %q", ref.Owner, parsed.Owner)
		}
		if ref.Number == 0 {
			ref.Number = parsed.Number
		} else if ref.Number != parsed.Number {
			return GitHubProjectReference{}, fmt.Errorf("project number %d does not match project URL number %d", ref.Number, parsed.Number)
		}
	}
	return ref, nil
}

func parseGitHubProjectURL(raw string) (GitHubProjectReference, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return GitHubProjectReference{}, fmt.Errorf("parse GitHub Project URL: %w", err)
	}
	if !strings.EqualFold(parsed.Host, "github.com") && !strings.EqualFold(parsed.Host, "www.github.com") {
		return GitHubProjectReference{}, fmt.Errorf("GitHub Project URL must use github.com")
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 4 || (parts[0] != "users" && parts[0] != "orgs") || parts[1] == "" || parts[2] != "projects" {
		return GitHubProjectReference{}, fmt.Errorf("GitHub Project URL must look like https://github.com/users/<owner>/projects/<number> or https://github.com/orgs/<owner>/projects/<number>, with optional trailing path segments")
	}
	number, err := strconv.Atoi(parts[3])
	if err != nil || number <= 0 {
		return GitHubProjectReference{}, fmt.Errorf("GitHub Project URL has invalid project number %q", parts[3])
	}
	return GitHubProjectReference{
		Owner:  parts[1],
		Number: number,
		URL:    strings.TrimSpace(raw),
	}, nil
}
