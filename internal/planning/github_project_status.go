package planning

import (
	"fmt"
	"strings"

	"plan/internal/workspace"
)

type GitHubProjectStatusInput struct {
	IssueNumber int
	Status      string
}

type GitHubProjectStatusResult struct {
	Repo         string `json:"repo"`
	IssueNumber  int    `json:"issue_number"`
	ProjectID    string `json:"project_id"`
	ProjectURL   string `json:"project_url"`
	ItemID       string `json:"item_id"`
	Status       string `json:"status"`
	StatusField  string `json:"status_field"`
	StatusOption string `json:"status_option"`
}

func (m *Manager) SetGitHubProjectIssueStatus(input GitHubProjectStatusInput) (*GitHubProjectStatusResult, error) {
	if input.IssueNumber <= 0 {
		return nil, fmt.Errorf("project status requires --issue")
	}
	status, err := normalizeProjectStatus(input.Status)
	if err != nil {
		return nil, err
	}
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	state, err := m.workspace.ReadGitHubState()
	if err != nil {
		return nil, err
	}
	repo, err := m.githubRepoForState(info.ProjectDir, state)
	if err != nil {
		return nil, err
	}
	record, ok := planningRecordForIssue(state, input.IssueNumber)
	if !ok {
		return nil, fmt.Errorf("issue #%d is not tracked in .plan GitHub planning metadata", input.IssueNumber)
	}
	decision, ok := projectDecisionForPlanningRecord(state, record)
	if !ok {
		return nil, fmt.Errorf("issue #%d has no project decision metadata; rerun promotion/adoption with --project-decision create|connect before setting project status", input.IssueNumber)
	}
	if reason := incompleteProjectDecisionReason(decision); reason != "" {
		return nil, fmt.Errorf("issue #%d has incomplete project decision metadata: %s", input.IssueNumber, reason)
	}
	project, err := m.github.GetProjectWorkspace(info.ProjectDir, repo, projectReferenceFromDecision(decision))
	if err != nil {
		return nil, err
	}
	statusField, ok := projectFieldByName(project.Fields, projectFieldStatus)
	if !ok {
		return nil, fmt.Errorf("GitHub Project %s is missing %q field; run `plan github reconcile --update-visible` to repair safe project drift", projectLabel(project), projectFieldStatus)
	}
	if strings.TrimSpace(statusField.Options[status]) == "" {
		return nil, fmt.Errorf("GitHub Project %s field %q is missing option %q; add it manually or recreate the project because Plan does not edit existing single-select options", projectLabel(project), projectFieldStatus, status)
	}
	item, err := m.github.GetProjectItemByIssue(info.ProjectDir, repo, project.ID, input.IssueNumber)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("issue #%d is missing from GitHub Project %s; run `plan github reconcile --update-visible` to add safe missing issue cards", input.IssueNumber, projectLabel(project))
	}
	if err := m.github.SetProjectItemField(info.ProjectDir, project.ID, item.ID, statusField, status); err != nil {
		return nil, err
	}
	return &GitHubProjectStatusResult{
		Repo:         repo,
		IssueNumber:  input.IssueNumber,
		ProjectID:    project.ID,
		ProjectURL:   project.URL,
		ItemID:       item.ID,
		Status:       status,
		StatusField:  statusField.ID,
		StatusOption: statusField.Options[status],
	}, nil
}

func normalizeProjectStatus(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(strings.ReplaceAll(value, "_", "-"))) {
	case "todo":
		return projectValueTodo, nil
	case "in-progress", "progress":
		return projectValueInProgress, nil
	case "in-review", "review":
		return projectValueInReview, nil
	case "done", "closed", "merged":
		return projectValueDone, nil
	default:
		return "", fmt.Errorf("unsupported project status %q; use todo, in-progress, in-review, or done", value)
	}
}

func (m *Manager) githubRepoForState(projectDir string, state *workspace.GitHubState) (string, error) {
	if state != nil && strings.TrimSpace(state.Repo) != "" {
		return strings.TrimSpace(state.Repo), nil
	}
	context, err := m.github.CurrentContext(projectDir)
	if err != nil {
		return "", err
	}
	return context.Repo.Repo, nil
}

func planningRecordForIssue(state *workspace.GitHubState, issueNumber int) (workspace.GitHubPlanningRecord, bool) {
	if state == nil {
		return workspace.GitHubPlanningRecord{}, false
	}
	for _, record := range state.Planning {
		if record.IssueNumber == issueNumber {
			return record, true
		}
	}
	return workspace.GitHubPlanningRecord{}, false
}

func projectDecisionForPlanningRecord(state *workspace.GitHubState, record workspace.GitHubPlanningRecord) (workspace.GitHubProjectDecisionRecord, bool) {
	if state == nil {
		return workspace.GitHubProjectDecisionRecord{}, false
	}
	if decision, ok := projectDecisionForMilestone(state, record.MilestoneNumber, record.MilestoneTitle); ok {
		return decision, true
	}
	for _, decision := range state.ProjectDecisions {
		if strings.EqualFold(strings.TrimSpace(decision.InitiativeSlug), strings.TrimSpace(record.Slug)) {
			return decision, true
		}
	}
	if record.ParentIssueNumber > 0 {
		parent, ok := planningRecordForIssue(state, record.ParentIssueNumber)
		if ok {
			if decision, ok := projectDecisionForMilestone(state, parent.MilestoneNumber, parent.MilestoneTitle); ok {
				return decision, true
			}
			for _, decision := range state.ProjectDecisions {
				if strings.EqualFold(strings.TrimSpace(decision.InitiativeSlug), strings.TrimSpace(parent.Slug)) {
					return decision, true
				}
			}
		}
	}
	return workspace.GitHubProjectDecisionRecord{}, false
}

func projectReferenceFromDecision(decision workspace.GitHubProjectDecisionRecord) GitHubProjectReference {
	return GitHubProjectReference{
		Owner:  decision.ProjectOwner,
		Number: decision.ProjectNumber,
		ID:     decision.ProjectID,
		URL:    decision.ProjectURL,
	}
}

func projectFieldByName(fields []GitHubProjectField, name string) (GitHubProjectField, bool) {
	for _, field := range fields {
		if strings.EqualFold(strings.TrimSpace(field.Name), strings.TrimSpace(name)) {
			copy := field
			copy.Options = copyStringMap(field.Options)
			return copy, true
		}
	}
	return GitHubProjectField{}, false
}

func projectLabel(project *GitHubProjectWorkspace) string {
	if project == nil {
		return "<unknown>"
	}
	if strings.TrimSpace(project.URL) != "" {
		return strings.TrimSpace(project.URL)
	}
	if project.Number > 0 && strings.TrimSpace(project.Owner) != "" {
		return fmt.Sprintf("%s/%d", project.Owner, project.Number)
	}
	return strings.TrimSpace(project.ID)
}
