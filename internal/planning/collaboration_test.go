package planning

import (
	"encoding/json"
	"strings"
	"testing"

	"plan/internal/workspace"
)

func TestAssessCollaborationSourceForLocalBrainstorm(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	if _, err := manager.CreateBrainstorm("Local Promotion"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.UpdateGuidedBrainstormIntake("local-promotion", GuidedBrainstormIntakeInput{
		Vision: "Promote a refined local brainstorm into a single spec when the work stays bounded.",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormRefinement("local-promotion", BrainstormRefinementInput{
		Problem:                "The planner needs a consistent maturity gate before promotion.",
		UserValue:              "The user gets one clean spec instead of guessing the next artifact.",
		Constraints:            "Keep the first slice read-only.\nDo not create GitHub issues during assessment.",
		Appetite:               "One bounded planning pass.",
		RemainingOpenQuestions: "Should the promote command default to json?",
		CandidateApproaches:    "Assess brainstorm maturity.\nGenerate a promotion draft.",
		DecisionSnapshot:       "Start with one spec because the work is still tightly bounded.",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormChallenge("local-promotion", BrainstormChallengeInput{
		RabbitHoles:           "Do not build automatic writes yet.",
		NoGos:                 "No custom collaboration UI.",
		Assumptions:           "The user can review a JSON draft before promotion.",
		LikelyOverengineering: "Building project automation in the first slice.",
		SimplerAlternative:    "Assess first, then draft promotion.",
	}); err != nil {
		t.Fatal(err)
	}

	assessment, err := manager.AssessCollaborationSource(CollaborationAssessInput{
		BrainstormSlug: "local-promotion",
	})
	if err != nil {
		t.Fatal(err)
	}
	if assessment.Kind != maturityAssessmentKind {
		t.Fatalf("unexpected assessment kind: %+v", assessment)
	}
	if assessment.Source.Mode != CollaborationSourceLocal || assessment.Source.EntryMode != EntryModeLocalPromotion {
		t.Fatalf("unexpected source: %+v", assessment.Source)
	}
	if assessment.Ownership.Mode != SourceOfTruthLocal {
		t.Fatalf("expected local ownership by default: %+v", assessment.Ownership)
	}
	if assessment.Decision.State != MaturityReadySingleSpec {
		t.Fatalf("expected single-spec readiness: %+v", assessment.Decision)
	}
	if assessment.Decision.RecommendedPath != PromotionSingleSpec {
		t.Fatalf("expected single-spec path: %+v", assessment.Decision)
	}
	if len(assessment.Decision.SuggestedTitles.Specs) != 1 || assessment.Decision.SuggestedTitles.Specs[0] != "Local Promotion" {
		t.Fatalf("unexpected suggested titles: %+v", assessment.Decision.SuggestedTitles)
	}
}

func TestAssessAndPromoteGitHubDiscussion(t *testing.T) {
	client := &stubGitHubClient{
		preflight: &GitHubRepoInfo{
			Repo:          "JimmyMcBride/plan",
			RepoURL:       "https://github.com/JimmyMcBride/plan",
			DefaultBranch: "develop",
		},
		context: &GitHubContext{
			Repo: GitHubRepoInfo{
				Repo:          "JimmyMcBride/plan",
				RepoURL:       "https://github.com/JimmyMcBride/plan",
				DefaultBranch: "develop",
			},
			CurrentBranch: "develop",
			CurrentSHA:    "abc123",
		},
		discussions: map[int]*GitHubDiscussion{
			49: {
				Number: 49,
				URL:    "https://github.com/JimmyMcBride/plan/discussions/49",
				Title:  "GitHub collaboration foundation",
				Body: strings.Join([]string{
					"## Problem",
					"GitHub-native collaboration needs a disciplined promotion flow.",
					"",
					"## Goals",
					"Create initiative/spec issues from mature discussions.",
					"",
					"## Non-Goals",
					"Do not build a custom UI.",
					"",
					"## Constraints",
					"Keep issue bodies canonical after promotion.",
					"",
					"## Proposed Shape",
					"Use discussions for brainstorming and issues for distilled planning artifacts.",
					"",
					"## Spec Split",
					"- Collaboration entry modes and maturity assessment",
					"- Promotion draft review and issue-body distillation",
					"",
					"Promotion draft review and issue-body distillation depends on Collaboration entry modes and maturity assessment.",
				}, "\n"),
			},
		},
	}
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	t.Cleanup(reset)

	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	assessment, err := manager.AssessCollaborationSource(CollaborationAssessInput{DiscussionRef: "49"})
	if err != nil {
		t.Fatal(err)
	}
	if assessment.Decision.State != MaturityReadyMultiSpec {
		t.Fatalf("expected multi-spec readiness: %+v", assessment.Decision)
	}
	if len(assessment.Decision.SuggestedTitles.Specs) != 2 {
		t.Fatalf("expected two suggested specs: %+v", assessment.Decision.SuggestedTitles)
	}

	draft, err := manager.BuildPromotionDraft(PromotionDraftInput{DiscussionRef: "49"})
	if err != nil {
		t.Fatal(err)
	}
	if draft.ProposedInitiativeIssue == nil {
		t.Fatalf("expected initiative draft: %+v", draft)
	}
	if len(draft.ProposedSpecIssues) != 2 {
		t.Fatalf("expected two spec drafts: %+v", draft)
	}
	if draft.MilestonePlan == nil || !draft.MilestonePlan.Create {
		t.Fatalf("expected milestone plan: %+v", draft)
	}
	if !draft.ConfirmationRequired {
		t.Fatalf("expected explicit confirmation requirement: %+v", draft)
	}
	if draft.ProposedSpecIssues[1].Readiness != ReadinessBlocked {
		t.Fatalf("expected second spec to be blocked by dependency chain: %+v", draft.ProposedSpecIssues[1])
	}

	result, err := manager.ApplyPromotionDraft(PromotionApplyInput{
		DiscussionRef: "49",
		Confirm:       true,
		TargetMode:    SourceOfTruthGitHub,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Initiative == nil || len(result.Specs) != 2 || result.Milestone == nil {
		t.Fatalf("expected promoted GitHub issue set: %+v", result)
	}

	meta, err := ws.ReadWorkspaceMeta()
	if err != nil {
		t.Fatal(err)
	}
	if meta.SourceMode != workspace.SourceOfTruthGitHub {
		t.Fatalf("expected source mode to switch to github: %+v", meta)
	}
	state, err := ws.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Planning) != 3 {
		t.Fatalf("expected initiative plus two specs in GitHub mirror: %+v", state.Planning)
	}
	second := state.Planning["promotion-draft-review-and-issue-body-distillation"]
	if second.ParentIssueNumber == 0 || second.MilestoneNumber == 0 {
		t.Fatalf("expected parent and milestone metadata on spec mirror: %+v", second)
	}
	if len(second.BlockedBy) != 1 || second.BlockedBy[0] != "collaboration-entry-modes-and-maturity-assessment" {
		t.Fatalf("expected blocked-by metadata in mirror: %+v", second)
	}
	if len(client.blockedByEdges) != 1 {
		t.Fatalf("expected one blocked-by edge to be created: %+v", client.blockedByEdges)
	}
}

func TestBuildPromotionDraftNotReadyUsesEmptySpecSliceInJSON(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)
	if _, err := manager.CreateBrainstorm("Not Ready"); err != nil {
		t.Fatal(err)
	}

	draft, err := manager.BuildPromotionDraft(PromotionDraftInput{BrainstormSlug: "not-ready"})
	if err != nil {
		t.Fatal(err)
	}
	if draft.Assessment.State != MaturityNotReady {
		t.Fatalf("expected not-ready draft: %+v", draft)
	}
	if draft.ProposedSpecIssues == nil || len(draft.ProposedSpecIssues) != 0 {
		t.Fatalf("expected empty proposed spec slice: %+v", draft.ProposedSpecIssues)
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"proposed_spec_issues":[]`) {
		t.Fatalf("expected stable empty array in json output: %s", string(raw))
	}
}

func TestApplyPromotionDraftWiresBlockedByAfterAllIssuesExist(t *testing.T) {
	client := &stubGitHubClient{
		preflight: &GitHubRepoInfo{
			Repo:          "JimmyMcBride/plan",
			RepoURL:       "https://github.com/JimmyMcBride/plan",
			DefaultBranch: "develop",
		},
		context: &GitHubContext{
			Repo: GitHubRepoInfo{
				Repo:          "JimmyMcBride/plan",
				RepoURL:       "https://github.com/JimmyMcBride/plan",
				DefaultBranch: "develop",
			},
			CurrentBranch: "develop",
			CurrentSHA:    "abc123",
		},
		discussions: map[int]*GitHubDiscussion{
			77: {
				Number: 77,
				URL:    "https://github.com/JimmyMcBride/plan/discussions/77",
				Title:  "Out of order dependency wiring",
				Body: strings.Join([]string{
					"## Problem",
					"Promotion dependency edges should work even when the dependent spec is listed first.",
					"",
					"## Goals",
					"Create correct blocked-by edges for later-created spec issues.",
					"",
					"## Non-Goals",
					"Do not require the spec list order to match the dependency order.",
					"",
					"## Constraints",
					"Use GitHub issue dependencies.",
					"",
					"## Proposed Shape",
					"Create all spec issues first, then add dependency edges.",
					"",
					"## Spec Split",
					"- Promotion draft review and issue-body distillation",
					"- Collaboration entry modes and maturity assessment",
					"",
					"Promotion draft review and issue-body distillation depends on Collaboration entry modes and maturity assessment.",
				}, "\n"),
			},
		},
	}
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	t.Cleanup(reset)

	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	result, err := manager.ApplyPromotionDraft(PromotionApplyInput{
		DiscussionRef: "77",
		Confirm:       true,
		TargetMode:    SourceOfTruthGitHub,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Specs) != 2 {
		t.Fatalf("expected two promoted specs: %+v", result)
	}
	if len(client.blockedByEdges) != 1 {
		t.Fatalf("expected one blocked-by edge: %+v", client.blockedByEdges)
	}
	firstSpecNumber := result.Specs[0].Number
	secondSpecNumber := result.Specs[1].Number
	if client.blockedByEdges[0] != [2]int{firstSpecNumber, secondSpecNumber} {
		t.Fatalf("expected first created spec to be blocked by second created spec: %+v", client.blockedByEdges)
	}
}

func TestBulletItemsStripsGitHubTaskMarkers(t *testing.T) {
	items := bulletItems(strings.Join([]string{
		"- [ ] First spec",
		"- [x] Second spec",
		"* Third spec",
	}, "\n"))
	if len(items) != 3 {
		t.Fatalf("expected three items: %+v", items)
	}
	if items[0] != "First spec" || items[1] != "Second spec" || items[2] != "Third spec" {
		t.Fatalf("expected task markers to be stripped: %+v", items)
	}
}
