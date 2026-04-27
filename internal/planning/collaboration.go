package planning

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"plan/internal/notes"
	"plan/internal/workspace"
)

const (
	CollaborationSchemaVersion = 1
	maturityAssessmentKind     = "maturity_assessment"
	promotionDraftKind         = "promotion_draft"
)

var safeShellArgPattern = regexp.MustCompile(`^[A-Za-z0-9_./:=+-]+$`)

type SourceOfTruthMode = workspace.SourceOfTruthMode

const (
	SourceOfTruthLocal  = workspace.SourceOfTruthLocal
	SourceOfTruthGitHub = workspace.SourceOfTruthGitHub
	SourceOfTruthHybrid = workspace.SourceOfTruthHybrid
)

type CollaborationSourceMode string

const (
	CollaborationSourceLocal            CollaborationSourceMode = "local"
	CollaborationSourceGitHubDiscussion CollaborationSourceMode = "github_discussion"
)

type CollaborationEntryMode string

const (
	EntryModeLocalPromotion      CollaborationEntryMode = "local_promotion"
	EntryModeGitHubCollaborative CollaborationEntryMode = "github_collaborative"
)

type MaturityState string

const (
	MaturityNotReady          MaturityState = "not_ready"
	MaturityNeedsSourceRepair MaturityState = "needs_source_repair"
	MaturityReadySingleSpec   MaturityState = "ready_single_spec"
	MaturityReadyMultiSpec    MaturityState = "ready_multi_spec"
)

type MaturityConfidence string

const (
	MaturityConfidenceLow    MaturityConfidence = "low"
	MaturityConfidenceMedium MaturityConfidence = "medium"
	MaturityConfidenceHigh   MaturityConfidence = "high"
)

type PromotionPath string

const (
	PromotionSingleSpec PromotionPath = "single_spec"
	PromotionMultiSpec  PromotionPath = "multi_spec"
)

type ReadinessState string

const (
	ReadinessClarifying      ReadinessState = "clarifying"
	ReadinessReady           ReadinessState = "ready"
	ReadinessBlocked         ReadinessState = "blocked"
	ReadinessNeedsRefinement ReadinessState = "needs-refinement"
	ReadinessDone            ReadinessState = "done"
)

type CollaborationSourceRef struct {
	Mode            CollaborationSourceMode `json:"mode"`
	EntryMode       CollaborationEntryMode  `json:"entry_mode"`
	BrainstormSlug  string                  `json:"brainstorm_slug,omitempty"`
	BrainstormPath  string                  `json:"brainstorm_path,omitempty"`
	Discussion      *GitHubDiscussionRef    `json:"discussion,omitempty"`
	SourceLinks     []string                `json:"source_links,omitempty"`
	CanonicalSource string                  `json:"canonical_source"`
}

type GitHubDiscussionRef struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
	Title  string `json:"title"`
}

type MaturityDependencyGuess struct {
	Spec      string   `json:"spec"`
	BlockedBy []string `json:"blocked_by"`
}

type MaturitySuggestedTitles struct {
	Initiative string   `json:"initiative,omitempty"`
	Specs      []string `json:"specs,omitempty"`
}

type CollaborationAssessment struct {
	SchemaVersion int                           `json:"schema_version"`
	Kind          string                        `json:"kind"`
	GeneratedAt   string                        `json:"generated_at"`
	Source        CollaborationSourceRef        `json:"source"`
	Ownership     CollaborationOwnership        `json:"ownership"`
	Decision      CollaborationMaturityDecision `json:"maturity_decision"`
}

type CollaborationMaturityDecision struct {
	State            MaturityState             `json:"state"`
	Confidence       MaturityConfidence        `json:"confidence"`
	SourceMode       CollaborationSourceMode   `json:"source_mode"`
	Reason           string                    `json:"reason"`
	BlockingReason   string                    `json:"blocking_reason,omitempty"`
	Strengths        []string                  `json:"strengths,omitempty"`
	Gaps             []string                  `json:"gaps,omitempty"`
	RecommendedPath  PromotionPath             `json:"recommended_path,omitempty"`
	SuggestedTitles  MaturitySuggestedTitles   `json:"suggested_titles"`
	DependencyGuess  []MaturityDependencyGuess `json:"dependency_guess,omitempty"`
	RepairCandidates []string                  `json:"repair_candidates,omitempty"`
	NextCommand      string                    `json:"next_command,omitempty"`
}

type CollaborationOwnership struct {
	Mode                      SourceOfTruthMode      `json:"mode"`
	EntryMode                 CollaborationEntryMode `json:"entry_mode"`
	CanonicalDiscussion       string                 `json:"canonical_discussion"`
	CanonicalPlanningArtifact string                 `json:"canonical_planning_artifact"`
	ReadinessSource           string                 `json:"readiness_source"`
	WriteTargets              []string               `json:"write_targets,omitempty"`
	MirrorLocalMeta           bool                   `json:"mirror_local_meta"`
}

type PromotionDraft struct {
	SchemaVersion             int                            `json:"schema_version"`
	Kind                      string                         `json:"kind"`
	GeneratedAt               string                         `json:"generated_at"`
	Source                    CollaborationSourceRef         `json:"source"`
	Ownership                 CollaborationOwnership         `json:"ownership"`
	Assessment                CollaborationMaturityDecision  `json:"assessment"`
	PromotionDecision         PromotionPath                  `json:"promotion_decision,omitempty"`
	WhyThisPath               string                         `json:"why_this_path,omitempty"`
	ProposedInitiativeIssue   *PromotionIssueDraft           `json:"proposed_initiative_issue,omitempty"`
	ProposedSpecIssues        []PromotionIssueDraft          `json:"proposed_spec_issues"`
	ParentSubIssuePlan        []string                       `json:"parent_sub_issue_plan,omitempty"`
	DependencyPlan            []PromotionDependencyPlan      `json:"dependency_plan,omitempty"`
	MilestonePlan             *PromotionMilestonePlan        `json:"milestone_plan,omitempty"`
	ProjectPrompt             *PromotionProjectPrompt        `json:"project_prompt,omitempty"`
	AgentPolicy               PromotionAgentPolicy           `json:"agent_policy"`
	ManualFallbackAllowed     bool                           `json:"manual_fallback_allowed"`
	NeedsRefinementExceptions []PromotionRefinementException `json:"needs_refinement_exceptions,omitempty"`
	ConfirmationRequired      bool                           `json:"confirmation_required"`
}

type PromotionAgentPolicy struct {
	AllowedMutations   []string `json:"allowed_mutations"`
	ForbiddenMutations []string `json:"forbidden_mutations"`
	ManualFallback     string   `json:"manual_fallback"`
}

type PromotionIssueDraft struct {
	Kind           string         `json:"kind"`
	Title          string         `json:"title"`
	Body           string         `json:"body"`
	Slug           string         `json:"slug"`
	Readiness      ReadinessState `json:"readiness"`
	Labels         []string       `json:"labels,omitempty"`
	SourceLinks    []string       `json:"source_links,omitempty"`
	BlockedBy      []string       `json:"blocked_by,omitempty"`
	ReadyByDefault bool           `json:"ready_by_default"`
}

type PromotionDependencyPlan struct {
	IssueTitle string   `json:"issue_title"`
	BlockedBy  []string `json:"blocked_by"`
}

type PromotionMilestonePlan struct {
	Create bool   `json:"create"`
	Title  string `json:"title,omitempty"`
	Why    string `json:"why,omitempty"`
}

type PromotionProjectPrompt struct {
	Recommended bool   `json:"recommended"`
	Reason      string `json:"reason,omitempty"`
}

const (
	projectDecisionCreate  = "create"
	projectDecisionSkip    = "skip"
	projectDecisionConnect = "connect"
)

type PromotionRefinementException struct {
	IssueTitle               string   `json:"issue_title"`
	Gap                      string   `json:"gap"`
	WhyNotReady              string   `json:"why_not_ready"`
	RecommendedClarification string   `json:"recommended_clarification"`
	ExitCriteria             []string `json:"exit_criteria,omitempty"`
}

func defaultPromotionAgentPolicy() PromotionAgentPolicy {
	return PromotionAgentPolicy{
		AllowedMutations:   []string{"plan discuss promote --apply"},
		ForbiddenMutations: []string{"gh issue create", "gh label create", "gh milestone create"},
		ManualFallback:     "only after plan emits manual_fallback_allowed=true",
	}
}

type CollaborationAssessInput struct {
	BrainstormSlug string
	DiscussionRef  string
}

type PromotionDraftInput struct {
	BrainstormSlug string
	DiscussionRef  string
}

type PromotionApplyInput struct {
	BrainstormSlug  string
	DiscussionRef   string
	Confirm         bool
	TargetMode      SourceOfTruthMode
	ProjectDecision string
	ProjectOwner    string
	ProjectNumber   int
	ProjectID       string
	ProjectURL      string
}

type PromotionApplyResult struct {
	Draft                 *PromotionDraft                        `json:"draft"`
	Initiative            *GitHubIssue                           `json:"initiative,omitempty"`
	Specs                 []GitHubIssue                          `json:"specs,omitempty"`
	Milestone             *GitHubMilestone                       `json:"milestone,omitempty"`
	ParentIssue           int                                    `json:"parent_issue,omitempty"`
	ProjectDecision       *workspace.GitHubProjectDecisionRecord `json:"project_decision,omitempty"`
	ProjectWorkspace      *GitHubProjectWorkspaceResult          `json:"project_workspace,omitempty"`
	ManualFallbackAllowed bool                                   `json:"manual_fallback_allowed"`
	ManualFallbackReason  string                                 `json:"manual_fallback_reason,omitempty"`
	NextCommand           string                                 `json:"next_command,omitempty"`
}

type PromotionApplyManualFallbackError struct {
	Result *PromotionApplyResult
	Err    error
}

func (e *PromotionApplyManualFallbackError) Error() string {
	if e == nil || e.Err == nil {
		return "promotion apply failed; manual fallback is allowed"
	}
	return e.Err.Error()
}

func (e *PromotionApplyManualFallbackError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type RepairSpecSplitInput struct {
	BrainstormSlug string
	DiscussionRef  string
	Specs          []string
	Confirm        bool
}

type RepairSpecSplitResult struct {
	SchemaVersion int                    `json:"schema_version"`
	Kind          string                 `json:"kind"`
	GeneratedAt   string                 `json:"generated_at"`
	Source        CollaborationSourceRef `json:"source"`
	Specs         []string               `json:"specs"`
	UpdatedPath   string                 `json:"updated_path,omitempty"`
	UpdatedURL    string                 `json:"updated_url,omitempty"`
	NextCommands  []string               `json:"next_commands"`
}

type GitHubAdoptInput struct {
	BrainstormSlug  string
	DiscussionRef   string
	IssueNumbers    []int
	ProjectDecision string
	ProjectOwner    string
	ProjectNumber   int
	ProjectID       string
	ProjectURL      string
}

type GitHubAdoptResult struct {
	Draft            *PromotionDraft                        `json:"draft"`
	Initiative       *GitHubIssue                           `json:"initiative,omitempty"`
	Specs            []GitHubIssue                          `json:"specs,omitempty"`
	Milestone        *GitHubMilestone                       `json:"milestone,omitempty"`
	ProjectDecision  *workspace.GitHubProjectDecisionRecord `json:"project_decision,omitempty"`
	ProjectWorkspace *GitHubProjectWorkspaceResult          `json:"project_workspace,omitempty"`
	Adopted          []workspace.GitHubPlanningRecord       `json:"adopted"`
}

type collaborationSourceData struct {
	source                  CollaborationSourceRef
	title                   string
	body                    string
	problem                 string
	goals                   string
	constraints             string
	nonGoals                string
	shape                   string
	openQs                  string
	suggested               []string
	deps                    []MaturityDependencyGuess
	explicitMultiSpecIntent bool
	repairCandidates        []string
}

func (m *Manager) AssessCollaborationSource(input CollaborationAssessInput) (*CollaborationAssessment, error) {
	data, ownership, err := m.loadCollaborationSourceData(input.BrainstormSlug, input.DiscussionRef)
	if err != nil {
		return nil, err
	}
	decision := assessCollaborationData(data)
	return &CollaborationAssessment{
		SchemaVersion: CollaborationSchemaVersion,
		Kind:          maturityAssessmentKind,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Source:        data.source,
		Ownership:     ownership,
		Decision:      decision,
	}, nil
}

func (m *Manager) BuildPromotionDraft(input PromotionDraftInput) (*PromotionDraft, error) {
	data, ownership, err := m.loadCollaborationSourceData(input.BrainstormSlug, input.DiscussionRef)
	if err != nil {
		return nil, err
	}
	decision := assessCollaborationData(data)
	draft := &PromotionDraft{
		SchemaVersion:        CollaborationSchemaVersion,
		Kind:                 promotionDraftKind,
		GeneratedAt:          time.Now().UTC().Format(time.RFC3339),
		Source:               data.source,
		Ownership:            ownership,
		Assessment:           decision,
		ProposedSpecIssues:   []PromotionIssueDraft{},
		AgentPolicy:          defaultPromotionAgentPolicy(),
		ConfirmationRequired: true,
	}
	if decision.State == MaturityNotReady || decision.State == MaturityNeedsSourceRepair {
		return draft, nil
	}
	draft.PromotionDecision = decision.RecommendedPath
	draft.WhyThisPath = decision.Reason
	specTitles := append([]string(nil), decision.SuggestedTitles.Specs...)
	if len(specTitles) == 0 {
		specTitles = []string{fallbackSpecTitle(data.title)}
	}
	if draft.PromotionDecision == PromotionSingleSpec {
		refinementExceptions := buildRefinementExceptions(data, specTitles)
		specDraft := buildPromotionSpecDraft(data, specTitles[0], nil, refinementExceptions[specTitles[0]])
		if exception, ok := refinementExceptions[specTitles[0]]; ok {
			draft.NeedsRefinementExceptions = append(draft.NeedsRefinementExceptions, exception)
		}
		draft.ProposedSpecIssues = []PromotionIssueDraft{specDraft}
		return draft, nil
	}
	initiativeTitle := strings.TrimSpace(decision.SuggestedTitles.Initiative)
	if initiativeTitle == "" {
		initiativeTitle = data.title
	}
	initDraft := buildPromotionInitiativeDraft(data, initiativeTitle, specTitles)
	draft.ProposedInitiativeIssue = &initDraft
	draft.ParentSubIssuePlan = []string{
		fmt.Sprintf("%s becomes the parent issue.", initiativeTitle),
		"Each promoted spec issue becomes a direct sub-issue.",
	}
	for _, dep := range decision.DependencyGuess {
		draft.DependencyPlan = append(draft.DependencyPlan, PromotionDependencyPlan{
			IssueTitle: dep.Spec,
			BlockedBy:  append([]string(nil), dep.BlockedBy...),
		})
	}
	refinementExceptions := buildRefinementExceptions(data, specTitles)
	for _, title := range specTitles {
		blockedBy := findBlockedByForTitle(decision.DependencyGuess, title)
		specDraft := buildPromotionSpecDraft(data, title, blockedBy, refinementExceptions[title])
		draft.ProposedSpecIssues = append(draft.ProposedSpecIssues, specDraft)
		if exception, ok := refinementExceptions[title]; ok {
			draft.NeedsRefinementExceptions = append(draft.NeedsRefinementExceptions, exception)
		}
	}
	draft.MilestonePlan = &PromotionMilestonePlan{
		Create: true,
		Title:  initiativeTitle,
		Why:    "Multi-spec promotion always gets a milestone.",
	}
	if shouldRecommendProject(len(specTitles), draft.DependencyPlan) {
		draft.ProjectPrompt = &PromotionProjectPrompt{
			Recommended: true,
			Reason:      "This promotion crosses the coordination threshold for project tracking.",
		}
	} else {
		draft.ProjectPrompt = &PromotionProjectPrompt{
			Recommended: false,
			Reason:      "Milestone tracking is enough for this spec count and dependency shape.",
		}
	}
	return draft, nil
}

func (m *Manager) RepairSpecSplit(input RepairSpecSplitInput) (*RepairSpecSplitResult, error) {
	data, _, err := m.loadCollaborationSourceData(input.BrainstormSlug, input.DiscussionRef)
	if err != nil {
		return nil, err
	}
	specs := dedupeTitles(input.Specs)
	if len(specs) == 0 {
		specs = dedupeTitles(data.repairCandidates)
	}
	if len(specs) < 2 {
		return nil, fmt.Errorf("repair requires at least two spec titles; pass repeated --spec values")
	}
	section := renderSpecSplitSection(specs)
	result := &RepairSpecSplitResult{
		SchemaVersion: CollaborationSchemaVersion,
		Kind:          "source_repair",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Source:        data.source,
		Specs:         specs,
		NextCommands: []string{
			refreshAssessCommand(data.source),
			refreshPromoteCommand(data.source),
		},
	}
	switch data.source.Mode {
	case CollaborationSourceGitHubDiscussion:
		if !input.Confirm {
			return nil, fmt.Errorf("GitHub Discussion repair requires --confirm")
		}
		info, err := m.workspace.EnsureInitialized()
		if err != nil {
			return nil, err
		}
		context, err := m.github.CurrentContext(info.ProjectDir)
		if err != nil {
			return nil, err
		}
		discussionNumber := discussionNumber(data.source.Discussion)
		discussion, err := m.github.GetDiscussion(info.ProjectDir, context.Repo.Repo, discussionNumber)
		if err != nil {
			return nil, err
		}
		body := notes.SetSection(discussion.Body, "Specs", section)
		updated, err := m.github.UpdateDiscussionBody(info.ProjectDir, context.Repo.Repo, discussionNumber, body)
		if err != nil {
			return nil, err
		}
		result.UpdatedURL = updated.URL
	default:
		info, err := m.workspace.EnsureInitialized()
		if err != nil {
			return nil, err
		}
		path := filepath.Join(info.BrainstormsDir, data.source.BrainstormSlug+".md")
		note, err := notes.Read(path)
		if err != nil {
			return nil, err
		}
		body := notes.SetSection(note.Content, "Specs", section)
		if _, err := notes.Update(note.Path, notes.UpdateInput{Body: &body}); err != nil {
			return nil, err
		}
		result.UpdatedPath = rel(info.ProjectDir, note.Path)
	}
	return result, nil
}

func (m *Manager) ApplyPromotionDraft(input PromotionApplyInput) (*PromotionApplyResult, error) {
	if !input.Confirm {
		return nil, fmt.Errorf("promotion apply requires explicit confirmation; rerun with --confirm")
	}
	draft, err := m.BuildPromotionDraft(PromotionDraftInput{
		BrainstormSlug: input.BrainstormSlug,
		DiscussionRef:  input.DiscussionRef,
	})
	if err != nil {
		return nil, err
	}
	if draft.Assessment.State != MaturityReadySingleSpec && draft.Assessment.State != MaturityReadyMultiSpec {
		if draft.Assessment.BlockingReason != "" {
			return nil, fmt.Errorf("source is not ready for promotion: %s", draft.Assessment.BlockingReason)
		}
		return nil, fmt.Errorf("source is not ready for promotion")
	}
	projectRef, err := normalizeProjectReference(input.ProjectOwner, input.ProjectNumber, input.ProjectID, input.ProjectURL)
	if err != nil {
		return nil, err
	}
	if err := validateProjectDecision(input.ProjectDecision, draft, projectRef); err != nil {
		return nil, err
	}
	mode := input.TargetMode
	if mode == "" {
		mode = draft.Ownership.Mode
	}
	if mode == SourceOfTruthLocal {
		return nil, fmt.Errorf("local promotion apply is not implemented yet; use `plan discuss promote --format json` or target github/hybrid")
	}
	if mode != SourceOfTruthGitHub && mode != SourceOfTruthHybrid {
		return nil, fmt.Errorf("unsupported promotion target mode %q", mode)
	}
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	context, err := m.github.CurrentContext(info.ProjectDir)
	if err != nil {
		return nil, err
	}
	meta, err := m.workspace.ReadWorkspaceMeta()
	if err != nil {
		return nil, err
	}
	meta.SourceMode = mode
	meta.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := m.workspace.WriteWorkspaceMeta(*meta); err != nil {
		return nil, err
	}
	state, err := m.workspace.ReadGitHubState()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(state.Repo) == "" {
		state.Repo = context.Repo.Repo
	}
	if strings.TrimSpace(state.RepoURL) == "" {
		state.RepoURL = context.Repo.RepoURL
	}
	if strings.TrimSpace(state.DefaultBranch) == "" {
		state.DefaultBranch = context.Repo.DefaultBranch
	}
	if state.Planning == nil {
		state.Planning = map[string]workspace.GitHubPlanningRecord{}
	}
	if state.ProjectDecisions == nil {
		state.ProjectDecisions = map[string]workspace.GitHubProjectDecisionRecord{}
	}
	result := &PromotionApplyResult{Draft: draft}
	fallback := func(err error) (*PromotionApplyResult, error) {
		return nil, promotionManualFallbackError(draft, result, err)
	}
	if err := m.ensurePromotionLabels(info.ProjectDir, state.Repo, draft); err != nil {
		return fallback(err)
	}
	var milestone *GitHubMilestone
	if draft.MilestonePlan != nil && draft.MilestonePlan.Create {
		milestone, err = m.ensureMilestone(info.ProjectDir, state.Repo, draft.MilestonePlan.Title)
		if err != nil {
			return fallback(err)
		}
		result.Milestone = milestone
	}
	var projectRecord *workspace.GitHubProjectDecisionRecord
	var preparedProject *preparedProjectWorkspace
	if strings.TrimSpace(input.ProjectDecision) != "" {
		record := buildProjectDecisionRecord(draft, input.ProjectDecision, milestone, projectRef)
		preparedProject, record, err = m.prepareProjectWorkspace(info.ProjectDir, state.Repo, draft, record)
		if err != nil {
			return fallback(err)
		}
		state.ProjectDecisions[record.Slug] = record
		result.ProjectDecision = &record
		projectRecord = &record
		if preparedProject != nil {
			result.ProjectWorkspace = preparedProject.result
		}
	}
	var initiativeIssue *GitHubIssue
	if draft.ProposedInitiativeIssue != nil {
		initInput := GitHubIssueInput{
			Title:  draft.ProposedInitiativeIssue.Title,
			Body:   draft.ProposedInitiativeIssue.Body,
			State:  "open",
			Labels: append([]string(nil), draft.ProposedInitiativeIssue.Labels...),
		}
		if milestone != nil {
			initInput.Milestone = &milestone.Number
		}
		initIssue, err := m.github.CreateIssue(info.ProjectDir, state.Repo, initInput)
		if err != nil {
			return fallback(err)
		}
		initiativeIssue = initIssue
		result.Initiative = initIssue
		result.ParentIssue = initIssue.Number
		record := workspace.GitHubPlanningRecord{
			Slug:            draft.ProposedInitiativeIssue.Slug,
			Kind:            "initiative",
			Title:           initIssue.Title,
			IssueNumber:     initIssue.Number,
			IssueURL:        initIssue.URL,
			RemoteState:     initIssue.State,
			Readiness:       string(ReadinessReady),
			OwnershipMode:   string(mode),
			EntryMode:       string(draft.Source.EntryMode),
			SourceMode:      string(draft.Source.Mode),
			MilestoneNumber: milestoneNumberOrZero(milestone),
			MilestoneTitle:  milestoneTitle(milestone),
			UpdatedAt:       time.Now().UTC().Format(time.RFC3339),
		}
		record.DiscussionNumber = discussionNumber(draft.Source.Discussion)
		record.DiscussionURL = discussionURL(draft.Source.Discussion)
		state.Planning[record.Slug] = record
	}
	specIssuesBySlug := make(map[string]*GitHubIssue, len(draft.ProposedSpecIssues))
	specDraftsBySlug := make(map[string]PromotionIssueDraft, len(draft.ProposedSpecIssues))
	for _, specDraft := range draft.ProposedSpecIssues {
		specInput := GitHubIssueInput{
			Title:  specDraft.Title,
			Body:   specDraft.Body,
			State:  "open",
			Labels: append([]string(nil), specDraft.Labels...),
		}
		if milestone != nil {
			specInput.Milestone = &milestone.Number
		}
		specIssue, err := m.github.CreateIssue(info.ProjectDir, state.Repo, specInput)
		if err != nil {
			return fallback(err)
		}
		if initiativeIssue != nil {
			if err := m.github.AddSubIssue(info.ProjectDir, state.Repo, initiativeIssue.Number, specIssue.Number); err != nil {
				return fallback(err)
			}
		}
		specIssuesBySlug[specDraft.Slug] = specIssue
		specDraftsBySlug[specDraft.Slug] = specDraft
		result.Specs = append(result.Specs, *specIssue)
	}
	for slug, specIssue := range specIssuesBySlug {
		specDraft := specDraftsBySlug[slug]
		for _, dep := range specDraft.BlockedBy {
			depIssue, ok := specIssuesBySlug[slugify(dep)]
			if !ok {
				return nil, fmt.Errorf("promotion dependency %q for %q was not created in this promotion set", dep, specDraft.Title)
			}
			if err := m.github.AddBlockedBy(info.ProjectDir, state.Repo, specIssue.Number, depIssue.Number); err != nil {
				return fallback(err)
			}
		}
	}
	if preparedProject != nil {
		if err := m.populateProjectWorkspaceItems(info.ProjectDir, state.Repo, preparedProject, initiativeIssue, specIssuesBySlug, draft.ProposedSpecIssues); err != nil {
			return fallback(err)
		}
		if projectRecord != nil {
			projectRecord.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			state.ProjectDecisions[projectRecord.Slug] = *projectRecord
			result.ProjectDecision = projectRecord
		}
		result.ProjectWorkspace = preparedProject.result
	}
	for slug, specIssue := range specIssuesBySlug {
		specDraft := specDraftsBySlug[slug]
		readiness := string(specDraft.Readiness)
		record := workspace.GitHubPlanningRecord{
			Slug:              specDraft.Slug,
			Kind:              "spec",
			Title:             specIssue.Title,
			IssueNumber:       specIssue.Number,
			IssueURL:          specIssue.URL,
			RemoteState:       specIssue.State,
			Readiness:         readiness,
			OwnershipMode:     string(mode),
			EntryMode:         string(draft.Source.EntryMode),
			SourceMode:        string(draft.Source.Mode),
			ParentIssueNumber: issueNumberOrZero(initiativeIssue),
			MilestoneNumber:   milestoneNumberOrZero(milestone),
			MilestoneTitle:    milestoneTitle(milestone),
			BlockedBy:         slugs(specDraft.BlockedBy),
			UpdatedAt:         time.Now().UTC().Format(time.RFC3339),
		}
		record.DiscussionNumber = discussionNumber(draft.Source.Discussion)
		record.DiscussionURL = discussionURL(draft.Source.Discussion)
		state.Planning[record.Slug] = record
	}
	state.LastUpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := m.workspace.WriteGitHubState(*state); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *Manager) AdoptGitHubPromotion(input GitHubAdoptInput) (*GitHubAdoptResult, error) {
	draft, err := m.BuildPromotionDraft(PromotionDraftInput{
		BrainstormSlug: input.BrainstormSlug,
		DiscussionRef:  input.DiscussionRef,
	})
	if err != nil {
		return nil, err
	}
	if draft.Assessment.State != MaturityReadySingleSpec && draft.Assessment.State != MaturityReadyMultiSpec {
		if draft.Assessment.BlockingReason != "" {
			return nil, fmt.Errorf("source is not ready for adoption: %s", draft.Assessment.BlockingReason)
		}
		return nil, fmt.Errorf("source is not ready for adoption")
	}
	projectRef, err := normalizeProjectReference(input.ProjectOwner, input.ProjectNumber, input.ProjectID, input.ProjectURL)
	if err != nil {
		return nil, err
	}
	if err := validateProjectDecision(input.ProjectDecision, draft, projectRef); err != nil {
		return nil, err
	}
	expected := len(draft.ProposedSpecIssues)
	if draft.ProposedInitiativeIssue != nil {
		expected++
	}
	if len(input.IssueNumbers) != expected {
		return nil, fmt.Errorf("adopt requires %d issue numbers in draft order; got %d", expected, len(input.IssueNumbers))
	}
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	context, err := m.github.CurrentContext(info.ProjectDir)
	if err != nil {
		return nil, err
	}
	meta, err := m.workspace.ReadWorkspaceMeta()
	if err != nil {
		return nil, err
	}
	mode := meta.SourceMode
	if mode == "" || mode == SourceOfTruthLocal {
		mode = SourceOfTruthGitHub
	}
	meta.SourceMode = mode
	meta.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := m.workspace.WriteWorkspaceMeta(*meta); err != nil {
		return nil, err
	}
	state, err := m.workspace.ReadGitHubState()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(state.Repo) == "" {
		state.Repo = context.Repo.Repo
	}
	if strings.TrimSpace(state.RepoURL) == "" {
		state.RepoURL = context.Repo.RepoURL
	}
	if strings.TrimSpace(state.DefaultBranch) == "" {
		state.DefaultBranch = context.Repo.DefaultBranch
	}
	if state.Planning == nil {
		state.Planning = map[string]workspace.GitHubPlanningRecord{}
	}
	if state.ProjectDecisions == nil {
		state.ProjectDecisions = map[string]workspace.GitHubProjectDecisionRecord{}
	}
	if err := m.ensurePromotionLabels(info.ProjectDir, state.Repo, draft); err != nil {
		return nil, err
	}
	var milestone *GitHubMilestone
	if draft.MilestonePlan != nil && draft.MilestonePlan.Create {
		milestone, err = m.ensureMilestone(info.ProjectDir, state.Repo, draft.MilestonePlan.Title)
		if err != nil {
			return nil, err
		}
	}
	result := &GitHubAdoptResult{Draft: draft, Milestone: milestone}
	var projectRecord *workspace.GitHubProjectDecisionRecord
	var preparedProject *preparedProjectWorkspace
	if strings.TrimSpace(input.ProjectDecision) != "" {
		record := buildProjectDecisionRecord(draft, input.ProjectDecision, milestone, projectRef)
		preparedProject, record, err = m.prepareProjectWorkspace(info.ProjectDir, state.Repo, draft, record)
		if err != nil {
			return nil, err
		}
		state.ProjectDecisions[record.Slug] = record
		result.ProjectDecision = &record
		projectRecord = &record
		if preparedProject != nil {
			result.ProjectWorkspace = preparedProject.result
		}
	}

	index := 0
	var initiativeIssue *GitHubIssue
	if draft.ProposedInitiativeIssue != nil {
		issue, err := m.adoptPromotionIssue(info.ProjectDir, state.Repo, input.IssueNumbers[index], *draft.ProposedInitiativeIssue, milestone)
		if err != nil {
			return nil, err
		}
		initiativeIssue = issue
		result.Initiative = issue
		record := workspace.GitHubPlanningRecord{
			Slug:            draft.ProposedInitiativeIssue.Slug,
			Kind:            "initiative",
			Title:           issue.Title,
			IssueNumber:     issue.Number,
			IssueURL:        issue.URL,
			RemoteState:     issue.State,
			Readiness:       string(ReadinessReady),
			OwnershipMode:   string(mode),
			EntryMode:       string(draft.Source.EntryMode),
			SourceMode:      string(draft.Source.Mode),
			MilestoneNumber: milestoneNumberOrZero(milestone),
			MilestoneTitle:  milestoneTitle(milestone),
			UpdatedAt:       time.Now().UTC().Format(time.RFC3339),
		}
		record.DiscussionNumber = discussionNumber(draft.Source.Discussion)
		record.DiscussionURL = discussionURL(draft.Source.Discussion)
		state.Planning[record.Slug] = record
		result.Adopted = append(result.Adopted, record)
		index++
	}

	specIssuesBySlug := make(map[string]*GitHubIssue, len(draft.ProposedSpecIssues))
	specDraftsBySlug := make(map[string]PromotionIssueDraft, len(draft.ProposedSpecIssues))
	for _, specDraft := range draft.ProposedSpecIssues {
		issue, err := m.adoptPromotionIssue(info.ProjectDir, state.Repo, input.IssueNumbers[index], specDraft, milestone)
		if err != nil {
			return nil, err
		}
		if initiativeIssue != nil {
			if err := m.github.AddSubIssue(info.ProjectDir, state.Repo, initiativeIssue.Number, issue.Number); err != nil {
				return nil, err
			}
		}
		specIssuesBySlug[specDraft.Slug] = issue
		specDraftsBySlug[specDraft.Slug] = specDraft
		result.Specs = append(result.Specs, *issue)
		index++
	}
	for slug, specIssue := range specIssuesBySlug {
		specDraft := specDraftsBySlug[slug]
		for _, dep := range specDraft.BlockedBy {
			depIssue, ok := specIssuesBySlug[slugify(dep)]
			if !ok {
				return nil, fmt.Errorf("promotion dependency %q for %q was not adopted in this promotion set", dep, specDraft.Title)
			}
			if err := m.github.AddBlockedBy(info.ProjectDir, state.Repo, specIssue.Number, depIssue.Number); err != nil {
				return nil, err
			}
		}
	}
	if preparedProject != nil {
		if err := m.populateProjectWorkspaceItems(info.ProjectDir, state.Repo, preparedProject, initiativeIssue, specIssuesBySlug, draft.ProposedSpecIssues); err != nil {
			return nil, err
		}
		if projectRecord != nil {
			projectRecord.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			state.ProjectDecisions[projectRecord.Slug] = *projectRecord
			result.ProjectDecision = projectRecord
		}
		result.ProjectWorkspace = preparedProject.result
	}
	for slug, specIssue := range specIssuesBySlug {
		specDraft := specDraftsBySlug[slug]
		record := workspace.GitHubPlanningRecord{
			Slug:              specDraft.Slug,
			Kind:              "spec",
			Title:             specIssue.Title,
			IssueNumber:       specIssue.Number,
			IssueURL:          specIssue.URL,
			RemoteState:       specIssue.State,
			Readiness:         string(specDraft.Readiness),
			OwnershipMode:     string(mode),
			EntryMode:         string(draft.Source.EntryMode),
			SourceMode:        string(draft.Source.Mode),
			ParentIssueNumber: issueNumberOrZero(initiativeIssue),
			MilestoneNumber:   milestoneNumberOrZero(milestone),
			MilestoneTitle:    milestoneTitle(milestone),
			BlockedBy:         slugs(specDraft.BlockedBy),
			UpdatedAt:         time.Now().UTC().Format(time.RFC3339),
		}
		record.DiscussionNumber = discussionNumber(draft.Source.Discussion)
		record.DiscussionURL = discussionURL(draft.Source.Discussion)
		state.Planning[record.Slug] = record
		result.Adopted = append(result.Adopted, record)
	}
	state.LastUpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := m.workspace.WriteGitHubState(*state); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *Manager) adoptPromotionIssue(projectDir, repo string, issueNumber int, draft PromotionIssueDraft, milestone *GitHubMilestone) (*GitHubIssue, error) {
	issue, err := m.github.GetIssue(projectDir, repo, issueNumber)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(strings.TrimSpace(issue.Title), strings.TrimSpace(draft.Title)) {
		return nil, fmt.Errorf("issue #%d title %q does not match promotion draft title %q", issue.Number, issue.Title, draft.Title)
	}
	input := GitHubIssueInput{
		Title:  issue.Title,
		Body:   draft.Body,
		State:  issue.State,
		Labels: mergeLabels(issue.Labels, draft.Labels),
	}
	if milestone != nil {
		input.Milestone = &milestone.Number
	}
	return m.github.UpdateIssue(projectDir, repo, issue.Number, input)
}

func (m *Manager) loadCollaborationSourceData(brainstormSlug, discussionRef string) (*collaborationSourceData, CollaborationOwnership, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, CollaborationOwnership{}, err
	}
	meta, err := m.workspace.ReadWorkspaceMeta()
	if err != nil {
		return nil, CollaborationOwnership{}, err
	}
	mode := meta.SourceMode
	if mode == "" {
		mode = SourceOfTruthLocal
	}
	switch {
	case strings.TrimSpace(brainstormSlug) != "" && strings.TrimSpace(discussionRef) != "":
		return nil, CollaborationOwnership{}, fmt.Errorf("choose either --brainstorm or --discussion, not both")
	case strings.TrimSpace(brainstormSlug) != "":
		source, err := m.loadLocalBrainstormSource(info, brainstormSlug)
		if err != nil {
			return nil, CollaborationOwnership{}, err
		}
		return source, buildOwnership(mode, EntryModeLocalPromotion), nil
	case strings.TrimSpace(discussionRef) != "":
		source, err := m.loadGitHubDiscussionSource(info, discussionRef)
		if err != nil {
			return nil, CollaborationOwnership{}, err
		}
		return source, buildOwnership(mode, EntryModeGitHubCollaborative), nil
	default:
		return nil, CollaborationOwnership{}, fmt.Errorf("either --brainstorm or --discussion is required")
	}
}

func buildOwnership(mode SourceOfTruthMode, entry CollaborationEntryMode) CollaborationOwnership {
	if mode == "" {
		mode = SourceOfTruthLocal
	}
	ownership := CollaborationOwnership{
		Mode:            mode,
		EntryMode:       entry,
		MirrorLocalMeta: true,
	}
	switch entry {
	case EntryModeGitHubCollaborative:
		ownership.CanonicalDiscussion = "github_discussion"
	default:
		ownership.CanonicalDiscussion = "local_brainstorm"
	}
	switch mode {
	case SourceOfTruthGitHub:
		ownership.CanonicalPlanningArtifact = "github_issue"
		ownership.ReadinessSource = "github_issue"
		ownership.WriteTargets = []string{"github_issue", "github_milestone"}
	case SourceOfTruthHybrid:
		ownership.CanonicalPlanningArtifact = "mixed"
		ownership.ReadinessSource = "github_issue"
		ownership.WriteTargets = []string{"local_meta", "github_issue", "github_milestone"}
	default:
		ownership.CanonicalPlanningArtifact = "local_spec"
		ownership.ReadinessSource = "local_spec"
		ownership.WriteTargets = []string{"local_spec"}
	}
	return ownership
}

func (m *Manager) loadLocalBrainstormSource(info *workspace.Info, slug string) (*collaborationSourceData, error) {
	note, err := notes.Read(filepath.Join(info.BrainstormsDir, slugify(slug)+".md"))
	if err != nil {
		return nil, err
	}
	refinement := extractBrainstormRefinement(note)
	challenge := extractBrainstormChallenge(note)
	content := note.Content
	title := strings.TrimSpace(note.Title)
	vision := strings.TrimSpace(notes.ExtractSection(content, "Vision"))
	if title == "" {
		title = strings.TrimSpace(vision)
	}
	source := CollaborationSourceRef{
		Mode:            CollaborationSourceLocal,
		EntryMode:       EntryModeLocalPromotion,
		BrainstormSlug:  slugify(slug),
		BrainstormPath:  rel(info.ProjectDir, note.Path),
		SourceLinks:     []string{rel(info.ProjectDir, note.Path)},
		CanonicalSource: "local_brainstorm",
	}
	specParse := extractSpecCandidateParse(content)
	return &collaborationSourceData{
		source:                  source,
		title:                   defaultString(title, slugify(slug)),
		body:                    content,
		problem:                 firstNonEmpty(refinement.Problem, notes.ExtractSection(content, "Problem"), notes.ExtractSection(content, "Focus Question")),
		goals:                   firstNonEmpty(refinement.UserValue, notes.ExtractSection(content, "Desired Outcome"), vision),
		constraints:             firstNonEmpty(refinement.Constraints, notes.ExtractSection(content, "Constraints")),
		nonGoals:                firstNonEmpty(challenge.NoGos, extractSubsection(content, "Challenge", "No-Gos")),
		shape:                   firstNonEmpty(refinement.DecisionSnapshot, refinement.CandidateApproaches, challenge.SimplerAlternative),
		openQs:                  firstNonEmpty(refinement.RemainingOpenQuestions, notes.ExtractSection(content, "Open Questions")),
		suggested:               specParse.Titles,
		deps:                    buildDependencyGuess(specParse.Titles, content),
		explicitMultiSpecIntent: specParse.ExplicitMultiSpecIntent,
		repairCandidates:        specParse.RepairCandidates,
	}, nil
}

func (m *Manager) loadGitHubDiscussionSource(info *workspace.Info, discussionRef string) (*collaborationSourceData, error) {
	context, err := m.github.CurrentContext(info.ProjectDir)
	if err != nil {
		return nil, err
	}
	number, err := parseDiscussionRef(discussionRef)
	if err != nil {
		return nil, err
	}
	discussion, err := m.github.GetDiscussion(info.ProjectDir, context.Repo.Repo, number)
	if err != nil {
		return nil, err
	}
	content := strings.TrimSpace(discussion.Body)
	for _, comment := range discussion.Comments {
		if strings.TrimSpace(comment.Body) == "" {
			continue
		}
		content += "\n\n## Comment\n" + strings.TrimSpace(comment.Body)
	}
	source := CollaborationSourceRef{
		Mode:      CollaborationSourceGitHubDiscussion,
		EntryMode: EntryModeGitHubCollaborative,
		Discussion: &GitHubDiscussionRef{
			Number: discussion.Number,
			URL:    discussion.URL,
			Title:  discussion.Title,
		},
		SourceLinks:     []string{discussion.URL},
		CanonicalSource: "github_discussion",
	}
	specParse := extractSpecCandidateParse(content)
	return &collaborationSourceData{
		source:                  source,
		title:                   discussion.Title,
		body:                    content,
		problem:                 firstSectionOrKeyword(content, "Problem", "problem"),
		goals:                   firstSectionOrKeyword(content, "Goals", "goal"),
		constraints:             firstSectionOrKeyword(content, "Constraints", "constraint"),
		nonGoals:                firstSectionOrKeyword(content, "Non-Goals", "non-goal"),
		shape:                   firstNonEmpty(firstSectionOrKeyword(content, "Proposed Shape", "shape"), firstSectionOrKeyword(content, "Decision Snapshot", "decision"), firstSectionOrKeyword(content, "Candidate Approaches", "approach")),
		openQs:                  firstNonEmpty(firstSectionOrKeyword(content, "Open Questions", "open question"), firstSectionOrKeyword(content, "Remaining Open Questions", "remaining open question")),
		suggested:               specParse.Titles,
		deps:                    buildDependencyGuess(specParse.Titles, content),
		explicitMultiSpecIntent: specParse.ExplicitMultiSpecIntent,
		repairCandidates:        specParse.RepairCandidates,
	}, nil
}

func assessCollaborationData(data *collaborationSourceData) CollaborationMaturityDecision {
	strengths := []string{}
	gaps := []string{}
	if strings.TrimSpace(data.problem) != "" {
		strengths = append(strengths, "clear problem")
	} else {
		gaps = append(gaps, "missing concrete problem statement")
	}
	if strings.TrimSpace(data.goals) != "" {
		strengths = append(strengths, "clear goals or user value")
	} else {
		gaps = append(gaps, "missing goals or user value")
	}
	if strings.TrimSpace(data.constraints) != "" {
		strengths = append(strengths, "clear constraints")
	} else {
		gaps = append(gaps, "missing constraints")
	}
	if strings.TrimSpace(data.nonGoals) != "" {
		strengths = append(strengths, "non-goals are called out")
	} else {
		gaps = append(gaps, "missing non-goals")
	}
	if strings.TrimSpace(data.shape) != "" {
		strengths = append(strengths, "solution shape is visible")
	} else {
		gaps = append(gaps, "missing initial solution shape")
	}

	decision := CollaborationMaturityDecision{
		State:      MaturityNotReady,
		Confidence: MaturityConfidenceLow,
		SourceMode: data.source.Mode,
		Reason:     "The source still has planning gaps that would force promotion to guess.",
		Strengths:  strengths,
		Gaps:       gaps,
		SuggestedTitles: MaturitySuggestedTitles{
			Initiative: defaultString(data.title, "Untitled initiative"),
		},
	}
	if len(gaps) > 0 {
		return decision
	}
	specTitles := append([]string(nil), dedupeTitles(data.suggested)...)
	if data.explicitMultiSpecIntent && len(specTitles) < 2 {
		candidates := dedupeTitles(append(append([]string(nil), data.repairCandidates...), specTitles...))
		decision.State = MaturityNeedsSourceRepair
		decision.Confidence = MaturityConfidenceMedium
		decision.Reason = "The source asks for multiple spec issues, but Plan could not parse at least two spec titles."
		decision.BlockingReason = "requested multi-spec promotion but source parsed as single spec"
		decision.Gaps = nil
		decision.SuggestedTitles = MaturitySuggestedTitles{
			Initiative: defaultString(data.title, "Untitled initiative"),
			Specs:      specTitles,
		}
		decision.RepairCandidates = candidates
		decision.NextCommand = repairSpecSplitCommand(data.source, candidates)
		return decision
	}
	if len(specTitles) == 0 {
		specTitles = []string{fallbackSpecTitle(data.title)}
	}
	decision.Confidence = MaturityConfidenceHigh
	decision.Gaps = nil
	decision.DependencyGuess = normalizeDependencyGuess(specTitles, data.deps)
	if len(specTitles) > 1 {
		decision.State = MaturityReadyMultiSpec
		decision.RecommendedPath = PromotionMultiSpec
		decision.Reason = fmt.Sprintf("The source is clear enough to promote as an initiative with %d initial specs.", len(specTitles))
		decision.SuggestedTitles = MaturitySuggestedTitles{
			Initiative: defaultString(data.title, "Untitled initiative"),
			Specs:      specTitles,
		}
		return decision
	}
	decision.State = MaturityReadySingleSpec
	decision.RecommendedPath = PromotionSingleSpec
	decision.Reason = "The source is clear enough to promote directly into one bounded spec."
	decision.SuggestedTitles = MaturitySuggestedTitles{Specs: specTitles}
	return decision
}

func buildPromotionInitiativeDraft(data *collaborationSourceData, title string, specs []string) PromotionIssueDraft {
	lines := []string{
		"## Initiative",
		data.problem,
		"",
		"## Outcome",
		"- " + defaultString(data.goals, "Deliver the agreed collaboration outcome."),
		"",
		"## Why",
		"- " + defaultString(data.problem, "Capture the accepted collaboration direction in an executable GitHub planning shape."),
		"",
		"## Scope Boundary",
		"- Includes the promoted spec set listed below.",
		"- Does not include implementation beyond those specs.",
		"",
		"## Specs",
	}
	for _, spec := range specs {
		lines = append(lines, "- [ ] "+spec)
	}
	lines = append(lines,
		"",
		"## Dependency Notes",
		"- Follow the promoted spec dependency plan.",
		"",
		"## Milestone",
		"- "+title,
		"",
		"## Source",
	)
	for _, link := range data.source.SourceLinks {
		lines = append(lines, "- "+link)
	}
	return PromotionIssueDraft{
		Kind:           "initiative",
		Title:          title,
		Body:           strings.TrimSpace(strings.Join(lines, "\n")),
		Slug:           slugify(title),
		Readiness:      ReadinessReady,
		Labels:         []string{"enhancement", planIssueInitiativeLabel},
		SourceLinks:    append([]string(nil), data.source.SourceLinks...),
		ReadyByDefault: true,
	}
}

func buildPromotionSpecDraft(data *collaborationSourceData, title string, blockedBy []string, exception PromotionRefinementException) PromotionIssueDraft {
	var exceptionPtr *PromotionRefinementException
	if strings.TrimSpace(exception.IssueTitle) != "" {
		exceptionPtr = &exception
	}
	body := renderPromotionSpecBody(title, data, blockedBy, exceptionPtr)
	readiness := ReadinessReady
	labels := []string{"enhancement", planIssueSpecLabel, planIssueReadyLabel}
	if exceptionPtr != nil {
		readiness = ReadinessNeedsRefinement
		labels = []string{"enhancement", planIssueSpecLabel}
	} else if len(blockedBy) > 0 {
		readiness = ReadinessBlocked
		labels = []string{"enhancement", planIssueSpecLabel, planIssueBlockedLabel}
	}
	return PromotionIssueDraft{
		Kind:           "spec",
		Title:          title,
		Body:           body,
		Slug:           slugify(title),
		Readiness:      readiness,
		Labels:         labels,
		SourceLinks:    append([]string(nil), data.source.SourceLinks...),
		BlockedBy:      append([]string(nil), blockedBy...),
		ReadyByDefault: len(blockedBy) == 0 && exceptionPtr == nil,
	}
}

func buildRefinementExceptions(data *collaborationSourceData, specTitles []string) map[string]PromotionRefinementException {
	out := map[string]PromotionRefinementException{}
	raw := strings.TrimSpace(data.openQs)
	if raw == "" {
		return out
	}
	lower := strings.ToLower(raw)
	if !strings.Contains(lower, "needs-refinement") && !strings.Contains(lower, "refinement gap") {
		return out
	}
	lines := bulletItems(raw)
	if len(specTitles) == 1 {
		out[specTitles[0]] = PromotionRefinementException{
			IssueTitle:               specTitles[0],
			Gap:                      strings.TrimSpace(raw),
			WhyNotReady:              "The source still carries an explicit refinement gap that would force execution to guess.",
			RecommendedClarification: "Resolve the explicit refinement gap before treating this spec as ready.",
			ExitCriteria:             []string{"The remaining refinement gap is resolved or explicitly deferred."},
		}
		return out
	}
	for _, title := range specTitles {
		for _, line := range lines {
			if !strings.Contains(strings.ToLower(line), strings.ToLower(title)) {
				continue
			}
			out[title] = PromotionRefinementException{
				IssueTitle:               title,
				Gap:                      strings.TrimSpace(line),
				WhyNotReady:              "The source calls out a spec-specific refinement gap that still needs resolution.",
				RecommendedClarification: "Resolve the explicit refinement note for this spec before execution starts.",
				ExitCriteria:             []string{"The spec-specific refinement gap is resolved or removed from the source."},
			}
			break
		}
	}
	return out
}

func renderPromotionSpecBody(title string, data *collaborationSourceData, blockedBy []string, exception *PromotionRefinementException) string {
	lines := []string{
		"## Spec",
		defaultString(data.problem, title),
		"",
		"## Problem",
		defaultString(data.problem, "-"),
		"",
		"## Goals",
		defaultString(data.goals, "-"),
		"",
		"## Non-Goals",
		defaultString(data.nonGoals, "-"),
		"",
		"## Constraints",
		defaultString(data.constraints, "-"),
		"",
		"## Proposed Shape",
		defaultString(data.shape, "-"),
		"",
		"## Verification",
		"- Add automated coverage for the promoted behavior.",
		"- Verify the CLI output and GitHub integration contract for this slice.",
		"",
		"## Dependencies",
	}
	if len(blockedBy) == 0 {
		lines = append(lines, "- blocked by: none")
	} else {
		lines = append(lines, "- blocked by: "+strings.Join(blockedBy, ", "))
	}
	lines = append(lines,
		"",
		"## Readiness",
	)
	if exception == nil {
		if len(blockedBy) == 0 {
			lines = append(lines, "- status: ready")
		} else {
			lines = append(lines, "- status: blocked")
		}
	} else {
		lines = append(lines, "- status: needs-refinement")
		lines = append(lines,
			"",
			"## Refinement Gap",
			exception.Gap,
			"",
			"## Why Not Ready",
			exception.WhyNotReady,
			"",
			"## Recommended Clarification",
			exception.RecommendedClarification,
			"",
			"## Exit Criteria",
		)
		for _, item := range exception.ExitCriteria {
			lines = append(lines, "- "+item)
		}
	}
	lines = append(lines,
		"",
		"## Source",
	)
	for _, link := range data.source.SourceLinks {
		lines = append(lines, "- "+link)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

type specCandidateParse struct {
	Titles                  []string
	RepairCandidates        []string
	ExplicitMultiSpecIntent bool
}

func extractSpecCandidates(content string) []string {
	return extractSpecCandidateParse(content).Titles
}

func extractSpecCandidateParse(content string) specCandidateParse {
	lower := strings.ToLower(content)
	explicitIssueIntent := containsAny(lower, []string{
		"create spec issues",
		"spec issues for:",
		"create specs for:",
		"create spec issue",
	})
	explicitMultiIntent := explicitIssueIntent ||
		strings.Contains(lower, "multiple specs") ||
		strings.Contains(lower, "multi-spec") ||
		strings.Contains(lower, "split into")

	for _, heading := range []string{"Spec Split", "Initial Spec Split", "Planned Specs", "Specs"} {
		if section := strings.TrimSpace(notes.ExtractSection(content, heading)); section != "" {
			items := bulletItems(section)
			if len(items) > 0 {
				return specCandidateParse{
					Titles:                  items,
					RepairCandidates:        items,
					ExplicitMultiSpecIntent: explicitMultiIntent || len(items) > 1,
				}
			}
		}
	}

	phraseItems := specIssuePhraseItems(content)
	if len(phraseItems) > 0 {
		return specCandidateParse{
			Titles:                  phraseItems,
			RepairCandidates:        phraseItems,
			ExplicitMultiSpecIntent: true,
		}
	}

	if explicitIssueIntent {
		if items := numberedItems(content); len(items) > 0 {
			return specCandidateParse{
				Titles:                  items,
				RepairCandidates:        items,
				ExplicitMultiSpecIntent: true,
			}
		}
	}

	if explicitMultiIntent {
		if section := strings.TrimSpace(notes.ExtractSection(content, "Desired Outcome")); section != "" {
			if items := bulletItems(section); len(items) > 0 {
				return specCandidateParse{
					Titles:                  items,
					RepairCandidates:        items,
					ExplicitMultiSpecIntent: true,
				}
			}
		}
		if section := strings.TrimSpace(extractSubsection(content, "Refinement", "Candidate Approaches")); section != "" {
			if items := bulletItems(section); len(items) > 1 {
				return specCandidateParse{Titles: items, RepairCandidates: items, ExplicitMultiSpecIntent: true}
			}
		}
		if section := strings.TrimSpace(notes.ExtractSection(content, "Ideas")); section != "" {
			if items := bulletItems(section); len(items) > 1 {
				return specCandidateParse{Titles: items, RepairCandidates: items, ExplicitMultiSpecIntent: true}
			}
		}
		if items := bulletItems(content); len(items) > 1 {
			return specCandidateParse{Titles: items, RepairCandidates: items, ExplicitMultiSpecIntent: true}
		}
	}

	return specCandidateParse{
		ExplicitMultiSpecIntent: explicitMultiIntent,
		RepairCandidates:        phraseItems,
	}
}

func buildDependencyGuess(specs []string, content string) []MaturityDependencyGuess {
	if len(specs) == 0 {
		return nil
	}
	guesses := make([]MaturityDependencyGuess, 0, len(specs))
	lower := strings.ToLower(content)
	for _, spec := range specs {
		guess := MaturityDependencyGuess{Spec: spec}
		for _, candidate := range specs {
			if candidate == spec {
				continue
			}
			pattern := fmt.Sprintf("%s depends on %s", strings.ToLower(spec), strings.ToLower(candidate))
			if strings.Contains(lower, pattern) {
				guess.BlockedBy = append(guess.BlockedBy, candidate)
			}
		}
		guesses = append(guesses, guess)
	}
	return guesses
}

func normalizeDependencyGuess(specTitles []string, guesses []MaturityDependencyGuess) []MaturityDependencyGuess {
	byTitle := map[string][]string{}
	for _, guess := range guesses {
		byTitle[guess.Spec] = append([]string(nil), guess.BlockedBy...)
	}
	out := make([]MaturityDependencyGuess, 0, len(specTitles))
	for i, title := range specTitles {
		blockedBy := dedupeTitles(byTitle[title])
		if len(blockedBy) == 0 && i > 0 && byTitle[title] == nil && len(specTitles) > 1 {
			// Keep the default sparse when no explicit dependency hints exist.
			blockedBy = nil
		}
		out = append(out, MaturityDependencyGuess{Spec: title, BlockedBy: blockedBy})
	}
	return out
}

func firstSectionOrKeyword(content, heading, keyword string) string {
	if section := strings.TrimSpace(notes.ExtractSection(content, heading)); section != "" {
		return section
	}
	pattern := regexp.MustCompile(`(?im)^\s*[-*]?\s*` + regexp.QuoteMeta(keyword) + `s?\s*:\s*(.+)$`)
	if match := pattern.FindStringSubmatch(content); len(match) == 2 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func fallbackSpecTitle(title string) string {
	return strings.TrimSpace(title)
}

func parseDiscussionRef(value string) (int, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return 0, fmt.Errorf("discussion reference is required")
	}
	if number, err := strconv.Atoi(raw); err == nil && number > 0 {
		return number, nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return 0, fmt.Errorf("parse discussion reference: %w", err)
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "discussions" {
			number, err := strconv.Atoi(parts[i+1])
			if err != nil {
				break
			}
			return number, nil
		}
	}
	return 0, fmt.Errorf("could not resolve discussion number from %q", value)
}

func specIssuePhraseItems(content string) []string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	var out []string
	pattern := regexp.MustCompile(`(?i)\b(?:create\s+)?spec\s+issues?\s+for\s*:\s*(.*)$`)
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		match := pattern.FindStringSubmatch(line)
		if len(match) != 2 {
			continue
		}
		if items := splitInlineSpecList(match[1]); len(items) > 0 {
			out = append(out, items...)
			continue
		}
		var block []string
		for j := i + 1; j < len(lines); j++ {
			next := strings.TrimSpace(lines[j])
			if next == "" {
				break
			}
			if strings.HasPrefix(next, "#") {
				break
			}
			block = append(block, next)
		}
		if items := bulletItems(strings.Join(block, "\n")); len(items) > 0 {
			out = append(out, items...)
			continue
		}
		if items := numberedItems(strings.Join(block, "\n")); len(items) > 0 {
			out = append(out, items...)
		}
	}
	return dedupeTitles(out)
}

func splitInlineSpecList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := regexp.MustCompile(`[;,]`).Split(raw, -1)
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := cleanCandidateTitle(part)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil
	}
	return dedupeTitles(out)
}

func bulletItems(section string) []string {
	lines := strings.Split(section, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ") && !strings.HasPrefix(trimmed, "* ") {
			continue
		}
		trimmed = strings.TrimPrefix(trimmed, "- ")
		trimmed = strings.TrimPrefix(trimmed, "* ")
		switch {
		case strings.HasPrefix(trimmed, "[ ] "):
			trimmed = strings.TrimPrefix(trimmed, "[ ] ")
		case strings.HasPrefix(strings.ToLower(trimmed), "[x] "):
			trimmed = trimmed[4:]
		}
		trimmed = cleanCandidateTitle(trimmed)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return dedupeTitles(out)
}

func numberedItems(section string) []string {
	lines := strings.Split(section, "\n")
	out := make([]string, 0, len(lines))
	pattern := regexp.MustCompile(`^\s*\d+[\.)]\s+(.+)$`)
	for _, line := range lines {
		match := pattern.FindStringSubmatch(line)
		if len(match) != 2 {
			continue
		}
		item := cleanCandidateTitle(match[1])
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return dedupeTitles(out)
}

func cleanCandidateTitle(item string) string {
	item = strings.TrimSpace(item)
	item = strings.TrimPrefix(item, "[ ] ")
	if strings.HasPrefix(strings.ToLower(item), "[x] ") {
		item = item[4:]
	}
	item = strings.Trim(item, " \t-*:;,.")
	return strings.TrimSpace(item)
}

func dedupeTitles(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func mergeLabels(existing, desired []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(existing)+len(desired))
	for _, label := range append(append([]string(nil), existing...), desired...) {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		key := strings.ToLower(label)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, label)
	}
	return out
}

func repairSpecSplitCommand(source CollaborationSourceRef, candidates []string) string {
	args := []string{"plan", "discuss", "repair", "--project", "."}
	switch source.Mode {
	case CollaborationSourceGitHubDiscussion:
		args = append(args, "--discussion", strconv.Itoa(discussionNumber(source.Discussion)))
	default:
		args = append(args, "--brainstorm", source.BrainstormSlug)
	}
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		args = append(args, "--spec", candidate)
	}
	for i := len(candidates); i < 2; i++ {
		placeholder := "<spec title>"
		if i > 0 {
			placeholder = "<another spec title>"
		}
		args = append(args, "--spec", placeholder)
	}
	if source.Mode == CollaborationSourceGitHubDiscussion {
		args = append(args, "--confirm")
	}
	args = append(args, "--format", "json")
	return shellCommand(args)
}

func refreshAssessCommand(source CollaborationSourceRef) string {
	args := []string{"plan", "discuss", "assess", "--project", "."}
	args = append(args, sourceSelectorArgs(source)...)
	args = append(args, "--format", "json")
	return shellCommand(args)
}

func refreshPromoteCommand(source CollaborationSourceRef) string {
	args := []string{"plan", "discuss", "promote", "--project", "."}
	args = append(args, sourceSelectorArgs(source)...)
	args = append(args, "--format", "json")
	return shellCommand(args)
}

func sourceSelectorArgs(source CollaborationSourceRef) []string {
	switch source.Mode {
	case CollaborationSourceGitHubDiscussion:
		return []string{"--discussion", strconv.Itoa(discussionNumber(source.Discussion))}
	default:
		return []string{"--brainstorm", source.BrainstormSlug}
	}
}

func renderSpecSplitSection(specs []string) string {
	lines := make([]string, 0, len(specs))
	for _, spec := range specs {
		if strings.TrimSpace(spec) == "" {
			continue
		}
		lines = append(lines, "- "+strings.TrimSpace(spec))
	}
	return strings.Join(lines, "\n")
}

func shellCommand(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, shellQuote(arg))
	}
	return strings.Join(quoted, " ")
}

func shellQuote(arg string) string {
	if arg == "" {
		return "''"
	}
	if safeShellArgPattern.MatchString(arg) {
		return arg
	}
	return "'" + strings.ReplaceAll(arg, "'", "'\"'\"'") + "'"
}

func findBlockedByForTitle(guesses []MaturityDependencyGuess, title string) []string {
	for _, guess := range guesses {
		if strings.EqualFold(strings.TrimSpace(guess.Spec), strings.TrimSpace(title)) {
			return append([]string(nil), guess.BlockedBy...)
		}
	}
	return nil
}

func slugs(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, slugify(item))
	}
	return out
}

func shouldRecommendProject(specCount int, deps []PromotionDependencyPlan) bool {
	if specCount >= 5 {
		return true
	}
	for _, dep := range deps {
		if len(dep.BlockedBy) > 0 {
			return true
		}
	}
	return false
}

func validateProjectDecision(decision string, draft *PromotionDraft, ref GitHubProjectReference) error {
	decision = normalizeProjectDecision(decision)
	switch decision {
	case "", projectDecisionCreate, projectDecisionSkip, projectDecisionConnect:
	default:
		return fmt.Errorf("unsupported project decision %q; use create, skip, or connect", decision)
	}
	if requiresProjectDecision(draft) && decision == "" {
		return fmt.Errorf("promotion has %d specs and requires --project-decision create|skip|connect before apply", len(draft.ProposedSpecIssues))
	}
	if decision == projectDecisionConnect {
		if strings.TrimSpace(ref.ID) == "" && (strings.TrimSpace(ref.Owner) == "" || ref.Number <= 0) {
			return fmt.Errorf("project decision %q requires --project-owner and --project-number, --project-url, or --project-id", decision)
		}
	}
	if decision == projectDecisionCreate {
		if ref.Number != 0 || strings.TrimSpace(ref.ID) != "" || strings.TrimSpace(ref.URL) != "" {
			return fmt.Errorf("project decision %q creates a new Project; omit existing project number, id, and url", decision)
		}
	}
	if decision == projectDecisionSkip {
		if strings.TrimSpace(ref.Owner) != "" || ref.Number != 0 || strings.TrimSpace(ref.ID) != "" || strings.TrimSpace(ref.URL) != "" {
			return fmt.Errorf("project decision %q skips Project provisioning; omit project reference flags", decision)
		}
	}
	return nil
}

func normalizeProjectDecision(decision string) string {
	return strings.ToLower(strings.TrimSpace(decision))
}

func requiresProjectDecision(draft *PromotionDraft) bool {
	return draft != nil && len(draft.ProposedSpecIssues) >= 5
}

func buildProjectDecisionRecord(draft *PromotionDraft, decision string, milestone *GitHubMilestone, ref GitHubProjectReference) workspace.GitHubProjectDecisionRecord {
	now := time.Now().UTC().Format(time.RFC3339)
	slug := promotionProjectDecisionSlug(draft)
	record := workspace.GitHubProjectDecisionRecord{
		Slug:             slug,
		Decision:         normalizeProjectDecision(decision),
		InitiativeSlug:   slug,
		SpecCount:        len(draft.ProposedSpecIssues),
		MilestoneNumber:  milestoneNumberOrZero(milestone),
		MilestoneTitle:   milestoneTitle(milestone),
		ProjectOwner:     ref.Owner,
		ProjectNumber:    ref.Number,
		ProjectID:        ref.ID,
		ProjectURL:       ref.URL,
		SourceMode:       string(draft.Source.Mode),
		EntryMode:        string(draft.Source.EntryMode),
		DiscussionNumber: discussionNumber(draft.Source.Discussion),
		DiscussionURL:    discussionURL(draft.Source.Discussion),
		UpdatedAt:        now,
	}
	if draft.ProjectPrompt != nil {
		record.Reason = draft.ProjectPrompt.Reason
	}
	return record
}

func promotionProjectDecisionSlug(draft *PromotionDraft) string {
	if draft != nil && draft.ProposedInitiativeIssue != nil {
		return draft.ProposedInitiativeIssue.Slug
	}
	if draft != nil && len(draft.ProposedSpecIssues) > 0 {
		return draft.ProposedSpecIssues[0].Slug
	}
	return "promotion"
}

func (m *Manager) ensurePromotionLabels(projectDir, repo string, draft *PromotionDraft) error {
	for _, label := range promotionLabelInputs(draft) {
		if err := m.github.EnsureLabel(projectDir, repo, label); err != nil {
			return err
		}
	}
	return nil
}

func promotionLabelInputs(draft *PromotionDraft) []GitHubLabelInput {
	colors := map[string]string{
		planIssueInitiativeLabel: "5319e7",
		planIssueSpecLabel:       "1d76db",
		planIssueReadyLabel:      "0e8a16",
		planIssueBlockedLabel:    "d93f0b",
	}
	descriptions := map[string]string{
		planIssueInitiativeLabel: "Plan-managed initiative issue",
		planIssueSpecLabel:       "Plan-managed spec issue",
		planIssueReadyLabel:      "Plan-managed issue is ready for execution",
		planIssueBlockedLabel:    "Plan-managed issue is blocked by another planning issue",
	}
	names := map[string]struct{}{}
	if draft != nil && draft.ProposedInitiativeIssue != nil {
		for _, label := range draft.ProposedInitiativeIssue.Labels {
			names[label] = struct{}{}
		}
	}
	if draft != nil {
		for _, spec := range draft.ProposedSpecIssues {
			for _, label := range spec.Labels {
				names[label] = struct{}{}
			}
		}
	}
	out := make([]GitHubLabelInput, 0, len(names))
	for name := range names {
		if strings.TrimSpace(name) == "" {
			continue
		}
		if !strings.HasPrefix(name, "plan:") {
			continue
		}
		color := colors[name]
		if color == "" {
			color = "ededed"
		}
		out = append(out, GitHubLabelInput{
			Name:        name,
			Color:       color,
			Description: descriptions[name],
		})
	}
	return out
}

func promotionManualFallbackError(draft *PromotionDraft, result *PromotionApplyResult, err error) error {
	if draft != nil {
		draft.ManualFallbackAllowed = true
	}
	if result == nil {
		result = &PromotionApplyResult{Draft: draft}
	}
	result.ManualFallbackAllowed = true
	if err != nil {
		result.ManualFallbackReason = err.Error()
	}
	result.NextCommand = manualFallbackAdoptCommand(draft)
	return &PromotionApplyManualFallbackError{Result: result, Err: err}
}

func manualFallbackAdoptCommand(draft *PromotionDraft) string {
	args := []string{"plan", "github", "adopt", "--project", "."}
	if draft != nil {
		args = append(args, sourceSelectorArgs(draft.Source)...)
	}
	placeholders := []string{}
	if draft != nil && draft.ProposedInitiativeIssue != nil {
		placeholders = append(placeholders, "<initiative issue>")
	}
	if draft != nil {
		for range draft.ProposedSpecIssues {
			placeholders = append(placeholders, "<spec issue>")
		}
	}
	if len(placeholders) == 0 {
		placeholders = append(placeholders, "<issue>")
	}
	args = append(args, "--issues", strings.Join(placeholders, ","))
	if requiresProjectDecision(draft) {
		args = append(args, "--project-decision", "<create|skip|connect>")
	}
	args = append(args, "--format", "json")
	return shellCommand(args)
}

func milestoneNumberOrZero(m *GitHubMilestone) int {
	if m == nil {
		return 0
	}
	return m.Number
}

func milestoneTitle(m *GitHubMilestone) string {
	if m == nil {
		return ""
	}
	return m.Title
}

func issueNumberOrZero(issue *GitHubIssue) int {
	if issue == nil {
		return 0
	}
	return issue.Number
}

func discussionNumber(ref *GitHubDiscussionRef) int {
	if ref == nil {
		return 0
	}
	return ref.Number
}

func discussionURL(ref *GitHubDiscussionRef) string {
	if ref == nil {
		return ""
	}
	return ref.URL
}

func (m *Manager) ensureMilestone(projectDir, repo, title string) (*GitHubMilestone, error) {
	existing, err := m.github.FindMilestone(projectDir, repo, title)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	return m.github.CreateMilestone(projectDir, repo, GitHubMilestoneInput{Title: title})
}
