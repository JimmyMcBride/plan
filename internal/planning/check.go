package planning

import (
	"fmt"
	"os"
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
	return r.ErrorCount() > 0
}

func (r *CheckReport) ErrorCount() int {
	count := 0
	for _, finding := range r.Findings {
		if finding.Severity == "error" {
			count++
		}
	}
	return count
}

func (r *CheckReport) WarningCount() int {
	count := 0
	for _, finding := range r.Findings {
		if finding.Severity == "warn" {
			count++
		}
	}
	return count
}

func (m *Manager) Check(input CheckInput) (*CheckReport, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	report := &CheckReport{Project: info.ProjectName}
	storyInfos, err := m.ListStories("", "")
	if err != nil {
		return nil, err
	}

	specs, err := m.specNotesForCheck(info, input)
	if err != nil {
		return nil, err
	}
	for _, spec := range specs {
		report.Findings = append(report.Findings, checkSpecNote(rel(info.ProjectDir, spec.Path), spec)...)
		report.Findings = append(report.Findings, checkSpecStoryReadiness(info, spec, storyInfos)...)
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
		story, err := m.ReadStory(input.StorySlug)
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
		note, err := m.ReadStory(story.Slug)
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

func checkSpecStoryReadiness(info *workspace.Info, spec *notes.Note, stories []StoryInfo) []CheckFinding {
	status := stringValue(spec.Metadata["status"])
	if status != "approved" && status != "implementing" {
		return nil
	}

	path := rel(info.ProjectDir, spec.Path)
	section := notes.ExtractSection(spec.Content, "Story Breakdown")
	parsed := parseStoryBreakdownCandidates(section)
	meaningful := meaningfulStoryBreakdownCandidates(parsed)
	var findings []CheckFinding

	if len(meaningful) == 0 {
		severity := "warn"
		if status == "implementing" {
			severity = "error"
		}
		findings = append(findings, CheckFinding{
			Severity:      severity,
			Rule:          "spec.story_breakdown_missing",
			ArtifactType:  "spec",
			ArtifactPath:  path,
			ArtifactTitle: spec.Title,
			Section:       "Story Breakdown",
			Message:       fmt.Sprintf("Spec is %q but ## Story Breakdown does not contain meaningful slice guidance.", status),
			Suggestion:    "Add concrete story slice entries under ## Story Breakdown or use plan story slice once the spec is ready.",
		})
		return findings
	}

	specSlug := slugFromPath(spec.Path)
	var childStories []StoryInfo
	for _, story := range stories {
		if story.Spec == specSlug {
			childStories = append(childStories, story)
		}
	}

	if status == "implementing" && len(childStories) == 0 {
		findings = append(findings, CheckFinding{
			Severity:      "error",
			Rule:          "spec.story_coverage_missing",
			ArtifactType:  "spec",
			ArtifactPath:  path,
			ArtifactTitle: spec.Title,
			Section:       "Story Breakdown",
			Message:       "Spec is \"implementing\" but no child stories are linked to it.",
			Suggestion:    "Create stories from the approved spec or correct the spec/story linkage before continuing implementation.",
		})
	}

	if len(childStories) > 0 && !storyBreakdownHasLinks(meaningful) {
		findings = append(findings, CheckFinding{
			Severity:      "warn",
			Rule:          "spec.story_breakdown_unlinked",
			ArtifactType:  "spec",
			ArtifactPath:  path,
			ArtifactTitle: spec.Title,
			Section:       "Story Breakdown",
			Message:       "Child stories exist, but ## Story Breakdown has not been refreshed with story links.",
			Suggestion:    "Refresh ## Story Breakdown with linked story entries so the spec-to-story handoff stays durable.",
		})
	}

	anyStarted := false
	for _, story := range childStories {
		if story.Status == "in_progress" || story.Status == "blocked" || story.Status == "done" {
			anyStarted = true
		}
	}

	if status == "implementing" && len(childStories) > 0 && !anyStarted {
		findings = append(findings, CheckFinding{
			Severity:      "warn",
			Rule:          "spec.story_status_mismatch",
			ArtifactType:  "spec",
			ArtifactPath:  path,
			ArtifactTitle: spec.Title,
			Section:       "Story Breakdown",
			Message:       "Spec is \"implementing\" but all linked stories are still todo.",
			Suggestion:    "Either start the relevant story work or move the spec back to approved until implementation begins.",
		})
	}

	expectedSlugs := map[string]struct{}{}
	for _, candidate := range meaningful {
		expectedSlugs[slugify(candidate.Title)] = struct{}{}
		if candidate.LinkTarget == "" {
			continue
		}
		if strings.HasPrefix(candidate.LinkTarget, "http://") || strings.HasPrefix(candidate.LinkTarget, "https://") {
			continue
		}
		linkedPath := filepath.Clean(filepath.Join(filepath.Dir(spec.Path), filepath.FromSlash(candidate.LinkTarget)))
		if _, err := os.Stat(linkedPath); err != nil {
			findings = append(findings, CheckFinding{
				Severity:      "error",
				Rule:          "spec.story_link_missing",
				ArtifactType:  "spec",
				ArtifactPath:  path,
				ArtifactTitle: spec.Title,
				Section:       "Story Breakdown",
				Message:       fmt.Sprintf("Story Breakdown links to %s, but that story file is missing.", filepath.ToSlash(linkedPath)),
				Suggestion:    "Remove the broken story link or recreate the missing story note.",
			})
		}
	}
	for _, story := range childStories {
		if _, ok := expectedSlugs[story.Slug]; ok {
			continue
		}
		findings = append(findings, CheckFinding{
			Severity:      "warn",
			Rule:          "spec.story_orphaned",
			ArtifactType:  "spec",
			ArtifactPath:  path,
			ArtifactTitle: spec.Title,
			Section:       "Story Breakdown",
			Message:       fmt.Sprintf("Story %s is linked to this spec but is not represented in ## Story Breakdown.", story.Path),
			Suggestion:    "Refresh ## Story Breakdown so it matches the stories currently linked to the spec.",
		})
	}

	return findings
}

func meaningfulStoryBreakdownCandidates(items []parsedStorySliceCandidate) []parsedStorySliceCandidate {
	var out []parsedStorySliceCandidate
	for _, item := range items {
		if item.Title == "" || strings.EqualFold(item.Title, seededStoryBreakdownPlaceholder) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func storyBreakdownHasLinks(items []parsedStorySliceCandidate) bool {
	for _, item := range items {
		if strings.TrimSpace(item.LinkTarget) != "" {
			return true
		}
	}
	return false
}

func normalizeSectionLine(line string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "- [ ] ")
	line = strings.TrimPrefix(line, "- [x] ")
	line = strings.TrimPrefix(line, "- ")
	line = strings.TrimPrefix(line, "* ")
	return strings.TrimSpace(line)
}
