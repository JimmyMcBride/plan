package planning

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"plan/internal/workspace"
)

func (m *Manager) checkGitHubProjectWorkspaceDrift(info *workspace.Info, state *workspace.GitHubState, repo string) ([]CheckFinding, error) {
	if state == nil || len(state.ProjectDecisions) == 0 {
		return nil, nil
	}
	var findings []CheckFinding
	for _, decision := range sortedProjectDecisions(state) {
		if normalizeProjectDecision(decision.Decision) == projectDecisionSkip {
			continue
		}
		if reason := incompleteProjectDecisionReason(decision); reason != "" {
			continue
		}
		project, err := m.github.GetProjectWorkspace(info.ProjectDir, repo, projectReferenceFromDecision(decision))
		if err != nil {
			findings = append(findings, projectDriftFinding("github_project.unavailable", decision.ProjectURL, decision.MilestoneTitle, fmt.Sprintf("GitHub Project workspace could not be loaded: %v.", err), "Repair project decision metadata or reconnect the project before project automation continues."))
			continue
		}
		findings = append(findings, checkProjectFields(project)...)
		for _, record := range projectPlanningRecords(state, decision) {
			issue, err := m.github.GetIssue(info.ProjectDir, repo, record.IssueNumber)
			if err != nil {
				return nil, err
			}
			item, err := m.github.GetProjectItemByIssue(info.ProjectDir, repo, project.ID, record.IssueNumber)
			if err != nil {
				return nil, err
			}
			if item == nil {
				findings = append(findings, projectDriftFinding("github_project.missing_item", issue.URL, issue.Title, fmt.Sprintf("Issue #%d is missing from GitHub Project %s.", record.IssueNumber, projectLabel(project)), "Run `plan github reconcile --update-visible` to add safe missing issue cards."))
				continue
			}
			for field, expected := range expectedProjectValues(record, decision, issue, item) {
				if strings.TrimSpace(expected) == "" {
					continue
				}
				if got := strings.TrimSpace(item.Values[field]); got != expected {
					findings = append(findings, projectDriftFinding("github_project.stale_item_field", issue.URL, issue.Title, fmt.Sprintf("Issue #%d Project field %q is %q; expected %q.", record.IssueNumber, field, got, expected), "Run `plan github reconcile --update-visible` to repair safe stale Project item fields."))
				}
			}
		}
	}
	return findings, nil
}

func (m *Manager) reconcileGitHubProjectWorkspaceDrift(info *workspace.Info, state *workspace.GitHubState, repo string, updateVisible bool) ([]string, bool, error) {
	if !updateVisible || state == nil || len(state.ProjectDecisions) == 0 {
		return nil, false, nil
	}
	var updated []string
	stateChanged := false
	for key, decision := range state.ProjectDecisions {
		if normalizeProjectDecision(decision.Decision) == projectDecisionSkip {
			continue
		}
		if reason := incompleteProjectDecisionReason(decision); reason != "" {
			return nil, false, fmt.Errorf("project decision %q is incomplete: %s", key, reason)
		}
		project, err := m.github.GetProjectWorkspace(info.ProjectDir, repo, projectReferenceFromDecision(decision))
		if err != nil {
			return nil, false, err
		}
		fields := map[string]GitHubProjectField{}
		fieldIDs := copyStringMap(decision.FieldIDs)
		decisionChanged := false
		for _, input := range projectWorkspaceFieldInputs() {
			field, err := m.github.EnsureProjectField(info.ProjectDir, *project, input)
			if err != nil {
				return nil, false, err
			}
			fields[input.Name] = *field
			project.Fields = upsertProjectField(project.Fields, *field)
			if fieldIDs[input.Name] != field.ID {
				fieldIDs[input.Name] = field.ID
				decisionChanged = true
			}
		}
		if decisionChanged {
			decision.FieldIDs = fieldIDs
			decision.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			state.ProjectDecisions[key] = decision
			stateChanged = true
		}
		for _, record := range projectPlanningRecords(state, decision) {
			issue, err := m.github.GetIssue(info.ProjectDir, repo, record.IssueNumber)
			if err != nil {
				return nil, false, err
			}
			item, err := m.github.GetProjectItemByIssue(info.ProjectDir, repo, project.ID, record.IssueNumber)
			if err != nil {
				return nil, false, err
			}
			if item == nil {
				item, err = m.github.AddProjectItemByIssue(info.ProjectDir, repo, project.ID, record.IssueNumber)
				if err != nil {
					return nil, false, err
				}
			}
			itemChanged := false
			for fieldName, expected := range expectedProjectValues(record, decision, issue, item) {
				if strings.TrimSpace(expected) == "" || strings.TrimSpace(item.Values[fieldName]) == expected {
					continue
				}
				field, ok := fields[fieldName]
				if !ok {
					return nil, false, fmt.Errorf("project field %q was not prepared", fieldName)
				}
				if err := m.github.SetProjectItemField(info.ProjectDir, project.ID, item.ID, field, expected); err != nil {
					return nil, false, err
				}
				item.Values[fieldName] = expected
				itemChanged = true
			}
			if itemChanged {
				updated = append(updated, fmt.Sprintf("#%d", record.IssueNumber))
			}
		}
	}
	return updated, stateChanged, nil
}

func sortedProjectDecisions(state *workspace.GitHubState) []workspace.GitHubProjectDecisionRecord {
	keys := make([]string, 0, len(state.ProjectDecisions))
	for key := range state.ProjectDecisions {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]workspace.GitHubProjectDecisionRecord, 0, len(keys))
	for _, key := range keys {
		out = append(out, state.ProjectDecisions[key])
	}
	return out
}

func checkProjectFields(project *GitHubProjectWorkspace) []CheckFinding {
	var findings []CheckFinding
	for _, input := range projectWorkspaceFieldInputs() {
		field, ok := projectFieldByName(project.Fields, input.Name)
		if !ok {
			findings = append(findings, projectDriftFinding("github_project.missing_field", project.URL, project.Title, fmt.Sprintf("GitHub Project %s is missing field %q.", projectLabel(project), input.Name), "Run `plan github reconcile --update-visible` to create safe missing Project fields."))
			continue
		}
		for _, option := range missingProjectFieldOptions(field, input.Options) {
			findings = append(findings, projectDriftFinding("github_project.missing_field_option", project.URL, project.Title, fmt.Sprintf("GitHub Project field %q is missing option %q.", input.Name, option), "Add the option manually or recreate the project; Plan does not edit existing single-select options because GitHub regenerates option ids."))
		}
	}
	return findings
}

func projectPlanningRecords(state *workspace.GitHubState, decision workspace.GitHubProjectDecisionRecord) []workspace.GitHubPlanningRecord {
	var records []workspace.GitHubPlanningRecord
	for _, record := range state.Planning {
		if record.IssueNumber == 0 {
			continue
		}
		if decision.MilestoneNumber > 0 && record.MilestoneNumber == decision.MilestoneNumber {
			records = append(records, record)
			continue
		}
		if strings.TrimSpace(decision.MilestoneTitle) != "" && strings.EqualFold(strings.TrimSpace(record.MilestoneTitle), strings.TrimSpace(decision.MilestoneTitle)) {
			records = append(records, record)
			continue
		}
		if strings.TrimSpace(decision.InitiativeSlug) != "" && strings.EqualFold(strings.TrimSpace(record.Slug), strings.TrimSpace(decision.InitiativeSlug)) {
			records = append(records, record)
		}
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].IssueNumber < records[j].IssueNumber
	})
	return records
}

func expectedProjectValues(record workspace.GitHubPlanningRecord, decision workspace.GitHubProjectDecisionRecord, issue *GitHubIssue, item *GitHubProjectItem) map[string]string {
	values := map[string]string{
		projectFieldStage: projectValueApproved,
		projectFieldArea:  defaultString(decision.InitiativeSlug, decision.Slug),
	}
	switch strings.ToLower(strings.TrimSpace(record.Kind)) {
	case "initiative":
		values[projectFieldType] = projectValueTracking
		values[projectFieldReady] = projectValueNo
	case "spec":
		values[projectFieldType] = projectValueSpec
		ready := projectValueYes
		if !strings.EqualFold(strings.TrimSpace(record.Readiness), string(ReadinessReady)) || len(record.BlockedBy) > 0 {
			ready = projectValueNo
		}
		values[projectFieldReady] = ready
		if issue != nil && strings.EqualFold(strings.TrimSpace(issue.State), "closed") {
			values[projectFieldStatus] = projectValueDone
		} else if item == nil || strings.TrimSpace(item.Values[projectFieldStatus]) == "" {
			values[projectFieldStatus] = projectValueTodo
		}
	}
	return values
}

func projectDriftFinding(rule, path, title, message, suggestion string) CheckFinding {
	return CheckFinding{
		Severity:      "error",
		Rule:          rule,
		ArtifactType:  "github_project",
		ArtifactPath:  path,
		ArtifactTitle: title,
		Section:       "GitHub Project",
		Message:       message,
		Suggestion:    suggestion,
	}
}
