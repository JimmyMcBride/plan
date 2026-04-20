package planning

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"plan/internal/notes"
	"plan/internal/workspace"
)

type GuidedBrainstormIntakeInput struct {
	Vision             string
	SupportingMaterial string
}

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
