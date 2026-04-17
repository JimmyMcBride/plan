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
		current.Constraints = normalizeBrainstormList(input.Constraints)
	}
	if strings.TrimSpace(input.Appetite) != "" {
		current.Appetite = strings.TrimSpace(input.Appetite)
	}
	if strings.TrimSpace(input.RemainingOpenQuestions) != "" {
		current.RemainingOpenQuestions = normalizeBrainstormList(input.RemainingOpenQuestions)
	}
	if strings.TrimSpace(input.CandidateApproaches) != "" {
		current.CandidateApproaches = normalizeBrainstormList(input.CandidateApproaches)
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
		Problem:                notes.ExtractSection(note.Content, "Problem"),
		UserValue:              notes.ExtractSection(note.Content, "User / Value"),
		Constraints:            notes.ExtractSection(note.Content, "Constraints"),
		Appetite:               notes.ExtractSection(note.Content, "Appetite"),
		RemainingOpenQuestions: notes.ExtractSection(note.Content, "Open Questions"),
		CandidateApproaches:    notes.ExtractSection(note.Content, "Candidate Approaches"),
		DecisionSnapshot:       notes.ExtractSection(note.Content, "Decision Snapshot"),
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

func appendSubsection(lines *[]string, heading, body string) {
	if len(*lines) > 0 {
		*lines = append(*lines, "")
	}
	*lines = append(*lines, "### "+heading, "")
	body = strings.TrimSpace(body)
	if body == "" {
		return
	}
	*lines = append(*lines, strings.Split(body, "\n")...)
}

func normalizeBrainstormList(body string) string {
	var items []string
	for _, line := range strings.Split(strings.TrimSpace(body), "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "))
		if line == "" {
			continue
		}
		items = append(items, "- "+line)
	}
	if len(items) == 0 {
		return ""
	}
	return strings.Join(items, "\n")
}
