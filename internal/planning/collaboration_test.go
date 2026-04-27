package planning

import (
	"encoding/json"
	"errors"
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

func TestSixSpecProsePromotesAsMultiSpecWithProjectDecision(t *testing.T) {
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
			88: {
				Number: 88,
				URL:    "https://github.com/JimmyMcBride/plan/discussions/88",
				Title:  "Pre-planning Center Product Readiness",
				Body: strings.Join([]string{
					"## Problem",
					"Product readiness work is being shaped inconsistently across operational surfaces.",
					"",
					"## Goals",
					"Create predictable GitHub planning issues for the readiness work.",
					"",
					"## Non-Goals",
					"Do not implement the product surfaces during promotion.",
					"",
					"## Constraints",
					"Keep the promoted issue set milestone-backed and reviewable.",
					"",
					"## Proposed Shape",
					"Create spec issues for: Operational Data UI CRUD, Product Readiness API, Import Pipeline, Permission Model, Audit Trail, Release Coordination",
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

	assessment, err := manager.AssessCollaborationSource(CollaborationAssessInput{DiscussionRef: "88"})
	if err != nil {
		t.Fatal(err)
	}
	if assessment.Decision.State != MaturityReadyMultiSpec {
		t.Fatalf("expected multi-spec readiness: %+v", assessment.Decision)
	}
	if len(assessment.Decision.SuggestedTitles.Specs) != 6 {
		t.Fatalf("expected six specs: %+v", assessment.Decision.SuggestedTitles.Specs)
	}

	draft, err := manager.BuildPromotionDraft(PromotionDraftInput{DiscussionRef: "88"})
	if err != nil {
		t.Fatal(err)
	}
	if draft.ProjectPrompt == nil || !draft.ProjectPrompt.Recommended {
		t.Fatalf("expected project prompt recommendation: %+v", draft.ProjectPrompt)
	}
	if draft.ManualFallbackAllowed {
		t.Fatalf("draft fallback should be disabled by default: %+v", draft)
	}
	if len(draft.AgentPolicy.ForbiddenMutations) == 0 || !containsString(draft.AgentPolicy.ForbiddenMutations, "gh issue create") {
		t.Fatalf("expected hard agent policy: %+v", draft.AgentPolicy)
	}

	_, err = manager.ApplyPromotionDraft(PromotionApplyInput{
		DiscussionRef: "88",
		Confirm:       true,
		TargetMode:    SourceOfTruthGitHub,
	})
	if err == nil || !strings.Contains(err.Error(), "--project-decision create|skip|connect") {
		t.Fatalf("expected project decision gate, got %v", err)
	}

	_, err = manager.ApplyPromotionDraft(PromotionApplyInput{
		DiscussionRef:   "88",
		Confirm:         true,
		TargetMode:      SourceOfTruthGitHub,
		ProjectDecision: "connect",
	})
	if err == nil || !strings.Contains(err.Error(), "requires project reference support") {
		t.Fatalf("expected connect project-reference error, got %v", err)
	}
	if len(client.issues) != 0 {
		t.Fatalf("connect decision should fail before mutating GitHub issues: %+v", client.issues)
	}

	result, err := manager.ApplyPromotionDraft(PromotionApplyInput{
		DiscussionRef:   "88",
		Confirm:         true,
		TargetMode:      SourceOfTruthGitHub,
		ProjectDecision: "create",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Initiative == nil || len(result.Specs) != 6 || result.Milestone == nil {
		t.Fatalf("expected initiative plus six specs and milestone: %+v", result)
	}
	if result.ProjectDecision == nil || result.ProjectDecision.Decision != "create" {
		t.Fatalf("expected project decision record: %+v", result.ProjectDecision)
	}
	if result.ProjectDecision.InitiativeSlug == "" || result.ProjectDecision.InitiativeSlug != result.ProjectDecision.Slug {
		t.Fatalf("expected initiative slug on project decision record: %+v", result.ProjectDecision)
	}
	if result.ProjectDecision.MilestoneNumber != result.Milestone.Number || result.ProjectDecision.MilestoneTitle != result.Milestone.Title {
		t.Fatalf("expected milestone identity on project decision record: %+v", result.ProjectDecision)
	}
	if result.ProjectDecision.ProjectOwner != "" || result.ProjectDecision.ProjectNumber != 0 || result.ProjectDecision.ProjectID != "" || result.ProjectDecision.ProjectURL != "" || len(result.ProjectDecision.FieldIDs) != 0 {
		t.Fatalf("project identity should stay unset until project provisioning: %+v", result.ProjectDecision)
	}
	if !containsString(result.Initiative.Labels, planIssueInitiativeLabel) {
		t.Fatalf("expected initiative label: %+v", result.Initiative.Labels)
	}
	for _, spec := range result.Specs {
		if !containsString(spec.Labels, planIssueSpecLabel) {
			t.Fatalf("expected spec label on %s: %+v", spec.Title, spec.Labels)
		}
	}
	state, err := ws.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Planning) != 7 {
		t.Fatalf("expected initiative plus six specs in metadata: %+v", state.Planning)
	}
	if len(state.ProjectDecisions) != 1 {
		t.Fatalf("expected project decision metadata: %+v", state.ProjectDecisions)
	}
}

func TestAssessBlocksExplicitMultiSpecSourceThatNeedsRepair(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)
	if _, err := manager.CreateBrainstorm("Repair Split"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.UpdateGuidedBrainstormIntake("repair-split", GuidedBrainstormIntakeInput{
		Vision: "Create spec issues for a multi-spec readiness initiative.",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormRefinement("repair-split", BrainstormRefinementInput{
		Problem:             "Agents can collapse multi-spec requests into one issue.",
		UserValue:           "The user gets deterministic promotion structure.",
		Constraints:         "Promotion must fail closed.",
		CandidateApproaches: "Create spec issues for: Operational Data UI CRUD",
		DecisionSnapshot:    "Repair the source split before promotion.",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormChallenge("repair-split", BrainstormChallengeInput{
		NoGos:              "Do not create GitHub issues manually.",
		SimplerAlternative: "Repair the Specs section.",
	}); err != nil {
		t.Fatal(err)
	}

	assessment, err := manager.AssessCollaborationSource(CollaborationAssessInput{BrainstormSlug: "repair-split"})
	if err != nil {
		t.Fatal(err)
	}
	if assessment.Decision.State != MaturityNeedsSourceRepair {
		t.Fatalf("expected source repair state: %+v", assessment.Decision)
	}
	if assessment.Decision.BlockingReason != "requested multi-spec promotion but source parsed as single spec" {
		t.Fatalf("unexpected blocking reason: %+v", assessment.Decision)
	}
	if !strings.Contains(assessment.Decision.NextCommand, "plan discuss repair") {
		t.Fatalf("expected repair command: %+v", assessment.Decision)
	}
}

func TestApplyPromotionDraftEmitsManualFallbackPayloadOnGitHubFailure(t *testing.T) {
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
		createIssueErr: errors.New("api unavailable"),
	}
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	t.Cleanup(reset)

	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)
	createReadyCollaborationBrainstormForTest(t, manager, "fallback-flow")

	_, err := manager.ApplyPromotionDraft(PromotionApplyInput{
		BrainstormSlug: "fallback-flow",
		Confirm:        true,
		TargetMode:     SourceOfTruthGitHub,
	})
	var fallback *PromotionApplyManualFallbackError
	if !errors.As(err, &fallback) {
		t.Fatalf("expected fallback error, got %v", err)
	}
	if fallback.Result == nil || !fallback.Result.ManualFallbackAllowed || fallback.Result.Draft == nil || !fallback.Result.Draft.ManualFallbackAllowed {
		t.Fatalf("expected fallback payload: %+v", fallback.Result)
	}
	if !strings.Contains(fallback.Result.NextCommand, "plan github adopt") {
		t.Fatalf("expected adopt command: %+v", fallback.Result)
	}
}

func TestAdoptGitHubPromotionMirrorsExistingIssues(t *testing.T) {
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
		issues: map[int]*GitHubIssue{
			201: {Number: 201, URL: "https://github.com/JimmyMcBride/plan/issues/201", Title: "Adopt Flow", State: "open"},
			202: {Number: 202, URL: "https://github.com/JimmyMcBride/plan/issues/202", Title: "Adopt schema", State: "open"},
			203: {Number: 203, URL: "https://github.com/JimmyMcBride/plan/issues/203", Title: "Adopt CLI", State: "open"},
		},
		discussions: map[int]*GitHubDiscussion{
			89: {
				Number: 89,
				URL:    "https://github.com/JimmyMcBride/plan/discussions/89",
				Title:  "Adopt Flow",
				Body: strings.Join([]string{
					"## Problem",
					"Manual issue creation needs a Plan-owned recovery path.",
					"",
					"## Goals",
					"Adopt existing issues into metadata.",
					"",
					"## Non-Goals",
					"Do not create unrelated work.",
					"",
					"## Constraints",
					"Validate issue order.",
					"",
					"## Proposed Shape",
					"Use an initiative issue and two specs.",
					"",
					"## Spec Split",
					"- Adopt schema",
					"- Adopt CLI",
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
	result, err := manager.AdoptGitHubPromotion(GitHubAdoptInput{
		DiscussionRef: "89",
		IssueNumbers:  []int{201, 202, 203},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Initiative == nil || len(result.Specs) != 2 || result.Milestone == nil {
		t.Fatalf("expected adopted initiative/spec set: %+v", result)
	}
	if len(client.subIssues) != 2 {
		t.Fatalf("expected adopted sub-issue edges: %+v", client.subIssues)
	}
	state, err := ws.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Planning) != 3 {
		t.Fatalf("expected adopted planning metadata: %+v", state.Planning)
	}
	if !containsString(client.issues[202].Labels, planIssueSpecLabel) {
		t.Fatalf("expected spec labels after adopt: %+v", client.issues[202].Labels)
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

func createReadyCollaborationBrainstormForTest(t *testing.T, manager *Manager, slug string) {
	t.Helper()
	title := strings.ReplaceAll(slug, "-", " ")
	if _, err := manager.CreateBrainstorm(title); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.UpdateGuidedBrainstormIntake(slug, GuidedBrainstormIntakeInput{
		Vision: "Create a reviewed promotion draft.",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormRefinement(slug, BrainstormRefinementInput{
		Problem:             "Promotion needs a deterministic gate.",
		UserValue:           "The user can review before GitHub writes happen.",
		Constraints:         "Use Plan-owned commands.",
		CandidateApproaches: "Build promotion draft review.",
		DecisionSnapshot:    "Promote directly into one spec.",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.UpdateBrainstormChallenge(slug, BrainstormChallengeInput{
		NoGos:              "No manual GitHub creation.",
		SimplerAlternative: "Use discuss promote.",
	}); err != nil {
		t.Fatal(err)
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

func TestExtractSpecCandidatesSupportsExplicitSpecIssuePatterns(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "comma semicolon phrase",
			content: "Create spec issues for: Operational Data UI CRUD, Product Readiness API; Audit Trail",
			want:    []string{"Operational Data UI CRUD", "Product Readiness API", "Audit Trail"},
		},
		{
			name: "numbered list behind explicit intent",
			content: strings.Join([]string{
				"Please create spec issues for:",
				"1. Operational Data UI CRUD",
				"2. Product Readiness API",
			}, "\n"),
			want: []string{"Operational Data UI CRUD", "Product Readiness API"},
		},
		{
			name: "desired outcome bullets",
			content: strings.Join([]string{
				"Create spec issues for this outcome.",
				"",
				"## Desired Outcome",
				"- Operational Data UI CRUD",
				"- Product Readiness API",
			}, "\n"),
			want: []string{"Operational Data UI CRUD", "Product Readiness API"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractSpecCandidates(tc.content)
			if len(got) != len(tc.want) {
				t.Fatalf("expected %d specs, got %+v", len(tc.want), got)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("expected %+v, got %+v", tc.want, got)
				}
			}
		})
	}
}
