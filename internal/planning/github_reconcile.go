package planning

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"time"

	"plan/internal/notes"
	"plan/internal/workspace"
)

type GitHubReconcileOptions struct {
	UpdateVisible bool
}

type GitHubReconcileResult struct {
	Repo            string
	CurrentBranch   string
	DefaultBranch   string
	UpdatedIssues   []string
	ReadyStories    []string
	BlockedStories  []string
	PlanningPromote bool
}

func (m *Manager) ReconcileGitHubStories(options GitHubReconcileOptions) (*GitHubReconcileResult, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	backend, err := m.storyBackendForInfo()
	if err != nil {
		return nil, err
	}
	if backend != workspace.StoryBackendGitHub {
		return nil, fmt.Errorf("GitHub reconcile is only available when the story backend is set to github")
	}

	state, err := m.readGitHubStateForStories()
	if err != nil {
		return nil, err
	}
	context, err := m.github.CurrentContext(info.ProjectDir)
	if err != nil {
		return nil, err
	}
	result := &GitHubReconcileResult{
		Repo:          state.Repo,
		CurrentBranch: context.CurrentBranch,
		DefaultBranch: context.Repo.DefaultBranch,
	}
	stateChanged := false
	if state.Repo != context.Repo.Repo {
		state.Repo = context.Repo.Repo
		stateChanged = true
	}
	if state.RepoURL != context.Repo.RepoURL {
		state.RepoURL = context.Repo.RepoURL
		stateChanged = true
	}
	if state.DefaultBranch != context.Repo.DefaultBranch {
		state.DefaultBranch = context.Repo.DefaultBranch
		stateChanged = true
	}
	if context.CurrentBranch == context.Repo.DefaultBranch {
		result.PlanningPromote = true
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if stateChanged {
		state.LastUpdatedAt = now
	}

	slugs := make([]string, 0, len(state.Stories))
	for slug := range state.Stories {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)
	milestoneCache := map[string]*int{}

	for _, slug := range slugs {
		record := state.Stories[slug]
		previous := record
		spec, err := notes.Read(filepath.Join(info.SpecsDir, record.Spec+".md"))
		if err != nil {
			return nil, err
		}
		issue, err := m.github.GetIssue(info.ProjectDir, state.Repo, record.IssueNumber)
		if err != nil {
			return nil, err
		}
		record.IssueURL = issue.URL
		record.RemoteState = issue.State
		record.VisibleReadyMarkerSet = containsString(issue.Labels, planIssueReadyLabel) || containsString(issue.Labels, planIssueBlockedLabel)
		if result.PlanningPromote {
			record.PlanningPRMerged = true
			record.DocRefMode = "main"
			record.DocRef = context.Repo.DefaultBranch
		}

		status, ready, _, _, _ := deriveGitHubStoryState(record, state)
		if ready {
			result.ReadyStories = append(result.ReadyStories, slug)
		}
		if status == "blocked" {
			result.BlockedStories = append(result.BlockedStories, slug)
		}

		body := mergeManagedIssueBody(issue.Body, renderGitHubStoryIssueBody(state, &context.Repo, spec, record))
		milestoneNumber, err := m.resolveSpecInitiativeMilestoneCached(info, state.Repo, spec, milestoneCache)
		if err != nil {
			return nil, err
		}
		milestoneChanged, desiredMilestone, clearMilestone := milestoneUpdate(issue.Milestone, milestoneNumber)
		labels := issue.Labels
		if options.UpdateVisible {
			labels = applyDerivedReadyLabels(labels, status, ready)
		}
		if body != issue.Body || !sameStringSlice(labels, issue.Labels) || milestoneChanged {
			updatedIssue, err := m.github.UpdateIssue(info.ProjectDir, state.Repo, record.IssueNumber, GitHubIssueInput{
				Title:          record.Title,
				Body:           body,
				State:          issue.State,
				Labels:         labels,
				Milestone:      desiredMilestone,
				ClearMilestone: clearMilestone,
			})
			if err != nil {
				return nil, err
			}
			record.IssueURL = updatedIssue.URL
			record.RemoteState = updatedIssue.State
			record.VisibleReadyMarkerSet = containsString(labels, planIssueReadyLabel) || containsString(labels, planIssueBlockedLabel)
			result.UpdatedIssues = append(result.UpdatedIssues, fmt.Sprintf("#%d", record.IssueNumber))
		}
		if gitHubStoryRecordChanged(previous, record) {
			record.UpdatedAt = now
			state.Stories[slug] = record
			stateChanged = true
		}
	}

	if stateChanged {
		state.LastReconciled = now
		state.LastUpdatedAt = now
		if err := m.workspace.WriteGitHubState(*state); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func gitHubStoryRecordChanged(before, after workspace.GitHubStoryRecord) bool {
	before.UpdatedAt = ""
	after.UpdatedAt = ""
	return !reflect.DeepEqual(before, after)
}

func sameStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
