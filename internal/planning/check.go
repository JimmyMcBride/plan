package planning

import (
	"fmt"
	"path/filepath"
	"strings"

	"plan/internal/notes"
	"plan/internal/workspace"
)

type CheckInput struct {
	EpicSlug  string
	SpecSlug  string
	StorySlug string
}

type CheckReport struct {
	Project  string
	Findings []CheckFinding
}

type CheckFinding struct {
	Severity      string
	Rule          string
	ArtifactType  string
	ArtifactPath  string
	ArtifactTitle string
	Section       string
	Message       string
	Suggestion    string
}

func (r *CheckReport) HasErrors() bool {
	for _, finding := range r.Findings {
		if finding.Severity == "error" {
			return true
		}
	}
	return false
}

func (m *Manager) Check(input CheckInput) (*CheckReport, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	report := &CheckReport{Project: info.ProjectName}

	specs, err := m.specNotesForCheck(info, input)
	if err != nil {
		return nil, err
	}
	for _, spec := range specs {
		report.Findings = append(report.Findings, checkSpecNote(rel(info.ProjectDir, spec.Path), spec)...)
	}
	stories, err := m.storyNotesForCheck(info, input)
	if err != nil {
		return nil, err
	}
	for _, story := range stories {
		report.Findings = append(report.Findings, checkStoryNote(rel(info.ProjectDir, story.Path), story)...)
	}

	return report, nil
}

func (m *Manager) specNotesForCheck(info *workspace.Info, input CheckInput) ([]*notes.Note, error) {
	switch {
	case strings.TrimSpace(input.StorySlug) != "":
		return nil, nil
	case strings.TrimSpace(input.SpecSlug) != "":
		spec, err := notes.Read(filepath.Join(info.SpecsDir, slugify(input.SpecSlug)+".md"))
		if err != nil {
			return nil, err
		}
		return []*notes.Note{spec}, nil
	case strings.TrimSpace(input.EpicSlug) != "":
		spec, err := notes.Read(filepath.Join(info.SpecsDir, m.specSlugForEpic(input.EpicSlug)+".md"))
		if err != nil {
			return nil, err
		}
		return []*notes.Note{spec}, nil
	default:
		return readNotesInDir(info.SpecsDir)
	}
}

func (m *Manager) storyNotesForCheck(info *workspace.Info, input CheckInput) ([]*notes.Note, error) {
	switch {
	case strings.TrimSpace(input.StorySlug) != "":
		story, err := notes.Read(filepath.Join(info.StoriesDir, slugify(input.StorySlug)+".md"))
		if err != nil {
			return nil, err
		}
		return []*notes.Note{story}, nil
	case strings.TrimSpace(input.SpecSlug) != "":
		return nil, nil
	case strings.TrimSpace(input.EpicSlug) != "":
		return m.readStoriesByFilter(info, func(story StoryInfo) bool {
			return story.Epic == slugify(input.EpicSlug)
		})
	default:
		return readNotesInDir(info.StoriesDir)
	}
}

func (m *Manager) readStoriesByFilter(info *workspace.Info, keep func(StoryInfo) bool) ([]*notes.Note, error) {
	stories, err := m.ListStories("", "")
	if err != nil {
		return nil, err
	}
	var out []*notes.Note
	for _, story := range stories {
		if !keep(story) {
			continue
		}
		note, err := notes.Read(filepath.Join(info.ProjectDir, filepath.FromSlash(story.Path)))
		if err != nil {
			return nil, err
		}
		out = append(out, note)
	}
	return out, nil
}

type specSectionRule struct {
	Heading    string
	Key        string
	Suggestion string
}

var requiredSpecSectionRules = []specSectionRule{
	{
		Heading:    "Problem",
		Key:        "problem",
		Suggestion: "Add a concrete problem statement under ## Problem that explains what is broken or missing today.",
	},
	{
		Heading:    "Goals",
		Key:        "goals",
		Suggestion: "Expand ## Goals with the specific outcomes this spec must deliver.",
	},
	{
		Heading:    "Non-Goals",
		Key:        "non_goals",
		Suggestion: "Use ## Non-Goals to define what this work will explicitly not do.",
	},
	{
		Heading:    "Constraints",
		Key:        "constraints",
		Suggestion: "List the design or implementation limits under ## Constraints so tradeoffs stay clear.",
	},
	{
		Heading:    "Verification",
		Key:        "verification",
		Suggestion: "Describe how this spec will be validated under ## Verification with explicit checks or test flows.",
	},
}

func checkSpecNote(path string, spec *notes.Note) []CheckFinding {
	var findings []CheckFinding
	for _, rule := range requiredSpecSectionRules {
		section := notes.ExtractSection(spec.Content, rule.Heading)
		switch {
		case strings.TrimSpace(section) == "":
			findings = append(findings, CheckFinding{
				Severity:      "error",
				Rule:          fmt.Sprintf("spec.missing_%s", rule.Key),
				ArtifactType:  "spec",
				ArtifactPath:  path,
				ArtifactTitle: spec.Title,
				Section:       rule.Heading,
				Message:       fmt.Sprintf("Missing required ## %s section content.", rule.Heading),
				Suggestion:    rule.Suggestion,
			})
		case sectionLooksThin(section):
			findings = append(findings, CheckFinding{
				Severity:      "warn",
				Rule:          fmt.Sprintf("spec.thin_%s", rule.Key),
				ArtifactType:  "spec",
				ArtifactPath:  path,
				ArtifactTitle: spec.Title,
				Section:       rule.Heading,
				Message:       fmt.Sprintf("## %s is present but too thin to guide execution.", rule.Heading),
				Suggestion:    rule.Suggestion,
			})
		}
	}
	return findings
}

var requiredStorySectionRules = []specSectionRule{
	{
		Heading:    "Description",
		Key:        "description",
		Suggestion: "Describe the concrete implementation slice under ## Description so execution starts from a clear brief.",
	},
	{
		Heading:    "Acceptance Criteria",
		Key:        "acceptance_criteria",
		Suggestion: "List the expected outcomes under ## Acceptance Criteria so the story has a clear finish line.",
	},
	{
		Heading:    "Verification",
		Key:        "verification",
		Suggestion: "Add explicit checks under ## Verification so the story can be validated after implementation.",
	},
}

func checkStoryNote(path string, story *notes.Note) []CheckFinding {
	status := stringValue(story.Metadata["status"])
	executionReady := storyBodyHasExecutionExpectations(story.Content)
	var findings []CheckFinding
	for _, rule := range requiredStorySectionRules {
		section := notes.ExtractSection(story.Content, rule.Heading)
		switch {
		case strings.TrimSpace(section) == "":
			message := fmt.Sprintf("Missing required ## %s section content.", rule.Heading)
			if rule.Heading != "Description" && requiresExecutionExpectations(status) {
				message = fmt.Sprintf("Story is %q but missing ## %s content required by the execution lifecycle.", status, rule.Heading)
			}
			findings = append(findings, CheckFinding{
				Severity:      "error",
				Rule:          fmt.Sprintf("story.missing_%s", rule.Key),
				ArtifactType:  "story",
				ArtifactPath:  path,
				ArtifactTitle: story.Title,
				Section:       rule.Heading,
				Message:       message,
				Suggestion:    rule.Suggestion,
			})
		case rule.Heading == "Description" && sectionLooksThin(section):
			findings = append(findings, CheckFinding{
				Severity:      "warn",
				Rule:          "story.thin_description",
				ArtifactType:  "story",
				ArtifactPath:  path,
				ArtifactTitle: story.Title,
				Section:       rule.Heading,
				Message:       "## Description is present but too thin to guide execution.",
				Suggestion:    rule.Suggestion,
			})
		}
	}
	if requiresExecutionExpectations(status) && !executionReady {
		findings = append(findings, CheckFinding{
			Severity:      "error",
			Rule:          "story.execution_expectations",
			ArtifactType:  "story",
			ArtifactPath:  path,
			ArtifactTitle: story.Title,
			Section:       "Acceptance Criteria / Verification",
			Message:       fmt.Sprintf("Story is %q but does not satisfy the acceptance-and-verification requirements enforced by the story lifecycle.", status),
			Suggestion:    "Restore both ## Acceptance Criteria and ## Verification before keeping the story in progress or done.",
		})
	}
	return findings
}

func sectionLooksThin(content string) bool {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	meaningfulLines := 0
	totalWords := 0
	for _, line := range lines {
		trimmed := normalizeSectionLine(line)
		if trimmed == "" {
			continue
		}
		meaningfulLines++
		totalWords += len(strings.Fields(trimmed))
	}
	if meaningfulLines == 0 {
		return true
	}
	if meaningfulLines >= 2 && totalWords >= 6 {
		return false
	}
	return totalWords < 6
}

func normalizeSectionLine(line string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "- [ ] ")
	line = strings.TrimPrefix(line, "- [x] ")
	line = strings.TrimPrefix(line, "- ")
	line = strings.TrimPrefix(line, "* ")
	return strings.TrimSpace(line)
}
