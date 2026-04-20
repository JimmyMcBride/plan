package planning

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"plan/internal/notes"
	"plan/internal/workspace"
)

type GuidedBrainstormIntakeInput struct {
	Vision             string
	SupportingMaterial string
}

type GuidedSessionUpdateInput struct {
	CurrentStage        string
	CurrentCluster      int
	CurrentClusterLabel string
	Summary             string
	NextAction          string
	StageStatus         string
}

var guidedStageOrder = []string{"brainstorm", "epic", "spec", "stories"}

func (m *Manager) EnsureGuidedBrainstormSession(brainstormSlug string) (*workspace.GuidedSessionRecord, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	brainstormPath := filepath.Join(info.BrainstormsDir, slugify(brainstormSlug)+".md")
	note, err := notes.Read(brainstormPath)
	if err != nil {
		return nil, err
	}
	if note.Type != "brainstorm" {
		return nil, fmt.Errorf("%s is not a brainstorm note", note.Path)
	}
	return m.upsertGuidedBrainstormSession(note)
}

func (m *Manager) UpdateGuidedBrainstormIntake(brainstormSlug string, input GuidedBrainstormIntakeInput) (*notes.Note, *workspace.GuidedSessionRecord, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, nil, err
	}
	path := filepath.Join(info.BrainstormsDir, slugify(brainstormSlug)+".md")
	note, err := notes.Read(path)
	if err != nil {
		return nil, nil, err
	}
	if note.Type != "brainstorm" {
		return nil, nil, fmt.Errorf("%s is not a brainstorm note", note.Path)
	}

	body := note.Content
	if strings.TrimSpace(input.Vision) != "" {
		body = notes.SetSection(body, "Vision", strings.TrimSpace(input.Vision))
	}
	if strings.TrimSpace(input.SupportingMaterial) != "" {
		body = notes.SetSection(body, "Supporting Material", normalizeBulletList(input.SupportingMaterial))
	}

	updated, err := notes.Update(path, notes.UpdateInput{Body: &body})
	if err != nil {
		return nil, nil, err
	}
	session, err := m.upsertGuidedBrainstormSession(updated)
	if err != nil {
		return nil, nil, err
	}
	return m.relNote(updated, info.ProjectDir), session, nil
}

func (m *Manager) ReadGuidedSession(chainID string) (*workspace.GuidedSessionRecord, error) {
	state, err := m.workspace.ReadGuidedSessionState()
	if err != nil {
		return nil, err
	}
	record, ok := state.Sessions[strings.TrimSpace(chainID)]
	if !ok {
		return nil, fmt.Errorf("guided session %q not found", chainID)
	}
	return &record, nil
}

func (m *Manager) ReadLastActiveGuidedSession() (*workspace.GuidedSessionRecord, error) {
	state, err := m.workspace.ReadGuidedSessionState()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(state.LastActiveChain) == "" {
		return nil, fmt.Errorf("no active guided session")
	}
	return m.ReadGuidedSession(state.LastActiveChain)
}

func (m *Manager) ListGuidedSessions() ([]workspace.GuidedSessionRecord, error) {
	state, err := m.workspace.ReadGuidedSessionState()
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(state.Sessions))
	for key := range state.Sessions {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]workspace.GuidedSessionRecord, 0, len(keys))
	for _, key := range keys {
		out = append(out, state.Sessions[key])
	}
	return out, nil
}

func (m *Manager) SwitchGuidedSession(chainID string) (*workspace.GuidedSessionRecord, error) {
	state, err := m.workspace.ReadGuidedSessionState()
	if err != nil {
		return nil, err
	}
	record, ok := state.Sessions[strings.TrimSpace(chainID)]
	if !ok {
		return nil, fmt.Errorf("guided session %q not found", chainID)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	record.UpdatedAt = now
	state.LastActiveChain = record.ChainID
	state.LastUpdatedAt = now
	state.Sessions[record.ChainID] = record
	if err := m.workspace.WriteGuidedSessionState(*state); err != nil {
		return nil, err
	}
	return &record, nil
}

func (m *Manager) UpdateGuidedSession(chainID string, input GuidedSessionUpdateInput) (*workspace.GuidedSessionRecord, error) {
	state, err := m.workspace.ReadGuidedSessionState()
	if err != nil {
		return nil, err
	}
	chainID = strings.TrimSpace(chainID)
	record, ok := state.Sessions[chainID]
	if !ok {
		return nil, fmt.Errorf("guided session %q not found", chainID)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if strings.TrimSpace(input.CurrentStage) != "" {
		record.CurrentStage = strings.TrimSpace(input.CurrentStage)
	}
	if input.CurrentCluster > 0 {
		record.CurrentCluster = input.CurrentCluster
	}
	if strings.TrimSpace(input.CurrentClusterLabel) != "" {
		record.CurrentClusterLabel = strings.TrimSpace(input.CurrentClusterLabel)
	}
	if strings.TrimSpace(input.Summary) != "" {
		record.Summary = strings.TrimSpace(input.Summary)
	}
	if strings.TrimSpace(input.NextAction) != "" {
		record.NextAction = strings.TrimSpace(input.NextAction)
	}
	if record.StageStatuses == nil {
		record.StageStatuses = map[string]string{}
	}
	stageKey := record.CurrentStage
	if stageKey == "" {
		stageKey = "brainstorm"
	}
	if strings.TrimSpace(input.StageStatus) != "" {
		record.StageStatuses[stageKey] = strings.TrimSpace(input.StageStatus)
	}
	record.UpdatedAt = now
	state.LastActiveChain = chainID
	state.LastUpdatedAt = now
	state.Sessions[chainID] = record
	if err := m.workspace.WriteGuidedSessionState(*state); err != nil {
		return nil, err
	}
	return &record, nil
}

func (m *Manager) ReopenGuidedSessionStage(chainID, stage string) (*workspace.GuidedSessionRecord, []string, error) {
	stage = strings.TrimSpace(stage)
	if stage == "" {
		return nil, nil, fmt.Errorf("stage is required")
	}
	state, err := m.workspace.ReadGuidedSessionState()
	if err != nil {
		return nil, nil, err
	}
	chainID = strings.TrimSpace(chainID)
	record, ok := state.Sessions[chainID]
	if !ok {
		return nil, nil, fmt.Errorf("guided session %q not found", chainID)
	}
	if !isGuidedStage(stage) {
		return nil, nil, fmt.Errorf("unsupported guided stage %q", stage)
	}
	if record.StageStatuses == nil {
		record.StageStatuses = map[string]string{}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	record.CurrentStage = stage
	record.StageStatuses[stage] = "in_progress"
	downstream := guidedDownstreamStages(stage)
	for _, later := range downstream {
		record.StageStatuses[later] = "needs_review"
	}
	record.NextAction = fmt.Sprintf("Review %s after reopening %s.", strings.Join(downstream, ", "), stage)
	record.UpdatedAt = now
	state.LastActiveChain = chainID
	state.LastUpdatedAt = now
	state.Sessions[chainID] = record
	if err := m.workspace.WriteGuidedSessionState(*state); err != nil {
		return nil, nil, err
	}
	return &record, downstream, nil
}

func (m *Manager) upsertGuidedBrainstormSession(note *notes.Note) (*workspace.GuidedSessionRecord, error) {
	state, err := m.workspace.ReadGuidedSessionState()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	slug := slugFromPath(note.Path)
	chainID := guidedChainID(slug)
	record, exists := state.Sessions[chainID]
	if !exists {
		record = workspace.GuidedSessionRecord{
			ChainID:       chainID,
			Brainstorm:    slug,
			StageStatuses: map[string]string{},
			CreatedAt:     now,
		}
	}
	if record.Brainstorm == "" {
		record.Brainstorm = slug
	}
	if record.CurrentStage == "" {
		record.CurrentStage = "brainstorm"
	}
	if record.CurrentCluster == 0 {
		record.CurrentCluster = 1
	}
	if record.CurrentClusterLabel == "" {
		record.CurrentClusterLabel = "vision-intake"
	}
	if record.StageStatuses == nil {
		record.StageStatuses = map[string]string{}
	}
	record.StageStatuses["brainstorm"] = "in_progress"
	record.Summary = summarizeBrainstormSession(note)
	record.NextAction = nextBrainstormSessionAction(note)
	record.UpdatedAt = now

	state.LastActiveChain = chainID
	state.LastUpdatedAt = now
	state.Sessions[chainID] = record
	if err := m.workspace.WriteGuidedSessionState(*state); err != nil {
		return nil, err
	}
	return &record, nil
}

func guidedChainID(brainstormSlug string) string {
	return fmt.Sprintf("brainstorm/%s", slugify(brainstormSlug))
}

func isGuidedStage(stage string) bool {
	for _, candidate := range guidedStageOrder {
		if candidate == stage {
			return true
		}
	}
	return false
}

func guidedDownstreamStages(stage string) []string {
	for index, candidate := range guidedStageOrder {
		if candidate == stage {
			return append([]string(nil), guidedStageOrder[index+1:]...)
		}
	}
	return nil
}

func summarizeBrainstormSession(note *notes.Note) string {
	var parts []string
	if strings.TrimSpace(notes.ExtractSection(note.Content, "Vision")) != "" {
		parts = append(parts, "Vision captured.")
	} else {
		parts = append(parts, "Vision still missing.")
	}
	if strings.TrimSpace(notes.ExtractSection(note.Content, "Supporting Material")) != "" {
		parts = append(parts, "Supporting material recorded.")
	} else {
		parts = append(parts, "No supporting material recorded yet.")
	}
	return strings.Join(parts, " ")
}

func nextBrainstormSessionAction(note *notes.Note) string {
	if strings.TrimSpace(notes.ExtractSection(note.Content, "Vision")) == "" {
		return "Capture the user's vision in plain language."
	}
	if strings.TrimSpace(notes.ExtractSection(note.Content, "Supporting Material")) == "" {
		return "Ask the user for any relevant docs, links, or research to carry into the brainstorm."
	}
	return "Continue guided brainstorm clarification."
}
