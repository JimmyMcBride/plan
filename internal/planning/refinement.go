package planning

import (
	"fmt"
	"path/filepath"
	"strings"

	"plan/internal/notes"
)

type BrainstormRefinement struct {
	Path                   string
	Problem                string
	UserValue              string
	Constraints            string
	Appetite               string
	RemainingOpenQuestions string
	CandidateApproaches    string
	DecisionSnapshot       string
}

type BrainstormRefinementInput struct {
	Problem                string
	UserValue              string
	Constraints            string
	Appetite               string
	RemainingOpenQuestions string
	CandidateApproaches    string
	DecisionSnapshot       string
}

func (r BrainstormRefinement) HasGaps() bool {
	return strings.TrimSpace(r.Problem) == "" ||
		strings.TrimSpace(r.UserValue) == "" ||
		strings.TrimSpace(r.Constraints) == "" ||
		strings.TrimSpace(r.Appetite) == "" ||
		strings.TrimSpace(r.RemainingOpenQuestions) == "" ||
		strings.TrimSpace(r.CandidateApproaches) == "" ||
		strings.TrimSpace(r.DecisionSnapshot) == ""
}

func (m *Manager) ReadBrainstormRefinement(brainstormSlug string) (*BrainstormRefinement, error) {
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
	state := extractBrainstormRefinement(note)
	state.Path = rel(info.ProjectDir, note.Path)
	return &state, nil
}

func (m *Manager) UpdateBrainstormRefinement(brainstormSlug string, input BrainstormRefinementInput) (*notes.Note, error) {
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

	current := extractBrainstormRefinement(note)
	if strings.TrimSpace(input.Problem) != "" {
		current.Problem = strings.TrimSpace(input.Problem)
	}
	if strings.TrimSpace(input.UserValue) != "" {
		current.UserValue = strings.TrimSpace(input.UserValue)
	}
	if strings.TrimSpace(input.Constraints) != "" {
		current.Constraints = normalizeBulletList(input.Constraints)
	}
	if strings.TrimSpace(input.Appetite) != "" {
		current.Appetite = strings.TrimSpace(input.Appetite)
	}
	if strings.TrimSpace(input.RemainingOpenQuestions) != "" {
		current.RemainingOpenQuestions = normalizeBulletList(input.RemainingOpenQuestions)
	}
	if strings.TrimSpace(input.CandidateApproaches) != "" {
		current.CandidateApproaches = normalizeBulletList(input.CandidateApproaches)
	}
	if strings.TrimSpace(input.DecisionSnapshot) != "" {
		current.DecisionSnapshot = strings.TrimSpace(input.DecisionSnapshot)
	}

	body := note.Content
	if strings.TrimSpace(current.Constraints) != "" {
		body = notes.SetSection(body, "Constraints", current.Constraints)
	}
	if strings.TrimSpace(current.RemainingOpenQuestions) != "" {
		body = notes.SetSection(body, "Open Questions", current.RemainingOpenQuestions)
	}
	body = notes.SetSection(body, "Refinement", renderBrainstormRefinement(current))

	updated, err := notes.Update(path, notes.UpdateInput{Body: &body})
	if err != nil {
		return nil, err
	}
	return m.relNote(updated, info.ProjectDir), nil
}

func extractBrainstormRefinement(note *notes.Note) BrainstormRefinement {
	return BrainstormRefinement{
		Problem:                extractSubsection(note.Content, "Refinement", "Problem"),
		UserValue:              extractSubsection(note.Content, "Refinement", "User / Value"),
		Constraints:            notes.ExtractSection(note.Content, "Constraints"),
		Appetite:               extractSubsection(note.Content, "Refinement", "Appetite"),
		RemainingOpenQuestions: extractSubsection(note.Content, "Refinement", "Remaining Open Questions"),
		CandidateApproaches:    extractSubsection(note.Content, "Refinement", "Candidate Approaches"),
		DecisionSnapshot:       extractSubsection(note.Content, "Refinement", "Decision Snapshot"),
	}
}

func renderBrainstormRefinement(refinement BrainstormRefinement) string {
	var lines []string
	appendSubsection(&lines, "Problem", refinement.Problem)
	appendSubsection(&lines, "User / Value", refinement.UserValue)
	appendSubsection(&lines, "Appetite", refinement.Appetite)
	appendSubsection(&lines, "Remaining Open Questions", refinement.RemainingOpenQuestions)
	appendSubsection(&lines, "Candidate Approaches", refinement.CandidateApproaches)
	appendSubsection(&lines, "Decision Snapshot", refinement.DecisionSnapshot)
	return strings.Join(lines, "\n")
}
