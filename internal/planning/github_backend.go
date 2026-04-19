package planning

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"plan/internal/workspace"
)

type GitHubEnableResult struct {
	Backend       string
	Repo          string
	RepoURL       string
	DefaultBranch string
	StatePath     string
}

func (m *Manager) StoryBackend() (workspace.StoryBackend, error) {
	meta, err := m.workspace.ReadWorkspaceMeta()
	if err != nil {
		return "", err
	}
	if meta.StoryBackend == "" {
		return workspace.StoryBackendLocal, nil
	}
	return meta.StoryBackend, nil
}

func (m *Manager) EnableGitHubBackend() (*GitHubEnableResult, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	if hasLocalStoryNotes(info.StoriesDir) {
		return nil, fmt.Errorf("cannot enable GitHub story mode while local story notes still exist under %s; keep local mode or migrate those stories first", rel(info.ProjectDir, info.StoriesDir))
	}

	repo, err := m.github.Preflight(info.ProjectDir)
	if err != nil {
		return nil, err
	}
	meta, err := m.workspace.ReadWorkspaceMeta()
	if err != nil {
		return nil, err
	}
	meta.StoryBackend = workspace.StoryBackendGitHub
	meta.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := m.workspace.WriteWorkspaceMeta(*meta); err != nil {
		return nil, err
	}

	state, err := m.workspace.ReadGitHubState()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	state.Repo = repo.Repo
	state.RepoURL = repo.RepoURL
	state.DefaultBranch = repo.DefaultBranch
	state.LastEnabledAt = now
	state.LastUpdatedAt = now
	if state.Stories == nil {
		state.Stories = map[string]workspace.GitHubStoryRecord{}
	}
	if err := m.workspace.WriteGitHubState(*state); err != nil {
		return nil, err
	}

	return &GitHubEnableResult{
		Backend:       string(workspace.StoryBackendGitHub),
		Repo:          repo.Repo,
		RepoURL:       repo.RepoURL,
		DefaultBranch: repo.DefaultBranch,
		StatePath:     rel(info.ProjectDir, info.GitHubFile),
	}, nil
}

func hasLocalStoryNotes(storiesDir string) bool {
	entries, err := os.ReadDir(storiesDir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
			return true
		}
	}
	return false
}
