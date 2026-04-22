package cmd

import (
	"bytes"
	"testing"

	"plan/internal/planning"
	"plan/internal/workspace"
)

type stubGitHubEnableClient struct {
	preflight *planning.GitHubRepoInfo
	err       error
}

func (s *stubGitHubEnableClient) Preflight(projectDir string) (*planning.GitHubRepoInfo, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.preflight, nil
}

func (s *stubGitHubEnableClient) CurrentContext(projectDir string) (*planning.GitHubContext, error) {
	panic("unexpected CurrentContext call")
}

func (s *stubGitHubEnableClient) CreateIssue(projectDir, repo string, input planning.GitHubIssueInput) (*planning.GitHubIssue, error) {
	panic("unexpected CreateIssue call")
}

func (s *stubGitHubEnableClient) UpdateIssue(projectDir, repo string, issueNumber int, input planning.GitHubIssueInput) (*planning.GitHubIssue, error) {
	panic("unexpected UpdateIssue call")
}

func (s *stubGitHubEnableClient) GetIssue(projectDir, repo string, issueNumber int) (*planning.GitHubIssue, error) {
	panic("unexpected GetIssue call")
}

func (s *stubGitHubEnableClient) FindMilestone(projectDir, repo, title string) (*planning.GitHubMilestone, error) {
	panic("unexpected FindMilestone call")
}

func TestGitHubEnableCommandPrintsBackendSummary(t *testing.T) {
	reset := planning.SetGitHubClientFactoryForTesting(func() planning.GitHubClient {
		return &stubGitHubEnableClient{
			preflight: &planning.GitHubRepoInfo{
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

	var buf bytes.Buffer
	command := newRootCmd()
	command.SetOut(&buf)
	command.SetErr(&buf)
	command.SetArgs([]string{"--project", root, "github", "enable"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected github enable to succeed: %v\n%s", err, buf.String())
	}

	output := buf.String()
	if !containsLine(output, "github_backend: github") || !containsLine(output, "repo: JimmyMcBride/plan") || !containsLine(output, "default_branch: main") {
		t.Fatalf("unexpected github enable output:\n%s", output)
	}
}

func containsLine(output, expected string) bool {
	for _, line := range bytes.Split([]byte(output), []byte("\n")) {
		if string(line) == expected {
			return true
		}
	}
	return false
}
