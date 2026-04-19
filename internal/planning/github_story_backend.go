package planning

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"plan/internal/notes"
	"plan/internal/workspace"
)

const (
	planIssueBlockStart  = "<!-- plan:start -->"
	planIssueBlockEnd    = "<!-- plan:end -->"
	planIssueMetaPrefix  = "<!-- plan:meta"
	planIssueReadyLabel  = "plan:ready"
	planIssueBlockedLabel = "plan:blocked"
)

func (m *Manager) storyBackendForInfo() (workspace.StoryBackend, error) {
	return m.StoryBackend()
}

func (m *Manager) readGitHubStateForStories() (*workspace.GitHubState, error) {
	state, err := m.workspace.ReadGitHubState()
	if err != nil {
		return nil, err
	}
	if state.Stories == nil {
		state.Stories = map[string]workspace.GitHubStoryRecord{}
	}
	return state, nil
}

func (m *Manager) createGitHubStory(info *workspace.Info, epicSlug, title, description string, criteria, verification, resources []string) (*notes.Note, error) {
	if len(trimmedItems(criteria)) == 0 {
		return nil, fmt.Errorf("at least one acceptance criterion is required")
	}
	if len(trimmedItems(verification)) == 0 {
		return nil, fmt.Errorf("at least one verification step is required")
	}
	epic, err := notes.Read(filepath.Join(info.EpicsDir, slugify(epicSlug)+".md"))
	if err != nil {
		return nil, err
	}
	specSlug := m.specSlugFromEpic(epic)
	spec, err := notes.Read(filepath.Join(info.SpecsDir, specSlug+".md"))
	if err != nil {
		return nil, err
	}
	if status := stringValue(spec.Metadata["status"]); status != "approved" {
		if status == "" {
			status = "draft"
		}
		return nil, fmt.Errorf("spec %s is %q; approve the spec before creating stories", rel(info.ProjectDir, spec.Path), status)
	}

	state, err := m.readGitHubStateForStories()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(state.Repo) == "" || strings.TrimSpace(state.RepoURL) == "" {
		return nil, fmt.Errorf("GitHub story mode is enabled but repo metadata is missing; rerun `plan github enable --project .`")
	}

	storySlug := slugify(title)
	if _, exists := state.Stories[storySlug]; exists {
		return nil, fmt.Errorf("GitHub-backed story already exists for slug %s", storySlug)
	}

	context, err := m.github.CurrentContext(info.ProjectDir)
	if err != nil {
		return nil, err
	}
	if context.Repo.Repo != state.Repo {
		return nil, fmt.Errorf("current repo resolved to %s, but plan workspace is configured for %s; rerun `plan github enable --project .` if the remote changed", context.Repo.Repo, state.Repo)
	}

	record := workspace.GitHubStoryRecord{
		Slug:               storySlug,
		Title:              title,
		Epic:               slugFromPath(epic.Path),
		Spec:               specSlug,
		Status:             "todo",
		Description:        strings.TrimSpace(description),
		AcceptanceCriteria: trimmedItems(criteria),
		Verification:       trimmedItems(verification),
		Resources:          trimmedItems(resources),
		RemoteState:        "open",
		UpdatedAt:          time.Now().UTC().Format(time.RFC3339),
	}
	if strings.TrimSpace(record.Description) == "" {
		record.Description = fmt.Sprintf("Deliver the %q execution slice from the canonical spec.", title)
	}
	if context.CurrentBranch != context.Repo.DefaultBranch {
		if context.PlanningPR == nil {
			return nil, fmt.Errorf("current branch %s has no planning PR; open a planning PR before creating GitHub-backed stories", context.CurrentBranch)
		}
		record.PlanningPRNumber = context.PlanningPR.Number
		record.PlanningPRURL = context.PlanningPR.URL
		record.PlanningPRMerged = context.PlanningPR.IsMerged
		record.DocRefMode = "sha"
		record.DocRef = context.CurrentSHA
	} else {
		record.PlanningPRMerged = true
		record.DocRefMode = "main"
		record.DocRef = context.Repo.DefaultBranch
	}

	body := mergeManagedIssueBody("", renderGitHubStoryIssueBody(state, &context.Repo, epic, spec, record))
	issue, err := m.github.CreateIssue(info.ProjectDir, state.Repo, GitHubIssueInput{
		Title: title,
		Body:  body,
		State: "open",
	})
	if err != nil {
		return nil, err
	}
	record.IssueNumber = issue.Number
	record.IssueURL = issue.URL
	record.RemoteState = issue.State
	record.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	state.Stories[storySlug] = record
	if err := m.workspace.WriteGitHubState(*state); err != nil {
		return nil, err
	}
	return m.readGitHubStoryFromState(info, state, storySlug)
}

func (m *Manager) updateGitHubStory(info *workspace.Info, storySlug string, changes StoryChanges) (*notes.Note, error) {
	state, err := m.readGitHubStateForStories()
	if err != nil {
		return nil, err
	}
	record, ok := state.Stories[slugify(storySlug)]
	if !ok {
		return nil, fmt.Errorf("GitHub-backed story %s does not exist", slugify(storySlug))
	}
	if changes.Status != "" && !isValidStoryStatus(changes.Status) {
		return nil, fmt.Errorf("invalid story status %q", changes.Status)
	}

	if changes.Status != "" {
		record.Status = changes.Status
	}
	record.AcceptanceCriteria = append(record.AcceptanceCriteria, trimmedItems(changes.AddCriteria)...)
	record.Verification = append(record.Verification, trimmedItems(changes.AddVerification)...)
	record.Resources = append(record.Resources, trimmedItems(changes.AddResources)...)
	if changes.SetBlockers != nil {
		record.Dependencies = normalizeStoryRefs(changes.SetBlockers)
	}

	if requiresExecutionExpectations(record.Status) && !gitHubStoryRecordHasExecutionExpectations(record) {
		return nil, fmt.Errorf("story %s needs acceptance criteria and verification steps before it can be marked %q", record.Slug, record.Status)
	}

	context, err := m.github.CurrentContext(info.ProjectDir)
	if err != nil {
		return nil, err
	}
	epic, err := notes.Read(filepath.Join(info.EpicsDir, record.Epic+".md"))
	if err != nil {
		return nil, err
	}
	spec, err := notes.Read(filepath.Join(info.SpecsDir, record.Spec+".md"))
	if err != nil {
		return nil, err
	}
	existingIssue, err := m.github.GetIssue(info.ProjectDir, state.Repo, record.IssueNumber)
	if err != nil {
		return nil, err
	}
	record.RemoteState = existingIssue.State
	managedBody := renderGitHubStoryIssueBody(state, &context.Repo, epic, spec, record)
	body := mergeManagedIssueBody(existingIssue.Body, managedBody)

	issueState := "open"
	if record.Status == "done" {
		issueState = "closed"
	}
	updatedIssue, err := m.github.UpdateIssue(info.ProjectDir, state.Repo, record.IssueNumber, GitHubIssueInput{
		Title:  record.Title,
		Body:   body,
		State:  issueState,
		Labels: existingIssue.Labels,
	})
	if err != nil {
		return nil, err
	}
	record.IssueURL = updatedIssue.URL
	record.RemoteState = updatedIssue.State
	record.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	state.Stories[record.Slug] = record
	if err := m.workspace.WriteGitHubState(*state); err != nil {
		return nil, err
	}
	note, err := m.readGitHubStoryFromState(info, state, record.Slug)
	if err != nil {
		return nil, err
	}
	if err := m.syncSpecStatusForStory(info, note); err != nil {
		return nil, err
	}
	return note, nil
}

func (m *Manager) listGitHubStories(info *workspace.Info, filterEpic, filterStatus string) ([]StoryInfo, error) {
	state, err := m.readGitHubStateForStories()
	if err != nil {
		return nil, err
	}
	specVersions, err := loadSpecTargetVersions(info)
	if err != nil {
		return nil, err
	}

	filterEpic = slugify(filterEpic)
	out := make([]StoryInfo, 0, len(state.Stories))
	for _, record := range state.Stories {
		status, ready, blockedReasons, blockedByPlan, blockedByDeps := deriveGitHubStoryState(record, state)
		item := StoryInfo{
			Slug:          record.Slug,
			Path:          record.IssueURL,
			Title:         record.Title,
			Status:        status,
			Epic:          record.Epic,
			Spec:          record.Spec,
			TargetVersion: specVersions[record.Spec],
			Blockers:      append([]string(nil), record.Dependencies...),
			Backend:       string(workspace.StoryBackendGitHub),
			IssueNumber:   record.IssueNumber,
			IssueURL:      record.IssueURL,
			Ready:         ready,
			BlockedByPlan: blockedByPlan,
			BlockedByDeps: blockedByDeps,
		}
		if len(blockedReasons) > 0 && status == "blocked" && len(item.Blockers) == 0 {
			item.Blockers = append(item.Blockers, blockedReasons...)
		}
		if filterEpic != "" && item.Epic != filterEpic {
			continue
		}
		if filterStatus != "" && item.Status != filterStatus {
			continue
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Ready != out[j].Ready {
			return out[i].Ready
		}
		return out[i].Title < out[j].Title
	})
	return out, nil
}

func (m *Manager) readGitHubStory(info *workspace.Info, storySlug string) (*notes.Note, error) {
	state, err := m.readGitHubStateForStories()
	if err != nil {
		return nil, err
	}
	return m.readGitHubStoryFromState(info, state, storySlug)
}

func (m *Manager) readGitHubStoryFromState(info *workspace.Info, state *workspace.GitHubState, storySlug string) (*notes.Note, error) {
	record, ok := state.Stories[slugify(storySlug)]
	if !ok {
		return nil, fmt.Errorf("GitHub-backed story %s does not exist", slugify(storySlug))
	}
	status, _, _, _, _ := deriveGitHubStoryState(record, state)
	content := renderGitHubStoryNoteContent(state, record)
	return &notes.Note{
		Path:  record.IssueURL,
		Title: record.Title,
		Type:  "story",
		Metadata: map[string]any{
			"slug":     record.Slug,
			"status":   status,
			"epic":     record.Epic,
			"spec":     record.Spec,
			"backend":  string(workspace.StoryBackendGitHub),
			"issue":    record.IssueNumber,
			"issue_url": record.IssueURL,
			"blockers": append([]string(nil), record.Dependencies...),
		},
		Content: content,
	}, nil
}

func gitHubStoryRecordHasExecutionExpectations(record workspace.GitHubStoryRecord) bool {
	return len(trimmedItems(record.AcceptanceCriteria)) > 0 && len(trimmedItems(record.Verification)) > 0
}

func deriveGitHubStoryState(record workspace.GitHubStoryRecord, state *workspace.GitHubState) (status string, ready bool, blockedReasons []string, blockedByPlan bool, blockedByDeps bool) {
	status = record.Status
	if status == "" {
		status = "todo"
	}
	if strings.EqualFold(record.RemoteState, "closed") {
		status = "done"
	}
	if status == "done" {
		return status, false, nil, false, false
	}

	if !record.PlanningPRMerged {
		blockedByPlan = true
		blockedReasons = append(blockedReasons, "planning PR not merged")
	}
	for _, dep := range normalizeStoryRefs(record.Dependencies) {
		dependency, ok := state.Stories[dep]
		if !ok {
			blockedByDeps = true
			blockedReasons = append(blockedReasons, fmt.Sprintf("dependency %s is missing", dep))
			continue
		}
		if dependency.Status != "done" && !strings.EqualFold(dependency.RemoteState, "closed") {
			blockedByDeps = true
			name := dependency.Title
			if strings.TrimSpace(name) == "" {
				name = dep
			}
			blockedReasons = append(blockedReasons, fmt.Sprintf("dependency %s is still open", name))
		}
	}
	if len(blockedReasons) > 0 && status != "in_progress" {
		return "blocked", false, blockedReasons, blockedByPlan, blockedByDeps
	}
	if status == "blocked" && len(blockedReasons) == 0 {
		status = "todo"
	}
	ready = status == "todo" && len(blockedReasons) == 0
	return status, ready, blockedReasons, blockedByPlan, blockedByDeps
}

func renderGitHubStoryIssueBody(state *workspace.GitHubState, repo *GitHubRepoInfo, epic, spec *notes.Note, record workspace.GitHubStoryRecord) string {
	content := renderGitHubStoryNoteContent(state, record)

	epicLink, specLink := gitHubPlanningLinks(repo, record)
	planningLines := []string{
		fmt.Sprintf("- Epic: [%s](%s)", epic.Title, epicLink),
		fmt.Sprintf("- Spec: [%s](%s)", spec.Title, specLink),
	}
	if record.PlanningPRURL != "" {
		planningLines = append(planningLines, fmt.Sprintf("- Planning PR: [#%d](%s)", record.PlanningPRNumber, record.PlanningPRURL))
	}
	if record.PlanningPRMerged {
		planningLines = append(planningLines, "- Planning Merge: merged")
	} else {
		planningLines = append(planningLines, "- Planning Merge: blocked until planning PR merges")
	}

	dependencies := renderGitHubDependenciesSection(state, record)
	asyncNotes := renderListOrPlaceholder(record.AsyncNotes, "- none")
	meta := renderGitHubStoryMetadataBlock(record)

	body := content
	body = notes.SetSection(body, "Planning Links", strings.Join(planningLines, "\n"))
	body = notes.SetSection(body, "Dependencies", dependencies)
	body = notes.SetSection(body, "Async Notes", asyncNotes)

	return strings.Join([]string{
		planIssueBlockStart,
		strings.TrimRight(body, "\n"),
		"",
		meta,
		planIssueBlockEnd,
		"",
	}, "\n")
}

func renderGitHubStoryNoteContent(state *workspace.GitHubState, record workspace.GitHubStoryRecord) string {
	var body strings.Builder
	body.WriteString("## Description\n\n")
	body.WriteString(strings.TrimSpace(record.Description))
	body.WriteString("\n\n## Acceptance Criteria\n\n")
	body.WriteString(renderChecklistOrPlaceholder(record.AcceptanceCriteria))
	body.WriteString("\n\n## Verification\n\n")
	body.WriteString(renderListOrPlaceholder(record.Verification, "- Validate the story against the canonical spec."))
	body.WriteString("\n\n## Critique\n\n")
	body.WriteString(renderStoryCritique(StoryCritique{
		ScopeFit:              record.ScopeFit,
		VerticalSliceCheck:    record.VerticalSliceCheck,
		HiddenPrerequisites:   record.HiddenPrerequisites,
		VerificationGaps:      record.VerificationGaps,
		RewriteRecommendation: record.RewriteRecommendation,
	}))
	body.WriteString("\n\n## Resources\n\n")
	resources := append([]string(nil), record.Resources...)
	if record.IssueURL != "" {
		resources = append([]string{fmt.Sprintf("- [GitHub Issue](%s)", record.IssueURL)}, resources...)
	}
	body.WriteString(renderListOrPlaceholder(resources, "- none"))
	body.WriteString("\n\n## Dependencies\n\n")
	body.WriteString(renderGitHubDependenciesSection(state, record))
	body.WriteString("\n\n## Async Notes\n\n")
	body.WriteString(renderListOrPlaceholder(record.AsyncNotes, "- none"))
	body.WriteString("\n")
	return body.String()
}

func gitHubPlanningLinks(repo *GitHubRepoInfo, record workspace.GitHubStoryRecord) (string, string) {
	ref := repo.DefaultBranch
	if record.DocRefMode == "sha" && strings.TrimSpace(record.DocRef) != "" {
		ref = record.DocRef
	}
	if record.DocRefMode == "main" && strings.TrimSpace(record.DocRef) != "" {
		ref = record.DocRef
	}
	base := strings.TrimSuffix(repo.RepoURL, "/")
	return fmt.Sprintf("%s/blob/%s/.plan/epics/%s.md", base, ref, record.Epic),
		fmt.Sprintf("%s/blob/%s/.plan/specs/%s.md", base, ref, record.Spec)
}

func renderGitHubDependenciesSection(state *workspace.GitHubState, record workspace.GitHubStoryRecord) string {
	if len(record.Dependencies) == 0 {
		return "- none"
	}
	var lines []string
	for _, dep := range normalizeStoryRefs(record.Dependencies) {
		dependency, ok := state.Stories[dep]
		if !ok {
			lines = append(lines, fmt.Sprintf("- [ ] blocked by %s", dep))
			continue
		}
		check := "[ ]"
		label := dependency.Title
		if strings.EqualFold(dependency.RemoteState, "closed") || dependency.Status == "done" {
			check = "[x]"
		}
		if label == "" {
			label = dep
		}
		if dependency.IssueURL != "" {
			lines = append(lines, fmt.Sprintf("- %s [%s](%s)", check, label, dependency.IssueURL))
			continue
		}
		lines = append(lines, fmt.Sprintf("- %s %s", check, label))
	}
	return strings.Join(lines, "\n")
}

func renderChecklistOrPlaceholder(items []string) string {
	items = trimmedItems(items)
	if len(items) == 0 {
		return "- [ ] Add acceptance criteria"
	}
	var lines []string
	for _, item := range items {
		lines = append(lines, checklist(item))
	}
	return strings.Join(lines, "\n")
}

func renderListOrPlaceholder(items []string, placeholder string) string {
	items = trimmedItems(items)
	if len(items) == 0 {
		return placeholder
	}
	var lines []string
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			lines = append(lines, trimmed)
			continue
		}
		lines = append(lines, "- "+trimmed)
	}
	return strings.Join(lines, "\n")
}

func renderGitHubStoryMetadataBlock(record workspace.GitHubStoryRecord) string {
	lines := []string{
		planIssueMetaPrefix,
		fmt.Sprintf("slug: %s", record.Slug),
		fmt.Sprintf("epic: %s", record.Epic),
		fmt.Sprintf("spec: %s", record.Spec),
		fmt.Sprintf("status: %s", record.Status),
		fmt.Sprintf("issue_number: %d", record.IssueNumber),
		fmt.Sprintf("doc_ref_mode: %s", record.DocRefMode),
		fmt.Sprintf("doc_ref: %s", record.DocRef),
		fmt.Sprintf("planning_pr_number: %d", record.PlanningPRNumber),
		fmt.Sprintf("planning_pr_merged: %t", record.PlanningPRMerged),
		fmt.Sprintf("depends_on: %s", strings.Join(normalizeStoryRefs(record.Dependencies), ",")),
		"-->",
	}
	return strings.Join(lines, "\n")
}

func mergeManagedIssueBody(existingBody, managedBody string) string {
	existing := strings.ReplaceAll(existingBody, "\r\n", "\n")
	start := strings.Index(existing, planIssueBlockStart)
	end := strings.Index(existing, planIssueBlockEnd)
	if start >= 0 && end > start {
		end += len(planIssueBlockEnd)
		return strings.TrimRight(existing[:start], "\n") + "\n\n" + strings.TrimRight(managedBody, "\n") + existing[end:]
	}
	if strings.TrimSpace(existing) == "" {
		return managedBody
	}
	return strings.TrimRight(existing, "\n") + "\n\n" + strings.TrimRight(managedBody, "\n") + "\n"
}
