package planning

import (
	"plan/internal/notes"
)

type InitiativeRef struct {
	Slug    string
	Title   string
	Summary string
}

func (m *Manager) SetSpecInitiative(specSlug string, initiative InitiativeRef) (*notes.Note, error) {
	return m.UpdateSpec(specSlug, notes.UpdateInput{
		Metadata: map[string]any{
			"initiative":         initiative.Slug,
			"initiative_title":   initiative.Title,
			"initiative_summary": initiative.Summary,
		},
	})
}

func (m *Manager) ClearSpecInitiative(specSlug string) (*notes.Note, error) {
	return m.SetSpecInitiative(specSlug, InitiativeRef{})
}
