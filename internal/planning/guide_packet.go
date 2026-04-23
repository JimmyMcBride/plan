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
	GuidePacketSchemaVersion = 1
	guidePacketKind          = "guide_packet"
	guidePlanningMode        = "guided"
)

var brainstormGuideCheckpoints = map[string]struct{}{
	"vision-intake":                {},
	"clarify-problem-user-value":   {},
	"clarify-constraints-appetite": {},
	"clarify-open-approaches":      {},
	"handoff-epic":                 {},
}

type GuidePacket struct {
	SchemaVersion  int                    `json:"schema_version"`
	Kind           string                 `json:"kind"`
	GeneratedAt    string                 `json:"generated_at"`
	Builder        GuidePacketBuilderMeta `json:"builder"`
	Workspace      GuidePacketWorkspace   `json:"workspace"`
	Ownership      CollaborationOwnership `json:"ownership"`
	Session        GuidePacketSession     `json:"session"`
	Artifact       GuidePacketArtifact    `json:"artifact"`
	Mode           GuidePacketMode        `json:"mode"`
	Sources        []string               `json:"sources"`
	Contract       GuidePacketContract    `json:"contract"`
	RenderedPrompt string                 `json:"rendered_prompt"`
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

func (m *Manager) CurrentGuidePacket() (*GuidePacket, error) {
	session, err := m.ReadLastActiveGuidedSession()
	if err != nil {
		if errors.Is(err, ErrNoActiveGuidedSession) {
			return nil, fmt.Errorf("no active guided session. Start one with `plan brainstorm start --project . \"<topic>\"`: %w", ErrNoActiveGuidedSession)
		}
		return nil, err
	}
	return m.buildGuidePacket("plan guide current", session, "", "")
}

func (m *Manager) GuidePacketForChain(chainID, stage, checkpoint string) (*GuidePacket, error) {
	session, err := m.ReadGuidedSession(strings.TrimSpace(chainID))
	if err != nil {
		return nil, err
	}
	return m.buildGuidePacket("plan guide show", session, stage, checkpoint)
}

func (m *Manager) buildGuidePacket(command string, session *workspace.GuidedSessionRecord, stageOverride, checkpointOverride string) (*GuidePacket, error) {
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
		return nil, fmt.Errorf("guide packets currently only support the brainstorm stage; use `--stage brainstorm` or switch back to a brainstorm session")
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
			PlanningMode:      guidePlanningMode,
			PlanningModel:     meta.PlanningModel,
			SourceMode:        string(meta.SourceMode),
			StoryBackend:      string(meta.StoryBackend),
			IntegrationBranch: branch,
		},
		Ownership: buildOwnership(meta.SourceMode, EntryModeLocalPromotion),
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
