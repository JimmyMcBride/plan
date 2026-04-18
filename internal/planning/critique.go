package planning

import (
	"fmt"
	"path/filepath"
	"strings"

	"plan/internal/notes"
)

type StoryCritique struct {
	Path                  string
	ScopeFit              string
	VerticalSliceCheck    string
	HiddenPrerequisites   string
	VerificationGaps      string
	RewriteRecommendation string
}

type StoryCritiqueInput struct {
	ScopeFit              string
	VerticalSliceCheck    string
	HiddenPrerequisites   string
	VerificationGaps      string
	RewriteRecommendation string
}

func (c StoryCritique) HasGaps() bool {
	return strings.TrimSpace(c.ScopeFit) == "" ||
		strings.TrimSpace(c.VerticalSliceCheck) == "" ||
		strings.TrimSpace(c.HiddenPrerequisites) == "" ||
		strings.TrimSpace(c.VerificationGaps) == "" ||
		strings.TrimSpace(c.RewriteRecommendation) == ""
}

func (c StoryCritique) RecommendationAction() string {
	value := strings.ToLower(strings.TrimSpace(c.RewriteRecommendation))
	switch value {
	case "keep", "rewrite", "reslice":
		return value
	default:
		return ""
	}
}

func (c StoryCritique) HasBlockingRecommendation() bool {
	switch c.RecommendationAction() {
	case "rewrite", "reslice":
		return true
	default:
		return false
	}
}

func (m *Manager) ReadStoryCritique(storySlug string) (*StoryCritique, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	note, err := notes.Read(filepath.Join(info.StoriesDir, slugify(storySlug)+".md"))
	if err != nil {
		return nil, err
	}
	if note.Type != "story" {
		return nil, fmt.Errorf("%s is not a story note", note.Path)
	}
	state := extractStoryCritique(note)
	state.Path = rel(info.ProjectDir, note.Path)
	return &state, nil
}

func (m *Manager) UpdateStoryCritique(storySlug string, input StoryCritiqueInput) (*notes.Note, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(info.StoriesDir, slugify(storySlug)+".md")
	note, err := notes.Read(path)
	if err != nil {
		return nil, err
	}
	if note.Type != "story" {
		return nil, fmt.Errorf("%s is not a story note", note.Path)
	}

	current := extractStoryCritique(note)
	if strings.TrimSpace(input.ScopeFit) != "" {
		current.ScopeFit = strings.TrimSpace(input.ScopeFit)
	}
	if strings.TrimSpace(input.VerticalSliceCheck) != "" {
		current.VerticalSliceCheck = strings.TrimSpace(input.VerticalSliceCheck)
	}
	if strings.TrimSpace(input.HiddenPrerequisites) != "" {
		current.HiddenPrerequisites = normalizeBulletList(input.HiddenPrerequisites)
	}
	if strings.TrimSpace(input.VerificationGaps) != "" {
		current.VerificationGaps = normalizeBulletList(input.VerificationGaps)
	}
	if strings.TrimSpace(input.RewriteRecommendation) != "" {
		current.RewriteRecommendation = strings.ToLower(strings.TrimSpace(input.RewriteRecommendation))
	}
	if current.RecommendationAction() == "" && strings.TrimSpace(current.RewriteRecommendation) != "" {
		return nil, fmt.Errorf("invalid rewrite recommendation %q", current.RewriteRecommendation)
	}

	body := notes.SetSection(note.Content, "Critique", renderStoryCritique(current))
	updated, err := notes.Update(path, notes.UpdateInput{Body: &body})
	if err != nil {
		return nil, err
	}
	return m.relNote(updated, info.ProjectDir), nil
}

func extractStoryCritique(note *notes.Note) StoryCritique {
	return StoryCritique{
		ScopeFit:              extractSubsection(note.Content, "Critique", "Scope Fit"),
		VerticalSliceCheck:    extractSubsection(note.Content, "Critique", "Vertical Slice Check"),
		HiddenPrerequisites:   extractSubsection(note.Content, "Critique", "Hidden Prerequisites"),
		VerificationGaps:      extractSubsection(note.Content, "Critique", "Verification Gaps"),
		RewriteRecommendation: extractSubsection(note.Content, "Critique", "Rewrite Recommendation"),
	}
}

func renderStoryCritique(critique StoryCritique) string {
	var lines []string
	appendSubsection(&lines, "Scope Fit", critique.ScopeFit)
	appendSubsection(&lines, "Vertical Slice Check", critique.VerticalSliceCheck)
	appendSubsection(&lines, "Hidden Prerequisites", critique.HiddenPrerequisites)
	appendSubsection(&lines, "Verification Gaps", critique.VerificationGaps)
	appendSubsection(&lines, "Rewrite Recommendation", critique.RewriteRecommendation)
	return strings.Join(lines, "\n")
}
