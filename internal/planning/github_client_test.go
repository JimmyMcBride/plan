package planning

import "testing"

func TestDecodeGraphQLResponseReturnsErrors(t *testing.T) {
	raw := []byte(`{"errors":[{"message":"discussion access denied"},{"message":"secondary failure"}]}`)
	var target struct{}
	err := decodeGraphQLResponse(raw, &target)
	if err == nil {
		t.Fatal("expected graphql errors to be returned")
	}
	if got := err.Error(); got != "graphql request failed: discussion access denied; secondary failure" {
		t.Fatalf("unexpected graphql error: %v", err)
	}
}

func TestDecodeGraphQLResponseAllowsNilTarget(t *testing.T) {
	raw := []byte(`{"data":{"ok":true}}`)
	if err := decodeGraphQLResponse(raw, nil); err != nil {
		t.Fatalf("expected nil target to succeed: %v", err)
	}
}

func TestParseGitHubIssueKeepsNodeID(t *testing.T) {
	issue, err := parseGitHubIssue([]byte(`{
		"number": 42,
		"node_id": "I_kwExample",
		"html_url": "https://github.com/JimmyMcBride/plan/issues/42",
		"title": "Project item",
		"body": "body",
		"state": "open",
		"labels": []
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if issue.NodeID != "I_kwExample" {
		t.Fatalf("expected node id to round-trip: %+v", issue)
	}
}

func TestParseGitHubIssueListKeepsNodeID(t *testing.T) {
	issues, err := parseGitHubIssueList([]byte(`[{
		"number": 43,
		"id": "I_kwList",
		"url": "https://github.com/JimmyMcBride/plan/issues/43",
		"title": "Listed project item",
		"body": "body",
		"state": "open",
		"labels": []
	}]`))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].NodeID != "I_kwList" {
		t.Fatalf("expected list node id to round-trip: %+v", issues)
	}
}

func TestIssueNodeIDCache(t *testing.T) {
	client := &cliGitHubClient{}
	client.cacheIssueNodeID("JimmyMcBride/plan", &GitHubIssue{Number: 44, NodeID: "I_kwCached"})
	if got := client.cachedIssueNodeID("JimmyMcBride/plan", 44); got != "I_kwCached" {
		t.Fatalf("expected cached issue node id, got %q", got)
	}
}

func TestParseGitHubProjectURLAllowsViewsAndWWW(t *testing.T) {
	ref, err := parseGitHubProjectURL("https://www.github.com/users/JimmyMcBride/projects/12/views/3")
	if err != nil {
		t.Fatal(err)
	}
	if ref.Owner != "JimmyMcBride" || ref.Number != 12 {
		t.Fatalf("unexpected project ref: %+v", ref)
	}
}
