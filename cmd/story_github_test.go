package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/planning"
	"plan/internal/workspace"
)

type stubGitHubStoryClient struct {
	preflight  *planning.GitHubRepoInfo
	context    *planning.GitHubContext
	issues     map[int]*planning.GitHubIssue
	nextIssue  int
}

func (s *stubGitHubStoryClient) Preflight(projectDir string) (*planning.GitHubRepoInfo, error) {
	return s.preflight, nil
}

func (s *stubGitHubStoryClient) CurrentContext(projectDir string) (*planning.GitHubContext, error) {
	return s.context, nil
}

func (s *stubGitHubStoryClient) CreateIssue(projectDir, repo string, input planning.GitHubIssueInput) (*planning.GitHubIssue, error) {
	if s.issues == nil {
		s.issues = map[int]*planning.GitHubIssue{}
	}
	if s.nextIssue == 0 {
		s.nextIssue = 101
	}
	issue := &planning.GitHubIssue{
		Number: s.nextIssue,
		URL:    fmt.Sprintf("https://github.com/%s/issues/%d", repo, s.nextIssue),
		Title:  input.Title,
		Body:   input.Body,
		State:  "open",
	}
	s.issues[issue.Number] = issue
	s.nextIssue++
	return issue, nil
}

func (s *stubGitHubStoryClient) UpdateIssue(projectDir, repo string, issueNumber int, input planning.GitHubIssueInput) (*planning.GitHubIssue, error) {
	issue := s.issues[issueNumber]
	issue.Title = input.Title
	issue.Body = input.Body
	if strings.TrimSpace(input.State) != "" {
		issue.State = input.State
	}
	return issue, nil
}

func (s *stubGitHubStoryClient) GetIssue(projectDir, repo string, issueNumber int) (*planning.GitHubIssue, error) {
	copy := *s.issues[issueNumber]
	return &copy, nil
}

func TestStoryCommandsSupportGitHubBackedStories(t *testing.T) {
	client := &stubGitHubStoryClient{
		preflight: &planning.GitHubRepoInfo{
			Repo:          "JimmyMcBride/plan",
			RepoURL:       "https://github.com/JimmyMcBride/plan",
			DefaultBranch: "main",
		},
		context: &planning.GitHubContext{
			Repo: planning.GitHubRepoInfo{
				Repo:          "JimmyMcBride/plan",
				RepoURL:       "https://github.com/JimmyMcBride/plan",
				DefaultBranch: "main",
			},
			CurrentBranch: "main",
			CurrentSHA:    "abc123def456",
		},
	}
	reset := planning.SetGitHubClientFactoryForTesting(func() planning.GitHubClient { return client })
	t.Cleanup(reset)

	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := planning.New(ws)
	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}
	specBody := strings.Join([]string{
		"# Billing Spec",
		"",
		"Created: now",
		"",
		"## Why",
		"",
		"Billing work needs GitHub-backed execution.",
		"",
		"## Problem",
		"",
		"Stories should live in GitHub issues.",
		"",
		"## Goals",
		"",
		"- create issue-backed stories",
		"",
		"## Non-Goals",
		"",
		"- tracker parity",
		"",
		"## Constraints",
		"",
		"- keep epics and specs local",
		"",
		"## Solution Shape",
		"",
		"Store execution work in GitHub issues.",
		"",
		"## Flows",
		"",
		"1. Enable GitHub mode.",
		"2. Create issue-backed stories.",
		"",
		"## Data / Interfaces",
		"",
		"- GitHub issue body contract",
		"",
		"## Risks / Open Questions",
		"",
		"- none",
		"",
		"## Rollout",
		"",
		"- dogfood locally",
		"",
		"## Verification",
		"",
		"- run GitHub story command tests",
		"",
	}, "\n")
	if _, err := manager.UpdateSpec("billing", notes.UpdateInput{
		Body: &specBody,
		Metadata: map[string]any{"status": "approved"},
	}); err != nil {
		t.Fatal(err)
	}

	command := newRootCmd()
	command.SetArgs([]string{"--project", root, "github", "enable"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}

	var createOut bytes.Buffer
	command = newRootCmd()
	command.SetOut(&createOut)
	command.SetErr(&createOut)
	command.SetArgs([]string{
		"--project", root, "story", "create", "billing", "Implement invoices",
		"--body", "Create invoice generation flow",
		"--criteria", "Generate invoices from line items",
		"--verify", "Run focused billing tests",
	})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected GitHub story create to succeed: %v\n%s", err, createOut.String())
	}
	if !strings.Contains(createOut.String(), "Created story https://github.com/JimmyMcBride/plan/issues/101") {
		t.Fatalf("unexpected GitHub story create output:\n%s", createOut.String())
	}

	var listOut bytes.Buffer
	command = newRootCmd()
	command.SetOut(&listOut)
	command.SetErr(&listOut)
	command.SetArgs([]string{"--project", root, "story", "list"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected GitHub story list to succeed: %v\n%s", err, listOut.String())
	}
	if !strings.Contains(listOut.String(), "Implement invoices [todo] epic=billing spec=billing") {
		t.Fatalf("unexpected GitHub story list output:\n%s", listOut.String())
	}

	var showOut bytes.Buffer
	command = newRootCmd()
	command.SetOut(&showOut)
	command.SetErr(&showOut)
	command.SetArgs([]string{"--project", root, "story", "show", "implement-invoices"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected GitHub story show to succeed: %v\n%s", err, showOut.String())
	}
	if !strings.Contains(showOut.String(), "https://github.com/JimmyMcBride/plan/issues/101") || !strings.Contains(showOut.String(), "## Dependencies") {
		t.Fatalf("unexpected GitHub story show output:\n%s", showOut.String())
	}
}
