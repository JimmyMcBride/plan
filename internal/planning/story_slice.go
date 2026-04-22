package planning

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"plan/internal/notes"
	"plan/internal/workspace"
)

const (
	seededExecutionPlanPlaceholder  = "Define execution slices when implementation begins"
	legacyStoryBreakdownPlaceholder = "Split approved spec into execution-ready stories"
)

type StorySliceCandidate struct {
	Title              string
	Slug               string
	Description        string
	AcceptanceCriteria []string
	Verification       []string
	StoryPath          string
	Status             string
}

type StorySlicePreview struct {
	Project    string
	EpicSlug   string
	SpecPath   string
	SpecTitle  string
	Candidates []StorySliceCandidate
}

type StorySliceApplyResult struct {
	Project      string
	EpicSlug     string
	SpecPath     string
	Candidates   []StorySliceCandidate
	CreatedPaths []string
	SkippedPaths []string
}

func (m *Manager) PreviewStorySlices(epicSlug string) (*StorySlicePreview, error) {
	info, epic, spec, err := m.loadEpicAndSpec(epicSlug)
	if err != nil {
		return nil, err
	}
	if status := stringValue(spec.Metadata["status"]); status != "approved" {
		if status == "" {
			status = "draft"
		}
		return nil, fmt.Errorf("spec %s is %q; approve the spec before slicing stories", rel(info.ProjectDir, spec.Path), status)
	}

	candidates, err := deriveStorySliceCandidates(info, epic, spec)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("spec %s has no slice-ready execution plan entries", rel(info.ProjectDir, spec.Path))
	}
	return &StorySlicePreview{
		Project:    info.ProjectName,
		EpicSlug:   slugFromPath(epic.Path),
		SpecPath:   rel(info.ProjectDir, spec.Path),
		SpecTitle:  spec.Title,
		Candidates: candidates,
	}, nil
}

func (m *Manager) ApplyStorySlices(epicSlug string) (*StorySliceApplyResult, error) {
	preview, err := m.PreviewStorySlices(epicSlug)
	if err != nil {
		return nil, err
	}
	info, _, spec, err := m.loadEpicAndSpec(epicSlug)
	if err != nil {
		return nil, err
	}

	result := &StorySliceApplyResult{
		Project:    preview.Project,
		EpicSlug:   preview.EpicSlug,
		SpecPath:   preview.SpecPath,
		Candidates: append([]StorySliceCandidate(nil), preview.Candidates...),
	}

	for i, candidate := range result.Candidates {
		if candidate.StoryPath != "" {
			result.SkippedPaths = append(result.SkippedPaths, candidate.StoryPath)
			continue
		}
		note, err := m.CreateStory(epicSlug, candidate.Title, candidate.Description, candidate.AcceptanceCriteria, candidate.Verification, nil)
		if err != nil {
			return nil, err
		}
		result.Candidates[i].StoryPath = note.Path
		result.Candidates[i].Status = "todo"
		result.CreatedPaths = append(result.CreatedPaths, note.Path)
	}

	breakdown := renderStoryBreakdownLinks(info.ProjectDir, spec.Path, result.Candidates)
	body := notes.SetSection(spec.Content, executionPlanHeading(spec.Content), breakdown)
	updatedSpec, err := notes.Update(spec.Path, notes.UpdateInput{Body: &body})
	if err != nil {
		return nil, err
	}
	result.SpecPath = rel(info.ProjectDir, updatedSpec.Path)
	return result, nil
}

func deriveStorySliceCandidates(info *workspace.Info, epic, spec *notes.Note) ([]StorySliceCandidate, error) {
	rawCandidates := parseStoryBreakdownCandidates(extractExecutionPlanSection(spec.Content))
	verificationDefaults := storySliceVerificationDefaults(spec)
	var candidates []StorySliceCandidate
	for _, raw := range rawCandidates {
		if raw.Title == "" || isSeededExecutionPlaceholder(raw.Title) {
			continue
		}
		candidate := StorySliceCandidate{
			Title:              raw.Title,
			Slug:               slugify(raw.Title),
			Description:        strings.TrimSpace(raw.Description),
			AcceptanceCriteria: trimmedItems(raw.AcceptanceCriteria),
			Verification:       trimmedItems(raw.Verification),
			Status:             "todo",
		}
		if candidate.Description == "" {
			candidate.Description = fmt.Sprintf("Deliver the %q slice described by the canonical spec.", candidate.Title)
		}
		if len(candidate.AcceptanceCriteria) == 0 {
			candidate.AcceptanceCriteria = []string{candidate.Title}
		}
		if len(candidate.Verification) == 0 {
			candidate.Verification = append([]string(nil), verificationDefaults...)
		}
		storyPath := filepath.Join(info.StoriesDir, candidate.Slug+".md")
		existing, err := notes.Read(storyPath)
		if err == nil {
			if stringValue(existing.Metadata["epic"]) != slugFromPath(epic.Path) || stringValue(existing.Metadata["spec"]) != slugFromPath(spec.Path) {
				return nil, fmt.Errorf("story slug collision for %s: existing note %s belongs to epic=%s spec=%s", candidate.Slug, rel(info.ProjectDir, existing.Path), stringValue(existing.Metadata["epic"]), stringValue(existing.Metadata["spec"]))
			}
			candidate.StoryPath = rel(info.ProjectDir, existing.Path)
			candidate.Status = stringValue(existing.Metadata["status"])
			if candidate.Status == "" {
				candidate.Status = "todo"
			}
		} else if !os.IsNotExist(err) {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}
	return candidates, nil
}

type parsedStorySliceCandidate struct {
	Title              string
	LinkTarget         string
	Description        string
	AcceptanceCriteria []string
	Verification       []string
}

func parseStoryBreakdownCandidates(section string) []parsedStorySliceCandidate {
	lines := strings.Split(strings.TrimSpace(section), "\n")
	var candidates []parsedStorySliceCandidate
	var current *parsedStorySliceCandidate
	for _, line := range lines {
		raw := strings.TrimRight(line, "\r")
		if strings.TrimSpace(raw) == "" {
			continue
		}
		if isTopLevelStoryBreakdownLine(raw) {
			title, linkTarget := extractStoryBreakdownTitle(strings.TrimSpace(raw))
			if title == "" {
				current = nil
				continue
			}
			candidate := parsedStorySliceCandidate{Title: title, LinkTarget: linkTarget}
			candidates = append(candidates, candidate)
			current = &candidates[len(candidates)-1]
			continue
		}
		if current == nil || !isIndentedStoryBreakdownLine(raw) {
			continue
		}
		trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(raw), "- "), "* "))
		lower := strings.ToLower(trimmed)
		switch {
		case strings.HasPrefix(lower, "desc:"), strings.HasPrefix(lower, "description:"):
			current.Description = strings.TrimSpace(afterColon(trimmed))
		case strings.HasPrefix(lower, "accept:"), strings.HasPrefix(lower, "criteria:"), strings.HasPrefix(lower, "criterion:"):
			if value := strings.TrimSpace(afterColon(trimmed)); value != "" {
				current.AcceptanceCriteria = append(current.AcceptanceCriteria, value)
			}
		case strings.HasPrefix(lower, "verify:"), strings.HasPrefix(lower, "verification:"):
			if value := strings.TrimSpace(afterColon(trimmed)); value != "" {
				current.Verification = append(current.Verification, value)
			}
		}
	}
	return candidates
}

func isTopLevelStoryBreakdownLine(line string) bool {
	if strings.TrimLeft(line, " \t") != line {
		return false
	}
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ")
}

func isIndentedStoryBreakdownLine(line string) bool {
	return strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t")
}

func extractStoryBreakdownTitle(line string) (string, string) {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "- [ ] ")
	line = strings.TrimPrefix(line, "- [x] ")
	line = strings.TrimPrefix(line, "- ")
	line = strings.TrimPrefix(line, "* ")
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "[") {
		if end := strings.Index(line, "]("); end > 1 {
			target := ""
			rest := line[end+2:]
			if close := strings.Index(rest, ")"); close >= 0 {
				target = strings.TrimSpace(rest[:close])
			}
			return strings.TrimSpace(line[1:end]), target
		}
	}
	return line, ""
}

func afterColon(value string) string {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

func storySliceVerificationDefaults(spec *notes.Note) []string {
	section := notes.ExtractSection(spec.Content, "Verification")
	var out []string
	for _, line := range strings.Split(section, "\n") {
		trimmed := normalizeSectionLine(line)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		return []string{"Validate the slice against the canonical spec."}
	}
	return out
}

func extractExecutionPlanSection(content string) string {
	if section := notes.ExtractSection(content, "Execution Plan"); strings.TrimSpace(section) != "" {
		return section
	}
	return notes.ExtractSection(content, "Story Breakdown")
}

func executionPlanHeading(content string) string {
	if section := notes.ExtractSection(content, "Execution Plan"); strings.TrimSpace(section) != "" {
		return "Execution Plan"
	}
	if section := notes.ExtractSection(content, "Story Breakdown"); strings.TrimSpace(section) != "" {
		return "Story Breakdown"
	}
	return "Execution Plan"
}

func isSeededExecutionPlaceholder(title string) bool {
	title = strings.TrimSpace(title)
	return strings.EqualFold(title, seededExecutionPlanPlaceholder) || strings.EqualFold(title, legacyStoryBreakdownPlaceholder)
}

func renderStoryBreakdownLinks(projectDir, specPath string, candidates []StorySliceCandidate) string {
	var lines []string
	for _, candidate := range candidates {
		if candidate.StoryPath == "" {
			continue
		}
		absStoryPath := filepath.Join(projectDir, filepath.FromSlash(candidate.StoryPath))
		check := "[ ]"
		if candidate.Status == "done" {
			check = "[x]"
		}
		lines = append(lines, fmt.Sprintf("- %s [%s](%s)", check, candidate.Title, relativeLinkPath(filepath.Dir(specPath), absStoryPath)))
	}
	return strings.Join(lines, "\n")
}

func (m *Manager) loadEpicAndSpec(epicSlug string) (*workspace.Info, *notes.Note, *notes.Note, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, nil, nil, err
	}
	epic, err := notes.Read(filepath.Join(info.EpicsDir, slugify(epicSlug)+".md"))
	if err != nil {
		return nil, nil, nil, err
	}
	spec, err := notes.Read(filepath.Join(info.SpecsDir, m.specSlugFromEpic(epic)+".md"))
	if err != nil {
		return nil, nil, nil, err
	}
	return info, epic, spec, nil
}
