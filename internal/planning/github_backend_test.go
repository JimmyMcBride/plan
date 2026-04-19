package planning

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"plan/internal/workspace"
)

type stubGitHubClient struct {
	preflight    *GitHubRepoInfo
	preflightErr error
}

func (s *stubGitHubClient) Preflight(projectDir string) (*GitHubRepoInfo, error) {
	if s.preflightErr != nil {
		return nil, s.preflightErr
	}
	return s.preflight, nil
}

func (s *stubGitHubClient) CurrentContext(projectDir string) (*GitHubContext, error) {
	panic("unexpected CurrentContext call")
}

func (s *stubGitHubClient) CreateIssue(projectDir, repo string, input GitHubIssueInput) (*GitHubIssue, error) {
	panic("unexpected CreateIssue call")
}

func (s *stubGitHubClient) UpdateIssue(projectDir, repo string, issueNumber int, input GitHubIssueInput) (*GitHubIssue, error) {
	panic("unexpected UpdateIssue call")
}

func (s *stubGitHubClient) GetIssue(projectDir, repo string, issueNumber int) (*GitHubIssue, error) {
	panic("unexpected GetIssue call")
}

func TestEnableGitHubBackendPersistsRepoConfigAfterPreflight(t *testing.T) {
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient {
		return &stubGitHubClient{
			preflight: &GitHubRepoInfo{
				Repo:          "JimmyMcBride/plan",
				RepoURL:       "https://github.com/JimmyMcBride/plan",
				DefaultBranch: "main",
			},
		}
	})
	t.Cleanup(reset)

	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	result, err := manager.EnableGitHubBackend()
	if err != nil {
		t.Fatal(err)
	}
	if result.Backend != "github" || result.Repo != "JimmyMcBride/plan" || result.DefaultBranch != "main" {
		t.Fatalf("unexpected enable result: %+v", result)
	}

	meta, err := ws.ReadWorkspaceMeta()
	if err != nil {
		t.Fatal(err)
	}
	if meta.StoryBackend != workspace.StoryBackendGitHub {
		t.Fatalf("expected GitHub story backend: %+v", meta)
	}

	state, err := ws.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	if state.Repo != "JimmyMcBride/plan" || state.DefaultBranch != "main" {
		t.Fatalf("unexpected github state: %+v", state)
	}
	if state.LastEnabledAt == "" || state.LastUpdatedAt == "" {
		t.Fatalf("expected github state timestamps: %+v", state)
	}
}

func TestEnableGitHubBackendFailsWhenLocalStoriesAlreadyExist(t *testing.T) {
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient {
		return &stubGitHubClient{
			preflight: &GitHubRepoInfo{
				Repo:          "JimmyMcBride/plan",
				RepoURL:       "https://github.com/JimmyMcBride/plan",
				DefaultBranch: "main",
			},
		}
	})
	t.Cleanup(reset)

	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".plan", "stories", "existing.md"), []byte("# Existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	manager := New(ws)
	if _, err := manager.EnableGitHubBackend(); err == nil {
		t.Fatal("expected local stories to block GitHub enablement")
	} else if !strings.Contains(err.Error(), "local story notes still exist") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnableGitHubBackendPropagatesPreflightFailures(t *testing.T) {
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient {
		return &stubGitHubClient{preflightErr: errors.New("gh auth status failed; run `gh auth login` before enabling GitHub mode")}
	})
	t.Cleanup(reset)

	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}

	manager := New(ws)
	if _, err := manager.EnableGitHubBackend(); err == nil {
		t.Fatal("expected preflight error")
	} else if !strings.Contains(err.Error(), "gh auth status failed") {
		t.Fatalf("unexpected preflight error: %v", err)
	}
}
