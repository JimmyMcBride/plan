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

	return report, nil
}

func (m *Manager) specNotesForCheck(info *workspace.Info, input CheckInput) ([]*notes.Note, error) {
	switch {
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
