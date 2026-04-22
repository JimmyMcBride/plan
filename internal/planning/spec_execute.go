package planning

import (
	"fmt"
	"path/filepath"
	"strings"

	"plan/internal/notes"
	"plan/internal/workspace"
)

type SpecExecutionSlice struct {
	Title        string
	Slug         string
	Goal         string
	Verification []string
}

type SpecExecutionPlan struct {
	Project         string
	SpecSlug        string
	SpecPath        string
	SpecTitle       string
	Status          string
	SuggestedBranch string
	Slices          []SpecExecutionSlice
}

func (m *Manager) PreviewSpecExecution(specSlug, branchPrefix string) (*SpecExecutionPlan, error) {
	info, spec, err := m.loadSpecBySlug(specSlug)
	if err != nil {
		return nil, err
	}
	status := stringValue(spec.Metadata["status"])
	if status == "" {
		status = "draft"
	}
	switch status {
	case "approved", "implementing":
	case "done":
		return nil, fmt.Errorf("spec %s is already %q; reopen it before starting a new execution pass", rel(info.ProjectDir, spec.Path), status)
	default:
		return nil, fmt.Errorf("spec %s is %q; approve the spec before starting execution", rel(info.ProjectDir, spec.Path), status)
	}
	return buildSpecExecutionPlan(info, spec, normalizeBranchPrefix(branchPrefix)), nil
}

func (m *Manager) BeginSpecExecution(specSlug, branchPrefix string) (*SpecExecutionPlan, error) {
	info, spec, err := m.loadSpecBySlug(specSlug)
	if err != nil {
		return nil, err
	}
	status := stringValue(spec.Metadata["status"])
	if status == "" {
		status = "draft"
	}
	switch status {
	case "approved":
		spec, err = notes.Update(spec.Path, notes.UpdateInput{
			Metadata: map[string]any{"status": "implementing"},
		})
		if err != nil {
			return nil, err
		}
	case "implementing":
	case "done":
		return nil, fmt.Errorf("spec %s is already %q; reopen it before starting a new execution pass", rel(info.ProjectDir, spec.Path), status)
	default:
		return nil, fmt.Errorf("spec %s is %q; approve the spec before starting execution", rel(info.ProjectDir, spec.Path), status)
	}
	return buildSpecExecutionPlan(info, spec, normalizeBranchPrefix(branchPrefix)), nil
}

func buildSpecExecutionPlan(info *workspace.Info, spec *notes.Note, branchPrefix string) *SpecExecutionPlan {
	status := stringValue(spec.Metadata["status"])
	if status == "" {
		status = "draft"
	}
	return &SpecExecutionPlan{
		Project:         info.ProjectName,
		SpecSlug:        slugFromPath(spec.Path),
		SpecPath:        rel(info.ProjectDir, spec.Path),
		SpecTitle:       spec.Title,
		Status:          status,
		SuggestedBranch: branchPrefix + slugFromPath(spec.Path),
		Slices:          deriveSpecExecutionSlices(spec),
	}
}

func deriveSpecExecutionSlices(spec *notes.Note) []SpecExecutionSlice {
	if candidates := parseStoryBreakdownCandidates(extractExecutionPlanSection(spec.Content)); len(candidates) > 0 {
		slices := make([]SpecExecutionSlice, 0, len(candidates))
		defaultVerification := storySliceVerificationDefaults(spec)
		for _, candidate := range candidates {
			if candidate.Title == "" || isSeededExecutionPlaceholder(candidate.Title) {
				continue
			}
			goal := strings.TrimSpace(candidate.Description)
			if goal == "" {
				if len(candidate.AcceptanceCriteria) > 0 {
					goal = strings.TrimSpace(candidate.AcceptanceCriteria[0])
				} else {
					goal = fmt.Sprintf("Complete the %q slice of %s.", candidate.Title, spec.Title)
				}
			}
			verification := trimmedItems(candidate.Verification)
			if len(verification) == 0 {
				verification = append([]string(nil), defaultVerification...)
			}
			slices = append(slices, SpecExecutionSlice{
				Title:        candidate.Title,
				Slug:         slugify(candidate.Title),
				Goal:         goal,
				Verification: verification,
			})
		}
		if len(slices) > 0 {
			return slices
		}
	}

	if flowSlices := deriveFlowExecutionSlices(spec); len(flowSlices) > 0 {
		return flowSlices
	}

	return deriveFallbackExecutionSlices(spec)
}

func deriveFlowExecutionSlices(spec *notes.Note) []SpecExecutionSlice {
	defaultVerification := storySliceVerificationDefaults(spec)
	var slices []SpecExecutionSlice
	for _, line := range strings.Split(notes.ExtractSection(spec.Content, "Flows"), "\n") {
		trimmed := normalizeSectionLine(line)
		if trimmed == "" {
			continue
		}
		slices = append(slices, SpecExecutionSlice{
			Title:        trimmed,
			Slug:         slugify(trimmed),
			Goal:         fmt.Sprintf("Implement and verify the %q flow for %s.", trimmed, spec.Title),
			Verification: append([]string(nil), defaultVerification...),
		})
		if len(slices) == 3 {
			break
		}
	}
	return slices
}

func deriveFallbackExecutionSlices(spec *notes.Note) []SpecExecutionSlice {
	baseTitle := strings.TrimSpace(strings.TrimSuffix(spec.Title, " Spec"))
	if baseTitle == "" {
		baseTitle = strings.TrimSpace(spec.Title)
	}
	defaultVerification := storySliceVerificationDefaults(spec)
	return []SpecExecutionSlice{
		{
			Title:        "Prepare " + baseTitle,
			Slug:         slugify("prepare " + baseTitle),
			Goal:         fmt.Sprintf("Set up the interfaces, constraints, and prerequisites required for %s.", baseTitle),
			Verification: append([]string(nil), defaultVerification...),
		},
		{
			Title:        "Implement " + baseTitle,
			Slug:         slugify("implement " + baseTitle),
			Goal:         fmt.Sprintf("Build the core behavior described by %s.", spec.Title),
			Verification: append([]string(nil), defaultVerification...),
		},
		{
			Title:        "Verify " + baseTitle,
			Slug:         slugify("verify " + baseTitle),
			Goal:         fmt.Sprintf("Run the end-to-end checks for %s and clean up rollout edges.", spec.Title),
			Verification: append([]string(nil), defaultVerification...),
		},
	}
}

func normalizeBranchPrefix(branchPrefix string) string {
	branchPrefix = strings.TrimSpace(branchPrefix)
	if branchPrefix == "" {
		branchPrefix = "feature/"
	}
	if !strings.HasSuffix(branchPrefix, "/") {
		branchPrefix += "/"
	}
	return branchPrefix
}

func (m *Manager) loadSpecBySlug(specSlug string) (*workspace.Info, *notes.Note, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, nil, err
	}
	spec, err := notes.Read(filepath.Join(info.SpecsDir, slugify(specSlug)+".md"))
	if err != nil {
		return nil, nil, err
	}
	return info, spec, nil
}
