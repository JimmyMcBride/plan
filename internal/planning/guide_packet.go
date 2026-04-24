package planning

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"plan/internal/notes"
	"plan/internal/workspace"
)

const (
	GuidePacketSchemaVersion = 2
	guidePacketKind          = "guide_packet"
	guidePlanningModeGuided  = "guided"
	guidePlanningModeCollab  = "collaboration"
)

var brainstormGuideCheckpoints = map[string]struct{}{
	"vision-intake":                {},
	"clarify-problem-user-value":   {},
	"clarify-constraints-appetite": {},
	"clarify-open-approaches":      {},
	"handoff-epic":                 {},
}

var collaborationGuideStages = map[string]struct{}{
	"discussion_assess": {},
	"promotion_review":  {},
	"initiative_draft":  {},
	"spec_draft":        {},
	"needs_refinement":  {},
}

type GuidePacket struct {
	SchemaVersion  int                        `json:"schema_version"`
	Kind           string                     `json:"kind"`
	GeneratedAt    string                     `json:"generated_at"`
	Builder        GuidePacketBuilderMeta     `json:"builder"`
	Workspace      GuidePacketWorkspace       `json:"workspace"`
	Ownership      CollaborationOwnership     `json:"ownership"`
	Session        GuidePacketSession         `json:"session"`
	Artifact       GuidePacketArtifact        `json:"artifact"`
	Mode           GuidePacketMode            `json:"mode"`
	Sources        []string                   `json:"sources"`
	Collaboration  *GuidePacketCollaboration  `json:"collaboration,omitempty"`
	Contract       GuidePacketContract        `json:"contract"`
	RenderedDrafts []GuidePacketRenderedDraft `json:"rendered_drafts,omitempty"`
	Actions        []GuidePacketAction        `json:"actions,omitempty"`
	RenderedPrompt string                     `json:"rendered_prompt"`
}

type GuidePacketCollaboration struct {
	Source         CollaborationSourceRef   `json:"source"`
	Assessment     *CollaborationAssessment `json:"assessment,omitempty"`
	PromotionDraft *PromotionDraft          `json:"promotion_draft,omitempty"`
}

type GuidePacketBuilderMeta struct {
	Command string `json:"command"`
	Format  string `json:"format"`
}

type GuidePacketWorkspace struct {
	ProjectRoot       string `json:"project_root"`
	PlanningMode      string `json:"planning_mode"`
	PlanningModel     string `json:"planning_model"`
	SourceMode        string `json:"source_mode"`
	StoryBackend      string `json:"story_backend"`
	IntegrationBranch string `json:"integration_branch,omitempty"`
}

type GuidePacketSession struct {
	ChainID             string            `json:"chain_id"`
	CurrentStage        string            `json:"current_stage"`
	CurrentCluster      int               `json:"current_cluster,omitempty"`
	CurrentClusterLabel string            `json:"current_cluster_label,omitempty"`
	StageStatuses       map[string]string `json:"stage_statuses,omitempty"`
	Summary             string            `json:"summary,omitempty"`
	NextAction          string            `json:"next_action,omitempty"`
}

type GuidePacketArtifact struct {
	Type   string `json:"type"`
	Slug   string `json:"slug"`
	Title  string `json:"title"`
	Path   string `json:"path"`
	Status string `json:"status,omitempty"`
}

type GuidePacketMode struct {
	Stage      string `json:"stage"`
	Checkpoint string `json:"checkpoint"`
	Pass       string `json:"pass"`
}

type GuidePacketRenderedDraft struct {
	Kind      string `json:"kind"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Slug      string `json:"slug,omitempty"`
	Readiness string `json:"readiness,omitempty"`
}

type GuidePacketContract struct {
	Role             string                      `json:"role"`
	Stance           []string                    `json:"stance"`
	Goal             string                      `json:"goal"`
	QuestionStrategy GuidePacketQuestionStrategy `json:"question_strategy"`
	ArtifactStrategy GuidePacketArtifactStrategy `json:"artifact_strategy"`
	Do               []string                    `json:"do"`
	Avoid            []string                    `json:"avoid"`
	QualityBar       []string                    `json:"quality_bar"`
	CompletionGate   []string                    `json:"completion_gate"`
	CommandHints     []GuidePacketCommandHint    `json:"command_hints"`
}

type GuidePacketQuestionStrategy struct {
	ClusterSizeMin        int      `json:"cluster_size_min"`
	ClusterSizeMax        int      `json:"cluster_size_max"`
	ReflectOncePerCluster bool     `json:"reflect_once_per_cluster"`
	GapGuidance           string   `json:"gap_guidance,omitempty"`
	MenuActions           []string `json:"menu_actions"`
}

type GuidePacketArtifactStrategy struct {
	WriteMode          string   `json:"write_mode"`
	DurableArtifact    string   `json:"durable_artifact"`
	StrengthenSections []string `json:"strengthen_sections"`
	PreserveRules      []string `json:"preserve_rules"`
}

type GuidePacketCommandHint struct {
	Purpose string `json:"purpose"`
	Command string `json:"command"`
}

type GuidePacketAction struct {
	ID                   string `json:"id"`
	Kind                 string `json:"kind"`
	Label                string `json:"label"`
	Description          string `json:"description,omitempty"`
	Command              string `json:"command,omitempty"`
	Target               string `json:"target,omitempty"`
	RequiresConfirmation bool   `json:"requires_confirmation,omitempty"`
	Available            bool   `json:"available"`
	BlockedReason        string `json:"blocked_reason,omitempty"`
}

func (m *Manager) CurrentGuidePacket() (*GuidePacket, error) {
	session, err := m.ReadLastActiveGuidedSession()
	if err != nil {
		if errors.Is(err, ErrNoActiveGuidedSession) {
			return nil, fmt.Errorf("no active guided session. Start one with `plan brainstorm start --project . \"<topic>\"`: %w", ErrNoActiveGuidedSession)
		}
		return nil, err
	}
	return m.buildBrainstormGuidePacket("plan guide current", session, "", "")
}

func (m *Manager) GuidePacketForChain(chainID, stage, checkpoint string) (*GuidePacket, error) {
	session, err := m.ReadGuidedSession(strings.TrimSpace(chainID))
	if err != nil {
		return nil, err
	}
	return m.buildBrainstormGuidePacket("plan guide show", session, stage, checkpoint)
}

func (m *Manager) GuidePacketForCollaborationSource(brainstormSlug, discussionRef, stage string) (*GuidePacket, error) {
	return m.buildCollaborationGuidePacket("plan guide show", brainstormSlug, discussionRef, stage)
}

func (m *Manager) buildBrainstormGuidePacket(command string, session *workspace.GuidedSessionRecord, stageOverride, checkpointOverride string) (*GuidePacket, error) {
	if session == nil {
		return nil, fmt.Errorf("guided session is required")
	}
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	meta, err := m.workspace.ReadWorkspaceMeta()
	if err != nil {
		return nil, err
	}
	branch := ""
	if githubState, err := m.workspace.ReadGitHubState(); err == nil {
		branch = strings.TrimSpace(githubState.DefaultBranch)
	}
	if strings.TrimSpace(session.Brainstorm) == "" {
		return nil, fmt.Errorf("guided session %q is not linked to a brainstorm artifact", session.ChainID)
	}

	effectiveStage := strings.TrimSpace(stageOverride)
	if effectiveStage == "" {
		effectiveStage = strings.TrimSpace(session.CurrentStage)
	}
	if effectiveStage == "" {
		effectiveStage = "brainstorm"
	}
	if effectiveStage != "brainstorm" {
		return nil, fmt.Errorf("guided session chain packets only support the brainstorm stage; use `--stage brainstorm` or a collaboration source")
	}

	effectiveCheckpoint := strings.TrimSpace(checkpointOverride)
	if effectiveCheckpoint == "" {
		effectiveCheckpoint = strings.TrimSpace(session.CurrentClusterLabel)
	}
	if effectiveCheckpoint == "" {
		effectiveCheckpoint = "vision-intake"
	}
	if _, ok := brainstormGuideCheckpoints[effectiveCheckpoint]; !ok {
		return nil, fmt.Errorf("unsupported brainstorm checkpoint %q", effectiveCheckpoint)
	}

	brainstormPath := filepath.Join(info.BrainstormsDir, slugify(session.Brainstorm)+".md")
	brainstorm, err := notes.Read(brainstormPath)
	if err != nil {
		return nil, fmt.Errorf("read brainstorm artifact for guided session %q: %w", session.ChainID, err)
	}
	if brainstorm.Type != "brainstorm" {
		return nil, fmt.Errorf("%s is not a brainstorm artifact", rel(info.ProjectDir, brainstorm.Path))
	}
	sourceMode, err := m.SourceMode()
	if err != nil {
		return nil, err
	}

	artifactPath := rel(info.ProjectDir, brainstorm.Path)
	contract := brainstormGuideContract(session, effectiveCheckpoint, artifactPath)
	packet := &GuidePacket{
		SchemaVersion: GuidePacketSchemaVersion,
		Kind:          guidePacketKind,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Builder: GuidePacketBuilderMeta{
			Command: command,
			Format:  "json",
		},
		Workspace: GuidePacketWorkspace{
			ProjectRoot:       info.ProjectDir,
			PlanningMode:      guidePlanningModeGuided,
			PlanningModel:     meta.PlanningModel,
			SourceMode:        string(sourceMode),
			StoryBackend:      string(meta.StoryBackend),
			IntegrationBranch: branch,
		},
		Ownership: buildOwnership(sourceMode, EntryModeLocalPromotion),
		Session: GuidePacketSession{
			ChainID:             session.ChainID,
			CurrentStage:        effectiveStage,
			CurrentCluster:      brainstormGuideCheckpointIndex(effectiveCheckpoint, session.CurrentCluster),
			CurrentClusterLabel: effectiveCheckpoint,
			StageStatuses:       copyStringMap(session.StageStatuses),
			Summary:             strings.TrimSpace(session.Summary),
			NextAction:          strings.TrimSpace(session.NextAction),
		},
		Artifact: GuidePacketArtifact{
			Type:   brainstorm.Type,
			Slug:   session.Brainstorm,
			Title:  brainstorm.Title,
			Path:   artifactPath,
			Status: defaultString(stringValue(brainstorm.Metadata["status"]), "active"),
		},
		Mode: GuidePacketMode{
			Stage:      effectiveStage,
			Checkpoint: effectiveCheckpoint,
			Pass:       brainstormGuidePass(effectiveCheckpoint),
		},
		Sources: []string{
			rel(info.ProjectDir, info.ProjectFile),
			rel(info.ProjectDir, info.RoadmapFile),
			rel(info.ProjectDir, info.SessionsFile),
			artifactPath,
		},
		Contract: contract,
	}
	sort.Strings(packet.Sources)
	packet.RenderedPrompt = renderGuidePrompt(packet)
	return packet, nil
}

func (m *Manager) buildCollaborationGuidePacket(command, brainstormSlug, discussionRef, stage string) (*GuidePacket, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	brainstormSlug = strings.TrimSpace(brainstormSlug)
	discussionRef = strings.TrimSpace(discussionRef)
	meta, err := m.workspace.ReadWorkspaceMeta()
	if err != nil {
		return nil, err
	}
	branch := ""
	if githubState, err := m.workspace.ReadGitHubState(); err == nil {
		branch = strings.TrimSpace(githubState.DefaultBranch)
	}
	sourceMode, err := m.SourceMode()
	if err != nil {
		return nil, err
	}
	data, ownership, err := m.loadCollaborationSourceData(brainstormSlug, discussionRef)
	if err != nil {
		return nil, err
	}

	effectiveStage := strings.TrimSpace(stage)
	if effectiveStage == "" {
		effectiveStage = "discussion_assess"
	}
	if _, ok := collaborationGuideStages[effectiveStage]; !ok {
		return nil, fmt.Errorf("unsupported collaboration stage %q", effectiveStage)
	}

	assessment, err := m.AssessCollaborationSource(CollaborationAssessInput{
		BrainstormSlug: brainstormSlug,
		DiscussionRef:  discussionRef,
	})
	if err != nil {
		return nil, err
	}

	var draft *PromotionDraft
	if effectiveStage != "discussion_assess" {
		draft, err = m.BuildPromotionDraft(PromotionDraftInput{
			BrainstormSlug: brainstormSlug,
			DiscussionRef:  discussionRef,
		})
		if err != nil {
			return nil, err
		}
	}
	if effectiveStage == "initiative_draft" && (draft == nil || draft.ProposedInitiativeIssue == nil) {
		return nil, fmt.Errorf("guide packets for %q require a multi-spec promotion draft with an initiative issue", effectiveStage)
	}
	if effectiveStage == "spec_draft" && (draft == nil || len(draft.ProposedSpecIssues) == 0) {
		return nil, fmt.Errorf("guide packets for %q require at least one promoted spec draft", effectiveStage)
	}

	artifact := buildCollaborationArtifact(data)
	contract := collaborationGuideContract(effectiveStage, artifact.Path)
	packet := &GuidePacket{
		SchemaVersion: GuidePacketSchemaVersion,
		Kind:          guidePacketKind,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Builder: GuidePacketBuilderMeta{
			Command: command,
			Format:  "json",
		},
		Workspace: GuidePacketWorkspace{
			ProjectRoot:       info.ProjectDir,
			PlanningMode:      guidePlanningModeCollab,
			PlanningModel:     meta.PlanningModel,
			SourceMode:        string(sourceMode),
			StoryBackend:      string(meta.StoryBackend),
			IntegrationBranch: branch,
		},
		Ownership: ownership,
		Session: GuidePacketSession{
			CurrentStage: effectiveStage,
			Summary:      collaborationStageSummary(effectiveStage, assessment, draft),
			NextAction:   collaborationNextAction(effectiveStage, ownership.Mode, assessment, draft),
		},
		Artifact: artifact,
		Mode: GuidePacketMode{
			Stage: effectiveStage,
			Pass:  collaborationGuidePass(effectiveStage),
		},
		Sources: append([]string{
			rel(info.ProjectDir, info.ProjectFile),
			rel(info.ProjectDir, info.RoadmapFile),
		}, data.source.SourceLinks...),
		Collaboration: &GuidePacketCollaboration{
			Source:         data.source,
			Assessment:     assessment,
			PromotionDraft: draft,
		},
		Contract:       contract,
		RenderedDrafts: buildCollaborationRenderedDrafts(effectiveStage, draft),
		Actions:        buildCollaborationActions(effectiveStage, ownership.Mode, brainstormSlug, discussionRef, assessment, draft),
	}
	sort.Strings(packet.Sources)
	packet.RenderedPrompt = renderGuidePrompt(packet)
	return packet, nil
}

func brainstormGuidePass(checkpoint string) string {
	switch checkpoint {
	case "vision-intake":
		return "brainstorm_intake"
	case "handoff-epic":
		return "brainstorm_handoff"
	default:
		return "brainstorm_refine"
	}
}

func brainstormGuideCheckpointIndex(checkpoint string, fallback int) int {
	switch checkpoint {
	case "vision-intake":
		return 1
	case "clarify-problem-user-value":
		return 2
	case "clarify-constraints-appetite":
		return 3
	case "clarify-open-approaches":
		return 4
	case "handoff-epic":
		return 5
	default:
		if fallback > 0 {
			return fallback
		}
		return 1
	}
}

func collaborationGuidePass(stage string) string {
	switch stage {
	case "discussion_assess":
		return "collaboration_assess"
	case "promotion_review":
		return "collaboration_review"
	case "initiative_draft":
		return "initiative_review"
	case "spec_draft":
		return "spec_review"
	case "needs_refinement":
		return "refinement_review"
	default:
		return "collaboration"
	}
}

func buildCollaborationArtifact(data *collaborationSourceData) GuidePacketArtifact {
	switch data.source.Mode {
	case CollaborationSourceGitHubDiscussion:
		title := data.title
		if data.source.Discussion != nil && strings.TrimSpace(data.source.Discussion.Title) != "" {
			title = data.source.Discussion.Title
		}
		path := firstNonEmpty(discussionURL(data.source.Discussion), firstNonEmpty(data.source.SourceLinks...))
		return GuidePacketArtifact{
			Type:   "github_discussion",
			Slug:   fmt.Sprintf("discussion-%d", discussionNumber(data.source.Discussion)),
			Title:  title,
			Path:   path,
			Status: "active",
		}
	default:
		return GuidePacketArtifact{
			Type:   "brainstorm",
			Slug:   data.source.BrainstormSlug,
			Title:  data.title,
			Path:   data.source.BrainstormPath,
			Status: "active",
		}
	}
}

func collaborationGuideContract(stage, artifactPath string) GuidePacketContract {
	questionStrategy := GuidePacketQuestionStrategy{
		ClusterSizeMin:        1,
		ClusterSizeMax:        3,
		ReflectOncePerCluster: true,
		GapGuidance:           "keep_feedback_specific_and_execution_relevant",
		MenuActions:           []string{"accept", "refine", "stop_for_now"},
	}
	artifactStrategy := GuidePacketArtifactStrategy{
		WriteMode:       "review_only",
		DurableArtifact: artifactPath,
		PreserveRules: []string{
			"Do not replace the canonical discuss payloads with ad hoc packet-only data.",
			"Do not mutate the collaboration source while rendering the packet.",
			"Do not invent execution-state automation during collaboration shaping.",
		},
		StrengthenSections: collaborationStrengthenSections(stage),
	}

	return GuidePacketContract{
		Role:             "collaboration_shaping_facilitator",
		Stance:           []string{"direct", "review_focused", "skeptical_when_needed", "preserve_canonical_contracts"},
		Goal:             collaborationGuideGoal(stage),
		QuestionStrategy: questionStrategy,
		ArtifactStrategy: artifactStrategy,
		Do:               collaborationGuideDo(stage),
		Avoid: []string{
			"Do not treat the guide packet as the canonical collaboration record.",
			"Do not bypass the explicit review and confirmation flow.",
			"Do not mutate GitHub or local planning state while previewing the packet.",
		},
		QualityBar: []string{
			"The packet should make the next collaboration decision obvious without re-deriving the discuss payloads.",
			"The rendered drafts and action objects should be specific enough for an agent runtime to follow directly.",
		},
		CompletionGate: collaborationCompletionGate(stage),
		CommandHints:   collaborationCommandHints(stage),
	}
}

func collaborationGuideGoal(stage string) string {
	switch stage {
	case "discussion_assess":
		return "Assess whether the current collaboration source is mature enough for promotion."
	case "promotion_review":
		return "Review the promotion draft before any apply step can mutate GitHub or local planning state."
	case "initiative_draft":
		return "Pressure-test the proposed initiative issue body and milestone-backed multi-spec wrapper."
	case "spec_draft":
		return "Pressure-test the proposed spec issue bodies, readiness states, and dependency shape."
	case "needs_refinement":
		return "Review any explicit refinement exceptions before promotion continues."
	default:
		return "Guide the collaboration shaping flow."
	}
}

func collaborationGuideDo(stage string) []string {
	items := []string{
		"Use the embedded collaboration payloads as the canonical source of truth.",
		"Keep feedback specific to the current stage instead of reopening the whole planning flow.",
	}
	switch stage {
	case "discussion_assess":
		return append(items,
			"Focus on whether the source is not ready, single-spec ready, or multi-spec ready.",
			"Call out concrete missing inputs rather than generic uncertainty.",
		)
	case "promotion_review":
		return append(items,
			"Review the draft issue set, dependency plan, and milestone/project guidance before any apply action.",
			"Prefer specific corrections to titles, bodies, and dependencies over vague approval.",
		)
	case "initiative_draft":
		return append(items,
			"Review whether the initiative body explains outcome, scope boundary, specs, and milestone clearly.",
			"Make sure the initiative stays lightweight and does not become a second execution layer.",
		)
	case "spec_draft":
		return append(items,
			"Check that proposed spec bodies are bounded, verifiable, and correctly marked ready or blocked.",
			"Preserve the distinction between parent/sub-issue grouping and blocked-by ordering.",
		)
	case "needs_refinement":
		return append(items,
			"Review only the specs that have explicit refinement exceptions.",
			"Translate each exception into a concrete next clarification instead of letting it stay vague.",
		)
	default:
		return items
	}
}

func collaborationCompletionGate(stage string) []string {
	switch stage {
	case "discussion_assess":
		return []string{
			"It is clear whether the source is not ready, ready for one spec, or ready for multiple specs.",
			"The next collaboration move is obvious from the packet actions.",
		}
	case "promotion_review":
		return []string{
			"The proposed issue set is concrete enough to confirm or refine.",
			"Any apply action is gated behind explicit confirmation.",
		}
	case "initiative_draft":
		return []string{
			"The initiative wrapper is specific enough to anchor the multi-spec set without redefining the work.",
			"The milestone-backed grouping is clear.",
		}
	case "spec_draft":
		return []string{
			"Each spec draft has an acceptable body, readiness state, and dependency shape.",
			"The packet makes clear which specs are ready, blocked, or need refinement.",
		}
	case "needs_refinement":
		return []string{
			"Each refinement exception names a concrete gap, why it blocks execution, and what would resolve it.",
			"If no refinement exceptions exist, the packet says so explicitly.",
		}
	default:
		return []string{"The current collaboration stage is specific enough to continue without guessing."}
	}
}

func collaborationCommandHints(stage string) []GuidePacketCommandHint {
	hints := []GuidePacketCommandHint{
		{
			Purpose: "refresh_assessment",
			Command: "plan discuss assess --project . <source> --format json",
		},
		{
			Purpose: "refresh_promotion_draft",
			Command: "plan discuss promote --project . <source> --format json",
		},
	}
	switch stage {
	case "promotion_review", "initiative_draft", "spec_draft", "needs_refinement":
		hints = append(hints, GuidePacketCommandHint{
			Purpose: "apply_confirmed_promotion",
			Command: "plan discuss promote --project . <source> --apply --confirm --target <github|hybrid> --format json",
		})
	}
	return hints
}

func collaborationStrengthenSections(stage string) []string {
	switch stage {
	case "discussion_assess":
		return []string{"Problem", "Goals", "Constraints", "Non-Goals", "Proposed Shape"}
	case "promotion_review":
		return []string{"Promotion Decision", "Dependency Plan", "Milestone Plan", "Project Prompt"}
	case "initiative_draft":
		return []string{"Initiative", "Outcome", "Scope Boundary", "Specs", "Milestone"}
	case "spec_draft":
		return []string{"Spec", "Goals", "Constraints", "Verification", "Dependencies", "Readiness"}
	case "needs_refinement":
		return []string{"Refinement Gap", "Why Not Ready", "Recommended Clarification", "Exit Criteria"}
	default:
		return []string{"Problem", "Goals"}
	}
}

func collaborationStageSummary(stage string, assessment *CollaborationAssessment, draft *PromotionDraft) string {
	switch stage {
	case "discussion_assess":
		if assessment != nil {
			return defaultString(assessment.Decision.Reason, "Assess the collaboration source for promotion readiness.")
		}
	case "promotion_review", "initiative_draft", "spec_draft", "needs_refinement":
		if draft != nil {
			return defaultString(draft.WhyThisPath, "Review the promotion draft before any write step.")
		}
	}
	return "Review the collaboration shaping packet."
}

func collaborationNextAction(stage string, mode SourceOfTruthMode, assessment *CollaborationAssessment, draft *PromotionDraft) string {
	switch stage {
	case "discussion_assess":
		if assessment != nil && assessment.Decision.State == MaturityNotReady {
			return "Refine the source material until the missing gaps are resolved, then reassess."
		}
		return "Review the promotion draft before any apply action."
	case "promotion_review":
		return nextApplyAction(mode, draft)
	case "initiative_draft":
		return "Finalize the initiative wrapper details, then return to the promotion review or apply step."
	case "spec_draft":
		return "Finalize the spec drafts and dependency shape, then confirm the promotion."
	case "needs_refinement":
		if draft != nil && len(draft.NeedsRefinementExceptions) == 0 {
			return "No spec starts in needs-refinement right now; return to promotion review."
		}
		return "Resolve the refinement exceptions or accept them before applying the promotion."
	default:
		return "Continue collaboration shaping."
	}
}

func buildCollaborationRenderedDrafts(stage string, draft *PromotionDraft) []GuidePacketRenderedDraft {
	if draft == nil {
		return nil
	}
	rendered := []GuidePacketRenderedDraft{}
	switch stage {
	case "initiative_draft":
		if draft.ProposedInitiativeIssue != nil {
			rendered = append(rendered, GuidePacketRenderedDraft{
				Kind:      "initiative_issue",
				Title:     draft.ProposedInitiativeIssue.Title,
				Body:      draft.ProposedInitiativeIssue.Body,
				Slug:      draft.ProposedInitiativeIssue.Slug,
				Readiness: string(draft.ProposedInitiativeIssue.Readiness),
			})
		}
	case "spec_draft", "promotion_review":
		for _, item := range draft.ProposedSpecIssues {
			rendered = append(rendered, GuidePacketRenderedDraft{
				Kind:      "spec_issue",
				Title:     item.Title,
				Body:      item.Body,
				Slug:      item.Slug,
				Readiness: string(item.Readiness),
			})
		}
		if stage == "promotion_review" && draft.ProposedInitiativeIssue != nil {
			rendered = append([]GuidePacketRenderedDraft{{
				Kind:      "initiative_issue",
				Title:     draft.ProposedInitiativeIssue.Title,
				Body:      draft.ProposedInitiativeIssue.Body,
				Slug:      draft.ProposedInitiativeIssue.Slug,
				Readiness: string(draft.ProposedInitiativeIssue.Readiness),
			}}, rendered...)
		}
	case "needs_refinement":
		if len(draft.NeedsRefinementExceptions) == 0 {
			rendered = append(rendered, GuidePacketRenderedDraft{
				Kind:  "refinement_summary",
				Title: "No specs currently need refinement",
				Body:  "The promotion draft does not currently mark any spec as `needs-refinement`.",
			})
			return rendered
		}
		for _, item := range draft.NeedsRefinementExceptions {
			rendered = append(rendered, GuidePacketRenderedDraft{
				Kind:      "refinement_exception",
				Title:     item.IssueTitle,
				Body:      renderRefinementException(item),
				Slug:      slugify(item.IssueTitle),
				Readiness: "needs-refinement",
			})
		}
	}
	return rendered
}

func buildCollaborationActions(stage string, mode SourceOfTruthMode, brainstormSlug, discussionRef string, assessment *CollaborationAssessment, draft *PromotionDraft) []GuidePacketAction {
	sourceArg, display := collaborationSourceCommand(brainstormSlug, discussionRef)
	actions := []GuidePacketAction{
		{
			ID:        "refresh_assessment",
			Kind:      "refresh",
			Label:     "Refresh assessment",
			Command:   "plan discuss assess --project . " + sourceArg + " --format json",
			Target:    "assessment",
			Available: true,
		},
	}
	if stage != "discussion_assess" {
		actions = append(actions, GuidePacketAction{
			ID:          "refresh_promotion_draft",
			Kind:        "review",
			Label:       "Refresh promotion draft",
			Description: "Rebuild the current promotion draft from " + display + ".",
			Command:     "plan discuss promote --project . " + sourceArg + " --format json",
			Target:      "promotion_draft",
			Available:   true,
		})
	}
	if stage == "promotion_review" || stage == "initiative_draft" || stage == "spec_draft" || stage == "needs_refinement" {
		actions = append(actions, collaborationApplyActions(mode, sourceArg, draft)...)
	}
	if stage == "discussion_assess" && assessment != nil && assessment.Decision.State != MaturityNotReady {
		actions = append(actions, GuidePacketAction{
			ID:          "build_promotion_draft",
			Kind:        "review",
			Label:       "Build promotion draft",
			Description: "Move from maturity assessment into the promotion draft review.",
			Command:     "plan discuss promote --project . " + sourceArg + " --format json",
			Target:      "promotion_draft",
			Available:   true,
		})
	}
	return actions
}

func collaborationApplyActions(mode SourceOfTruthMode, sourceArg string, draft *PromotionDraft) []GuidePacketAction {
	if draft == nil {
		return nil
	}
	makeAction := func(id, label, target string, targetMode SourceOfTruthMode, available bool, blockedReason string) GuidePacketAction {
		return GuidePacketAction{
			ID:                   id,
			Kind:                 "apply",
			Label:                label,
			Description:          "Apply the reviewed promotion draft to " + string(targetMode) + ".",
			Command:              "plan discuss promote --project . " + sourceArg + " --apply --confirm --target " + string(targetMode) + " --format json",
			Target:               target,
			RequiresConfirmation: true,
			Available:            available,
			BlockedReason:        blockedReason,
		}
	}

	switch mode {
	case SourceOfTruthGitHub:
		return []GuidePacketAction{makeAction("apply_github", "Apply to GitHub", "promotion_apply", SourceOfTruthGitHub, true, "")}
	case SourceOfTruthHybrid:
		return []GuidePacketAction{makeAction("apply_hybrid", "Apply to Hybrid", "promotion_apply", SourceOfTruthHybrid, true, "")}
	default:
		return []GuidePacketAction{
			makeAction("apply_github", "Apply to GitHub", "promotion_apply", SourceOfTruthGitHub, true, ""),
			makeAction("apply_hybrid", "Apply to Hybrid", "promotion_apply", SourceOfTruthHybrid, true, ""),
		}
	}
}

func collaborationSourceCommand(brainstormSlug, discussionRef string) (string, string) {
	brainstormSlug = strings.TrimSpace(brainstormSlug)
	discussionRef = strings.TrimSpace(discussionRef)
	if brainstormSlug != "" {
		return "--brainstorm " + brainstormSlug, "local brainstorm " + brainstormSlug
	}
	return "--discussion " + discussionRef, "GitHub Discussion " + discussionRef
}

func nextApplyAction(mode SourceOfTruthMode, draft *PromotionDraft) string {
	if draft == nil {
		return "Refresh the promotion draft before applying it."
	}
	switch mode {
	case SourceOfTruthGitHub:
		return "Review the draft and confirm `plan discuss promote --apply --confirm --target github` when it is ready."
	case SourceOfTruthHybrid:
		return "Review the draft and confirm `plan discuss promote --apply --confirm --target hybrid` when it is ready."
	default:
		return "Review the draft and choose whether to apply it to `github` or `hybrid` ownership."
	}
}

func renderRefinementException(item PromotionRefinementException) string {
	lines := []string{
		"## Refinement Gap",
		defaultString(item.Gap, "-"),
		"",
		"## Why Not Ready",
		defaultString(item.WhyNotReady, "-"),
		"",
		"## Recommended Clarification",
		defaultString(item.RecommendedClarification, "-"),
	}
	if len(item.ExitCriteria) > 0 {
		lines = append(lines, "", "## Exit Criteria")
		for _, criterion := range item.ExitCriteria {
			lines = append(lines, "- "+criterion)
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func brainstormGuideContract(session *workspace.GuidedSessionRecord, checkpoint, artifactPath string) GuidePacketContract {
	questionStrategy := GuidePacketQuestionStrategy{
		ClusterSizeMin:        2,
		ClusterSizeMax:        4,
		ReflectOncePerCluster: true,
		GapGuidance:           "one_recommended_plus_up_to_two_alternatives",
		MenuActions:           []string{"continue", "refine", "stop_for_now"},
	}
	artifactStrategy := GuidePacketArtifactStrategy{
		WriteMode:       "additive",
		DurableArtifact: artifactPath,
		PreserveRules: []string{
			"Use the user's own language when it clarifies intent.",
			"Do not invent a new durable planning layer during brainstorm guidance.",
			"Do not draft implementation slices during brainstorm guidance.",
		},
		StrengthenSections: brainstormStrengthenSections(checkpoint),
	}

	goal := "Turn the brainstorm into a clearer, promotable planning artifact."
	do := []string{
		"Ask small clusters of focused questions instead of dumping a long form.",
		"Reflect back what changed before moving to the next checkpoint.",
		"Keep scope bounded and surface simpler alternatives when the work sprawls.",
	}
	avoid := []string{
		"Do not call model APIs from `plan`.",
		"Do not mutate guided session state while rendering the packet.",
		"Do not jump into execution slices during brainstorm guidance.",
	}
	quality := []string{
		"The user should leave this checkpoint with clearer scope and fewer hidden assumptions.",
		"The brainstorm note should be stronger than it was before the checkpoint started.",
	}
	completion := []string{
		"The checkpoint-specific sections are strong enough to keep moving without guessing.",
		"The recap and next action are specific enough for the next guided move.",
	}
	commands := []GuidePacketCommandHint{
		{
			Purpose: "resume_current_stage",
			Command: "plan brainstorm resume " + session.Brainstorm + " --project .",
		},
		{
			Purpose: "preview_current_checkpoint",
			Command: "plan guide show --project . --chain " + session.ChainID + " --stage brainstorm --checkpoint " + checkpoint + " --format json",
		},
	}

	switch checkpoint {
	case "vision-intake":
		goal = "Capture the vision and supporting material before deeper shaping starts."
		do = append(do,
			"Start from the outcome the user sees in their head.",
			"Capture any supporting docs, links, or references that should shape later refinement.",
		)
		completion = []string{
			"The Vision section describes the outcome in plain language.",
			"Supporting Material captures any docs, links, or references the user wants to carry forward.",
		}
	case "clarify-problem-user-value":
		goal = "Clarify the problem and the user value so the brainstorm has a concrete center of gravity."
	case "clarify-constraints-appetite":
		goal = "Set the hard boundaries and appetite that keep the work honest."
		do = append(do, "Push for at least one hard boundary the work will not cross.")
	case "clarify-open-approaches":
		goal = "Trim the open questions down to the real blockers and identify the best candidate approaches."
		do = append(do, "Prefer one recommended path plus up to two alternatives.")
	case "handoff-epic":
		goal = "Prepare the brainstorm for the next durable planning step."
		do = append(do, "Call out whether the work is small enough for one spec or broad enough to justify multiple specs under one initiative.")
		avoid = append(avoid, "Do not force a multi-spec initiative when the work is still one bounded spec.")
		completion = []string{
			"It is clear whether the work should stay one spec or split into multiple specs.",
			"The recap is strong enough to hand the work into the next planning step without re-interviewing the user.",
		}
	}

	return GuidePacketContract{
		Role:             "co_planning_facilitator",
		Stance:           []string{"collaborative", "direct", "skeptical_when_needed", "keep_scope_small"},
		Goal:             goal,
		QuestionStrategy: questionStrategy,
		ArtifactStrategy: artifactStrategy,
		Do:               do,
		Avoid:            avoid,
		QualityBar:       quality,
		CompletionGate:   completion,
		CommandHints:     commands,
	}
}

func brainstormStrengthenSections(checkpoint string) []string {
	switch checkpoint {
	case "vision-intake":
		return []string{"Vision", "Supporting Material"}
	case "clarify-problem-user-value":
		return []string{"Problem", "User / Value"}
	case "clarify-constraints-appetite":
		return []string{"Constraints", "Appetite"}
	case "clarify-open-approaches":
		return []string{"Remaining Open Questions", "Candidate Approaches"}
	case "handoff-epic":
		return []string{"Decision Snapshot", "Problem", "User / Value", "Constraints"}
	default:
		return []string{"Vision"}
	}
}

func renderGuidePrompt(packet *GuidePacket) string {
	if packet.Collaboration != nil {
		return renderCollaborationGuidePrompt(packet)
	}
	return renderBrainstormGuidePrompt(packet)
}

func renderBrainstormGuidePrompt(packet *GuidePacket) string {
	var lines []string
	lines = append(lines, "You are guiding the brainstorm stage for `plan`.")
	lines = append(lines, "Goal: "+packet.Contract.Goal)
	lines = append(lines, "Current summary: "+defaultString(packet.Session.Summary, "No recap saved yet."))
	lines = append(lines, "Next action: "+defaultString(packet.Session.NextAction, "Continue guided clarification."))
	lines = append(lines, "Durable artifact: "+packet.Contract.ArtifactStrategy.DurableArtifact)
	lines = append(lines, "Checkpoint: "+packet.Mode.Checkpoint)
	lines = append(lines, "Do:")
	for _, item := range packet.Contract.Do {
		lines = append(lines, "- "+item)
	}
	lines = append(lines, "Avoid:")
	for _, item := range packet.Contract.Avoid {
		lines = append(lines, "- "+item)
	}
	lines = append(lines, "Completion gate:")
	for _, item := range packet.Contract.CompletionGate {
		lines = append(lines, "- "+item)
	}
	return strings.Join(lines, "\n")
}

func renderCollaborationGuidePrompt(packet *GuidePacket) string {
	var lines []string
	source := packet.Collaboration.Source.CanonicalSource
	lines = append(lines, "You are guiding the collaboration shaping stage for `plan`.")
	lines = append(lines, "Goal: "+packet.Contract.Goal)
	lines = append(lines, "Stage: "+packet.Mode.Stage)
	lines = append(lines, "Source: "+defaultString(source, "collaboration_source"))
	lines = append(lines, "Current summary: "+defaultString(packet.Session.Summary, "No collaboration summary saved yet."))
	lines = append(lines, "Next action: "+defaultString(packet.Session.NextAction, "Review the current collaboration packet."))
	lines = append(lines, "Durable artifact: "+packet.Contract.ArtifactStrategy.DurableArtifact)
	if packet.Collaboration.Assessment != nil {
		lines = append(lines, "Assessment: "+string(packet.Collaboration.Assessment.Decision.State))
	}
	if packet.Collaboration.PromotionDraft != nil && packet.Collaboration.PromotionDraft.PromotionDecision != "" {
		lines = append(lines, "Promotion decision: "+string(packet.Collaboration.PromotionDraft.PromotionDecision))
	}
	lines = append(lines, "Do:")
	for _, item := range packet.Contract.Do {
		lines = append(lines, "- "+item)
	}
	lines = append(lines, "Avoid:")
	for _, item := range packet.Contract.Avoid {
		lines = append(lines, "- "+item)
	}
	lines = append(lines, "Completion gate:")
	for _, item := range packet.Contract.CompletionGate {
		lines = append(lines, "- "+item)
	}
	if len(packet.Actions) > 0 {
		lines = append(lines, "Actions:")
		for _, action := range packet.Actions {
			status := "available"
			if !action.Available {
				status = "blocked"
			}
			line := fmt.Sprintf("- %s (%s)", action.Label, status)
			if action.RequiresConfirmation {
				line += " requires confirmation"
			}
			if strings.TrimSpace(action.BlockedReason) != "" {
				line += ": " + action.BlockedReason
			}
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func copyStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
