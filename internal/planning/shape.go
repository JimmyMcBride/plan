package planning

import (
	"fmt"
	"path/filepath"
	"strings"

	"plan/internal/notes"
)

type EpicShape struct {
	Path          string
	Appetite      string
	Outcome       string
	ScopeBoundary string
	OutOfScope    string
	SuccessSignal string
}

type EpicShapeInput struct {
	Appetite      string
	Outcome       string
	ScopeBoundary string
	OutOfScope    string
	SuccessSignal string
}

func (s EpicShape) HasGaps() bool {
	return strings.TrimSpace(s.Appetite) == "" ||
		strings.TrimSpace(s.Outcome) == "" ||
		strings.TrimSpace(s.ScopeBoundary) == "" ||
		strings.TrimSpace(s.OutOfScope) == "" ||
		strings.TrimSpace(s.SuccessSignal) == ""
}

func (m *Manager) ReadEpicShape(epicSlug string) (*EpicShape, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	note, err := notes.Read(filepath.Join(info.EpicsDir, slugify(epicSlug)+".md"))
	if err != nil {
		return nil, err
	}
	if note.Type != "epic" {
		return nil, fmt.Errorf("%s is not an epic note", note.Path)
	}
	state := extractEpicShape(note)
	state.Path = rel(info.ProjectDir, note.Path)
	return &state, nil
}

func (m *Manager) UpdateEpicShape(epicSlug string, input EpicShapeInput) (*notes.Note, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(info.EpicsDir, slugify(epicSlug)+".md")
	note, err := notes.Read(path)
	if err != nil {
		return nil, err
	}
	if note.Type != "epic" {
		return nil, fmt.Errorf("%s is not an epic note", note.Path)
	}

	current := extractEpicShape(note)
	if strings.TrimSpace(input.Appetite) != "" {
		current.Appetite = strings.TrimSpace(input.Appetite)
	}
	if strings.TrimSpace(input.Outcome) != "" {
		current.Outcome = strings.TrimSpace(input.Outcome)
	}
	if strings.TrimSpace(input.ScopeBoundary) != "" {
		current.ScopeBoundary = strings.TrimSpace(input.ScopeBoundary)
	}
	if strings.TrimSpace(input.OutOfScope) != "" {
		current.OutOfScope = normalizeBulletList(input.OutOfScope)
	}
	if strings.TrimSpace(input.SuccessSignal) != "" {
		current.SuccessSignal = strings.TrimSpace(input.SuccessSignal)
	}

	body := note.Content
	if strings.TrimSpace(current.Outcome) != "" {
		body = notes.SetSection(body, "Outcome", current.Outcome)
	}
	if scopeSummary := renderEpicScopeSummary(current.ScopeBoundary, current.OutOfScope); scopeSummary != "" {
		body = notes.SetSection(body, "Scope Boundary", scopeSummary)
	}
	body = notes.SetSection(body, "Shape", renderEpicShape(current))

	updated, err := notes.Update(path, notes.UpdateInput{Body: &body})
	if err != nil {
		return nil, err
	}
	return m.relNote(updated, info.ProjectDir), nil
}

func extractEpicShape(note *notes.Note) EpicShape {
	return EpicShape{
		Appetite:      extractSubsection(note.Content, "Shape", "Appetite"),
		Outcome:       extractSubsection(note.Content, "Shape", "Outcome"),
		ScopeBoundary: extractSubsection(note.Content, "Shape", "Scope Boundary"),
		OutOfScope:    extractSubsection(note.Content, "Shape", "Out of Scope"),
		SuccessSignal: extractSubsection(note.Content, "Shape", "Success Signal"),
	}
}

func renderEpicShape(shape EpicShape) string {
	var lines []string
	appendSubsection(&lines, "Appetite", shape.Appetite)
	appendSubsection(&lines, "Outcome", shape.Outcome)
	appendSubsection(&lines, "Scope Boundary", shape.ScopeBoundary)
	appendSubsection(&lines, "Out of Scope", shape.OutOfScope)
	appendSubsection(&lines, "Success Signal", shape.SuccessSignal)
	return strings.Join(lines, "\n")
}

func renderEpicScopeSummary(scopeBoundary, outOfScope string) string {
	scopeBoundary = strings.TrimSpace(scopeBoundary)
	outOfScope = strings.TrimSpace(outOfScope)
	switch {
	case scopeBoundary == "" && outOfScope == "":
		return ""
	case outOfScope == "":
		return scopeBoundary
	case scopeBoundary == "":
		return "Not in scope:\n\n" + outOfScope
	default:
		return scopeBoundary + "\n\nNot in scope:\n\n" + outOfScope
	}
}
