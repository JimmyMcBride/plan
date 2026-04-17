package planning

import (
	"fmt"
	"path/filepath"
	"strings"

	"plan/internal/notes"
)

type BrainstormChallenge struct {
	Path                   string
	RabbitHoles            string
	NoGos                  string
	Assumptions            string
	LikelyOverengineering  string
	SimplerAlternative     string
}

type BrainstormChallengeInput struct {
	RabbitHoles           string
	NoGos                 string
	Assumptions           string
	LikelyOverengineering string
	SimplerAlternative    string
}

func (c BrainstormChallenge) HasGaps() bool {
	return strings.TrimSpace(c.RabbitHoles) == "" ||
		strings.TrimSpace(c.NoGos) == "" ||
		strings.TrimSpace(c.Assumptions) == "" ||
		strings.TrimSpace(c.LikelyOverengineering) == "" ||
		strings.TrimSpace(c.SimplerAlternative) == ""
}

func (m *Manager) ReadBrainstormChallenge(brainstormSlug string) (*BrainstormChallenge, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	note, err := notes.Read(filepath.Join(info.BrainstormsDir, slugify(brainstormSlug)+".md"))
	if err != nil {
		return nil, err
	}
	if note.Type != "brainstorm" {
		return nil, fmt.Errorf("%s is not a brainstorm note", note.Path)
	}
	state := extractBrainstormChallenge(note)
	state.Path = rel(info.ProjectDir, note.Path)
	return &state, nil
}

func (m *Manager) UpdateBrainstormChallenge(brainstormSlug string, input BrainstormChallengeInput) (*notes.Note, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(info.BrainstormsDir, slugify(brainstormSlug)+".md")
	note, err := notes.Read(path)
	if err != nil {
		return nil, err
	}
	if note.Type != "brainstorm" {
		return nil, fmt.Errorf("%s is not a brainstorm note", note.Path)
	}

	current := extractBrainstormChallenge(note)
	if strings.TrimSpace(input.RabbitHoles) != "" {
		current.RabbitHoles = normalizeBulletList(input.RabbitHoles)
	}
	if strings.TrimSpace(input.NoGos) != "" {
		current.NoGos = normalizeBulletList(input.NoGos)
	}
	if strings.TrimSpace(input.Assumptions) != "" {
		current.Assumptions = normalizeBulletList(input.Assumptions)
	}
	if strings.TrimSpace(input.LikelyOverengineering) != "" {
		current.LikelyOverengineering = strings.TrimSpace(input.LikelyOverengineering)
	}
	if strings.TrimSpace(input.SimplerAlternative) != "" {
		current.SimplerAlternative = strings.TrimSpace(input.SimplerAlternative)
	}

	body := notes.SetSection(note.Content, "Challenge", renderBrainstormChallenge(current))
	updated, err := notes.Update(path, notes.UpdateInput{Body: &body})
	if err != nil {
		return nil, err
	}
	return m.relNote(updated, info.ProjectDir), nil
}

func extractBrainstormChallenge(note *notes.Note) BrainstormChallenge {
	return BrainstormChallenge{
		RabbitHoles:           extractSubsection(note.Content, "Challenge", "Rabbit Holes"),
		NoGos:                 extractSubsection(note.Content, "Challenge", "No-Gos"),
		Assumptions:           extractSubsection(note.Content, "Challenge", "Assumptions"),
		LikelyOverengineering: extractSubsection(note.Content, "Challenge", "Likely Overengineering"),
		SimplerAlternative:    extractSubsection(note.Content, "Challenge", "Simpler Alternative"),
	}
}

func renderBrainstormChallenge(challenge BrainstormChallenge) string {
	var lines []string
	appendSubsection(&lines, "Rabbit Holes", challenge.RabbitHoles)
	appendSubsection(&lines, "No-Gos", challenge.NoGos)
	appendSubsection(&lines, "Assumptions", challenge.Assumptions)
	appendSubsection(&lines, "Likely Overengineering", challenge.LikelyOverengineering)
	appendSubsection(&lines, "Simpler Alternative", challenge.SimplerAlternative)
	return strings.Join(lines, "\n")
}
