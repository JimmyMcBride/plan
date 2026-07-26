package planning

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
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
	if draft.ProposedInitiativeIssue.Action != PromotionActionCreate ||
		draft.ProposedSpecIssues[0].Action != PromotionActionCreate ||
		draft.ProposedSpecIssues[1].Action != PromotionActionCreate ||
		draft.MilestonePlan.Action != PromotionActionCreate {
		t.Fatalf("expected fresh promotion to classify every artifact as create: %+v", draft)
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
	if err == nil || !strings.Contains(err.Error(), "requires either --project-owner and --project-number") {
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
	if result.ProjectDecision.ProjectOwner != "JimmyMcBride" || result.ProjectDecision.ProjectNumber == 0 || result.ProjectDecision.ProjectID == "" || result.ProjectDecision.ProjectURL == "" || len(result.ProjectDecision.FieldIDs) != 5 {
		t.Fatalf("expected project identity after provisioning: %+v", result.ProjectDecision)
	}
	if result.ProjectWorkspace == nil || len(result.ProjectWorkspace.Items) != 7 || len(result.ProjectWorkspace.SavedViewInstructions) == 0 {
		t.Fatalf("expected project workspace provisioning result: %+v", result.ProjectWorkspace)
	}
	if len(client.createdProjects) != 1 || client.createdProjects[0].Title != "Pre-planning Center Product Readiness" {
		t.Fatalf("expected one created project: %+v", client.createdProjects)
	}
	if len(client.projectItems) != 7 {
		t.Fatalf("expected initiative plus spec project items: %+v", client.projectItems)
	}
	if !stubHasProjectValue(client.projectValues, result.Initiative.Number, projectFieldType, projectValueTracking) {
		t.Fatalf("expected initiative type tracking value: %+v", client.projectValues)
	}
	if !stubHasProjectValue(client.projectValues, result.Specs[0].Number, projectFieldReady, projectValueYes) {
		t.Fatalf("expected ready spec value: %+v", client.projectValues)
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

func TestApplyPromotionDraftConnectsExistingProjectWorkspace(t *testing.T) {
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
		projects: map[int]*GitHubProjectWorkspace{
			12: {
				Owner:  "JimmyMcBride",
				Number: 12,
				ID:     "PVT_existing",
				URL:    "https://github.com/users/JimmyMcBride/projects/12",
				Title:  "Existing Workspace",
			},
		},
		discussions: map[int]*GitHubDiscussion{
			90: {
				Number: 90,
				URL:    "https://github.com/JimmyMcBride/plan/discussions/90",
				Title:  "Connected Workspace",
				Body: strings.Join([]string{
					"## Problem",
					"Existing project workspaces need Plan-managed cards.",
					"",
					"## Goals",
					"Connect existing projects and seed initiative/spec items.",
					"",
					"## Non-Goals",
					"Do not create saved views.",
					"",
					"## Constraints",
					"Keep field setup deterministic.",
					"",
					"## Proposed Shape",
					"Use a connected GitHub Project and add issue items.",
					"",
					"## Spec Split",
					"- Project connect schema",
					"- Project item values",
					"",
					"Project item values depends on Project connect schema.",
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
		DiscussionRef:   "90",
		Confirm:         true,
		TargetMode:      SourceOfTruthGitHub,
		ProjectDecision: "connect",
		ProjectOwner:    "JimmyMcBride",
		ProjectNumber:   12,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(client.createdProjects) != 0 {
		t.Fatalf("connect should not create a project: %+v", client.createdProjects)
	}
	if result.ProjectDecision == nil || result.ProjectDecision.Decision != "connect" || result.ProjectDecision.ProjectID != "PVT_existing" || result.ProjectDecision.ProjectNumber != 12 {
		t.Fatalf("expected connected project decision metadata: %+v", result.ProjectDecision)
	}
	if result.ProjectWorkspace == nil || len(result.ProjectWorkspace.Items) != 3 || len(result.ProjectWorkspace.SavedViewInstructions) != 3 {
		t.Fatalf("expected project workspace output: %+v", result.ProjectWorkspace)
	}
	if !stubHasProjectValue(client.projectValues, result.Specs[0].Number, projectFieldReady, projectValueYes) {
		t.Fatalf("expected first spec to be ready: %+v", client.projectValues)
	}
	if !stubHasProjectValue(client.projectValues, result.Specs[1].Number, projectFieldReady, projectValueNo) {
		t.Fatalf("expected blocked spec to be not ready: %+v", client.projectValues)
	}

	state, err := ws.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	record := state.ProjectDecisions["connected-workspace"]
	if record.ProjectID != "PVT_existing" || len(record.FieldIDs) != 5 {
		t.Fatalf("expected persisted project metadata: %+v", record)
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

func TestValidateProjectDecisionRejectsReferenceFlagsWithoutDecision(t *testing.T) {
	draft := &PromotionDraft{
		ProposedSpecIssues: []PromotionIssueDraft{
			{Title: "Small spec", Slug: "small-spec"},
		},
	}
	err := validateProjectDecision("", draft, GitHubProjectReference{Owner: "JimmyMcBride", Number: 12})
	if err == nil || !strings.Contains(err.Error(), "project reference flags require --project-decision") {
		t.Fatalf("expected project reference flag validation error, got %v", err)
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

func stubHasProjectValue(values []stubProjectValue, issueNumber int, field, value string) bool {
	itemID := "PVTI_" + strconv.Itoa(issueNumber)
	for _, current := range values {
		if current.ItemID == itemID && current.Field == field && current.Value == value {
			return true
		}
	}
	return false
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

func TestVoltPromotionPreviewReconcilesExistingInitiativeSpecsAndMilestone(t *testing.T) {
	manager, _, client := newVoltPromotionFixture(t)

	draft, err := manager.BuildPromotionDraft(PromotionDraftInput{DiscussionRef: "1"})
	if err != nil {
		t.Fatal(err)
	}
	if !draft.ConfirmationRequired || draft.ManualFallbackAllowed {
		t.Fatalf("unexpected preview safety flags: %+v", draft)
	}
	if draft.ProposedInitiativeIssue == nil {
		t.Fatalf("expected reconciled initiative: %+v", draft)
	}
	if draft.ProposedInitiativeIssue.IssueNumber != 2 || draft.ProposedInitiativeIssue.Action != PromotionActionUpdate {
		t.Fatalf("expected initiative #2 update: %+v", draft.ProposedInitiativeIssue)
	}
	if draft.MilestonePlan == nil ||
		draft.MilestonePlan.Action != PromotionActionReuse ||
		draft.MilestonePlan.Create ||
		draft.MilestonePlan.Number != 1 ||
		draft.MilestonePlan.Title != "Volt v0 — Evidence-Ready Research Prototype" {
		t.Fatalf("expected milestone #1 reuse: %+v", draft.MilestonePlan)
	}

	expected := []struct {
		number  int
		title   string
		blocked []string
	}{
		{3, "Research evidence and evaluation protocol", nil},
		{4, "Volt v0 language kernel and canonical syntax", []string{"Research evidence and evaluation protocol"}},
		{5, "Reference interpreter, program graph, and DiagnosticV1 protocol", []string{"Research evidence and evaluation protocol", "Volt v0 language kernel and canonical syntax"}},
		{6, "Benchmark corpus and controlled agent study", []string{"Research evidence and evaluation protocol", "Volt v0 language kernel and canonical syntax", "Reference interpreter, program graph, and DiagnosticV1 protocol"}},
	}
	if len(draft.ProposedSpecIssues) != len(expected) {
		t.Fatalf("expected four specs: %+v", draft.ProposedSpecIssues)
	}
	bodies := map[string]struct{}{}
	for i, want := range expected {
		got := draft.ProposedSpecIssues[i]
		if got.IssueNumber != want.number || got.Title != want.title || got.Action != PromotionActionUpdate {
			t.Fatalf("unexpected spec %d reconciliation: %+v", i+1, got)
		}
		if !sameStrings(got.BlockedBy, want.blocked) {
			t.Fatalf("unexpected dependencies for #%d: got=%v want=%v", got.IssueNumber, got.BlockedBy, want.blocked)
		}
		if !strings.Contains(got.Body, "## Acceptance Criteria") || !strings.Contains(got.Body, "## Verification") {
			t.Fatalf("expected acceptance criteria and verification in #%d body:\n%s", got.IssueNumber, got.Body)
		}
		if _, duplicate := bodies[got.Body]; duplicate {
			t.Fatalf("expected materially distinct spec bodies; duplicate body for #%d", got.IssueNumber)
		}
		bodies[got.Body] = struct{}{}
	}
	if !strings.Contains(draft.ProposedSpecIssues[0].Body, "Safe evolution remains explicitly unvalidated.") ||
		!strings.Contains(draft.ProposedSpecIssues[3].Body, "Runs are reproducible from content-addressed pinned inputs.") {
		t.Fatalf("expected per-spec acceptance criteria to survive rendering")
	}
	for _, relationship := range draft.RelationshipPlan {
		if relationship.Action != PromotionActionReuse {
			t.Fatalf("expected existing relationship reuse, got %+v", relationship)
		}
	}
	if client.createIssueCalls != 0 || client.updateIssueCalls != 0 || len(client.labels) != 0 {
		t.Fatalf("preview mutated GitHub stub: creates=%d updates=%d labels=%v", client.createIssueCalls, client.updateIssueCalls, client.labels)
	}
	if len(client.milestones) != 1 || len(client.subIssues) != 4 || len(client.blockedByEdges) != 6 {
		t.Fatalf("preview changed existing artifacts or relationships")
	}
}

func TestVoltPromotionApplyIsIdempotentAndPreservesDependencies(t *testing.T) {
	manager, ws, client := newVoltPromotionFixture(t)

	result, err := manager.ApplyPromotionDraft(PromotionApplyInput{
		DiscussionRef: "1",
		Confirm:       true,
		TargetMode:    SourceOfTruthGitHub,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Initiative == nil || result.Initiative.Number != 2 || len(result.Specs) != 4 {
		t.Fatalf("unexpected reconciliation apply result: %+v", result)
	}
	if client.createIssueCalls != 0 || client.updateIssueCalls != 5 {
		t.Fatalf("expected five existing issue updates and no creates: creates=%d updates=%d", client.createIssueCalls, client.updateIssueCalls)
	}
	if len(client.milestones) != 1 || len(client.subIssues) != 4 || len(client.blockedByEdges) != 6 {
		t.Fatalf("apply duplicated milestone or relationships")
	}
	state, err := ws.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	expectedBlocked := map[string][]string{
		"research-evidence-and-evaluation-protocol":                     nil,
		"volt-v0-language-kernel-and-canonical-syntax":                  {"research-evidence-and-evaluation-protocol"},
		"reference-interpreter-program-graph-and-diagnosticv1-protocol": {"research-evidence-and-evaluation-protocol", "volt-v0-language-kernel-and-canonical-syntax"},
		"benchmark-corpus-and-controlled-agent-study":                   {"research-evidence-and-evaluation-protocol", "volt-v0-language-kernel-and-canonical-syntax", "reference-interpreter-program-graph-and-diagnosticv1-protocol"},
	}
	for slug, blocked := range expectedBlocked {
		if !sameStrings(state.Planning[slug].BlockedBy, blocked) {
			t.Fatalf("unexpected mirrored dependencies for %s: got=%v want=%v", slug, state.Planning[slug].BlockedBy, blocked)
		}
	}

	second, err := manager.ApplyPromotionDraft(PromotionApplyInput{
		DiscussionRef: "1",
		Confirm:       true,
		TargetMode:    SourceOfTruthGitHub,
	})
	if err != nil {
		t.Fatal(err)
	}
	if client.createIssueCalls != 0 || client.updateIssueCalls != 5 {
		t.Fatalf("repeated apply should not create or update unchanged issues: creates=%d updates=%d", client.createIssueCalls, client.updateIssueCalls)
	}
	if len(client.milestones) != 1 || len(client.subIssues) != 4 || len(client.blockedByEdges) != 6 {
		t.Fatalf("repeated apply duplicated milestone or relationships")
	}
	if second.Draft.ProposedInitiativeIssue.Action != PromotionActionUnchanged {
		t.Fatalf("expected second apply preview to classify initiative unchanged: %+v", second.Draft.ProposedInitiativeIssue)
	}
	for _, spec := range second.Draft.ProposedSpecIssues {
		if spec.Action != PromotionActionUnchanged {
			t.Fatalf("expected second apply preview to classify specs unchanged: %+v", spec)
		}
	}
}

func TestPromotionPreviewBlocksAmbiguousPlanIdentity(t *testing.T) {
	manager, ws, client := newVoltPromotionFixture(t)
	duplicate := *client.issues[2]
	duplicate.Number = 20
	duplicate.URL = "https://github.com/JimmyMcBride/volt/issues/20"
	client.issues[20] = &duplicate
	state, err := ws.ReadGitHubState()
	if err != nil {
		t.Fatal(err)
	}
	delete(state.Planning, "volt-language-thesis-and-minimum-semantic-core")
	if err := ws.WriteGitHubState(*state); err != nil {
		t.Fatal(err)
	}

	_, err = manager.BuildPromotionDraft(PromotionDraftInput{DiscussionRef: "1"})
	if err == nil || !strings.Contains(err.Error(), "ambiguous Plan identity") {
		t.Fatalf("expected blocking ambiguity error, got %v", err)
	}
	if client.createIssueCalls != 0 || client.updateIssueCalls != 0 {
		t.Fatalf("ambiguous preview must not mutate GitHub")
	}
}

func TestParsePromotionMapPreservesExplicitVerification(t *testing.T) {
	content := strings.Join([]string{
		"## Promotion map",
		"",
		"Target milestone: **Example milestone**",
		"",
		"### Spec 1 — Independent parser",
		"",
		"Parse one independent brief.",
		"",
		"Scope:",
		"",
		"Only this parser.",
		"",
		"Acceptance criteria:",
		"",
		"- Parser keeps the brief.",
		"",
		"Verification:",
		"",
		"- Run the parser fixture.",
		"",
		"Dependencies: none.",
		"",
		"Readiness: ready.",
	}, "\n")
	briefs, titles, milestone := parsePromotionMap(content)
	if len(titles) != 1 || milestone != "Example milestone" {
		t.Fatalf("unexpected promotion map parse: titles=%v milestone=%q", titles, milestone)
	}
	brief := briefs[normalizePromotionTitle(titles[0])]
	if !sameStrings(brief.AcceptanceCriteria, []string{"Parser keeps the brief."}) ||
		!sameStrings(brief.Verification, []string{"Run the parser fixture."}) ||
		brief.Scope != "Only this parser." ||
		brief.Readiness != ReadinessReady {
		t.Fatalf("unexpected parsed brief: %+v", brief)
	}
}

func newVoltPromotionFixture(t *testing.T) (*Manager, *workspace.Manager, *stubGitHubClient) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "collaboration", "volt-discussion-1.md"))
	if err != nil {
		t.Fatal(err)
	}
	milestone := &GitHubMilestone{Number: 1, Title: "Volt v0 — Evidence-Ready Research Prototype"}
	sourceURL := "https://github.com/JimmyMcBride/volt/discussions/1"
	issue := func(number int, title, kind string, labels ...string) *GitHubIssue {
		return &GitHubIssue{
			Number:    number,
			URL:       "https://github.com/JimmyMcBride/volt/issues/" + strconv.Itoa(number),
			Title:     title,
			Body:      "## Existing " + kind + "\n\n## Source\n\n- [" + sourceURL + "](" + sourceURL + ")",
			State:     "open",
			Labels:    append([]string{"enhancement"}, labels...),
			Milestone: &GitHubMilestone{Number: milestone.Number, Title: milestone.Title},
		}
	}
	client := &stubGitHubClient{
		preflight: &GitHubRepoInfo{
			Repo:          "JimmyMcBride/volt",
			RepoURL:       "https://github.com/JimmyMcBride/volt",
			DefaultBranch: "main",
		},
		context: &GitHubContext{
			Repo: GitHubRepoInfo{
				Repo:          "JimmyMcBride/volt",
				RepoURL:       "https://github.com/JimmyMcBride/volt",
				DefaultBranch: "main",
			},
			CurrentBranch: "main",
			CurrentSHA:    "abc123",
		},
		issues: map[int]*GitHubIssue{
			2: issue(2, "Volt language thesis and minimum semantic core", "initiative", planIssueInitiativeLabel),
			3: issue(3, "Research evidence and evaluation protocol", "spec", planIssueSpecLabel, planIssueReadyLabel),
			4: issue(4, "Volt v0 language kernel and canonical syntax", "spec", planIssueSpecLabel),
			5: issue(5, "Reference interpreter and DiagnosticV1 protocol", "spec", planIssueSpecLabel),
			6: issue(6, "Benchmark corpus and controlled agent study", "spec", planIssueSpecLabel),
		},
		milestones: map[string]*GitHubMilestone{
			milestone.Title: milestone,
		},
		discussions: map[int]*GitHubDiscussion{
			1: {
				Number: 1,
				URL:    sourceURL,
				Title:  "Volt language thesis and minimum semantic core",
				Body:   string(body),
			},
		},
		subIssues:      [][2]int{{2, 3}, {2, 4}, {2, 5}, {2, 6}},
		blockedByEdges: [][2]int{{4, 3}, {5, 3}, {5, 4}, {6, 3}, {6, 4}, {6, 5}},
		nextIssue:      7,
	}
	reset := SetGitHubClientFactoryForTesting(func() GitHubClient { return client })
	t.Cleanup(reset)

	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	meta, err := ws.ReadWorkspaceMeta()
	if err != nil {
		t.Fatal(err)
	}
	meta.SourceMode = workspace.SourceOfTruthGitHub
	if err := ws.WriteWorkspaceMeta(*meta); err != nil {
		t.Fatal(err)
	}
	planning := map[string]workspace.GitHubPlanningRecord{}
	addRecord := func(slug, kind, title string, number, parent int) {
		planning[slug] = workspace.GitHubPlanningRecord{
			Slug:              slug,
			Kind:              kind,
			Title:             title,
			IssueNumber:       number,
			IssueURL:          "https://github.com/JimmyMcBride/volt/issues/" + strconv.Itoa(number),
			RemoteState:       "open",
			Readiness:         "ready",
			OwnershipMode:     "github",
			EntryMode:         "github_collaborative",
			SourceMode:        "github_discussion",
			DiscussionNumber:  1,
			DiscussionURL:     sourceURL,
			ParentIssueNumber: parent,
			MilestoneNumber:   1,
			MilestoneTitle:    milestone.Title,
		}
	}
	addRecord("volt-language-thesis-and-minimum-semantic-core", "initiative", "Volt language thesis and minimum semantic core", 2, 0)
	addRecord("research-evidence-and-evaluation-protocol", "spec", "Research evidence and evaluation protocol", 3, 2)
	addRecord("volt-v0-language-kernel-and-canonical-syntax", "spec", "Volt v0 language kernel and canonical syntax", 4, 2)
	addRecord("reference-interpreter-and-diagnosticv1-protocol", "spec", "Reference interpreter and DiagnosticV1 protocol", 5, 2)
	addRecord("benchmark-corpus-and-controlled-agent-study", "spec", "Benchmark corpus and controlled agent study", 6, 2)
	if err := ws.WriteGitHubState(workspace.GitHubState{
		Repo:          "JimmyMcBride/volt",
		RepoURL:       "https://github.com/JimmyMcBride/volt",
		DefaultBranch: "main",
		Stories:       map[string]workspace.GitHubStoryRecord{},
		Planning:      planning,
	}); err != nil {
		t.Fatal(err)
	}
	return New(ws), ws, client
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
