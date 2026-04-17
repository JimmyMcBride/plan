package planning

import (
	"path/filepath"
	"slices"
	"strings"

	"plan/internal/notes"
)

const (
	analysisCategoryMissingConstraints = "Missing Constraints"
	analysisCategorySuccessGaps        = "Success Criteria Gaps"
	analysisCategoryHiddenDeps         = "Hidden Dependencies"
	analysisCategoryRiskGaps           = "Risk Gaps"
	analysisCategoryLeakage            = "What/Why vs How Leakage"
	analysisCategoryRecommended        = "Recommended Revisions"
)

var specAnalysisCategoryOrder = []string{
	analysisCategoryMissingConstraints,
	analysisCategorySuccessGaps,
	analysisCategoryHiddenDeps,
	analysisCategoryRiskGaps,
	analysisCategoryLeakage,
	analysisCategoryRecommended,
}

type SpecAnalysisFinding struct {
	Severity       string
	Category       string
	Message        string
	Recommendation string
}

type SpecAnalysisReport struct {
	Project  string
	SpecPath string
	Title    string
	Findings []SpecAnalysisFinding
}

func (r *SpecAnalysisReport) BlockingCount() int {
	count := 0
	for _, finding := range r.Findings {
		if finding.Severity == "error" {
			count++
		}
	}
	return count
}

func (r *SpecAnalysisReport) WarningCount() int {
	count := 0
	for _, finding := range r.Findings {
		if finding.Severity == "warn" {
			count++
		}
	}
	return count
}

func (r *SpecAnalysisReport) HasBlockingFindings() bool {
	return r.BlockingCount() > 0
}

func (r *SpecAnalysisReport) FindingsFor(category string) []SpecAnalysisFinding {
	var out []SpecAnalysisFinding
	for _, finding := range r.Findings {
		if finding.Category == category {
			out = append(out, finding)
		}
	}
	return out
}

func SpecAnalysisCategories() []string {
	return append([]string(nil), specAnalysisCategoryOrder...)
}

func (m *Manager) AnalyzeSpec(epicSlug string) (*SpecAnalysisReport, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(info.SpecsDir, m.specSlugForEpic(epicSlug)+".md")
	spec, err := notes.Read(path)
	if err != nil {
		return nil, err
	}

	report := &SpecAnalysisReport{
		Project:  info.ProjectName,
		SpecPath: rel(info.ProjectDir, path),
		Title:    spec.Title,
	}
	report.Findings = analyzeSpecFindings(report.SpecPath, spec)

	body := notes.SetSection(spec.Content, "Analysis", renderSpecAnalysis(report))
	updated, err := notes.Update(path, notes.UpdateInput{Body: &body})
	if err != nil {
		return nil, err
	}
	report.SpecPath = rel(info.ProjectDir, updated.Path)
	return report, nil
}

func analyzeSpecFindings(path string, spec *notes.Note) []SpecAnalysisFinding {
	var findings []SpecAnalysisFinding
	for _, finding := range checkSpecNote(path, spec) {
		category := structuralFindingCategory(finding.Rule)
		if category == "" {
			category = analysisCategoryRecommended
		}
		findings = append(findings, SpecAnalysisFinding{
			Severity:       finding.Severity,
			Category:       category,
			Message:        finding.Message,
			Recommendation: finding.Suggestion,
		})
	}

	addRiskGapFindings(spec, &findings)
	addDependencyGapFindings(spec, &findings)
	addNarrativeLeakageFindings(spec, &findings)
	addRecommendedRevisions(&findings)
	return dedupeAnalysisFindings(findings)
}

func structuralFindingCategory(rule string) string {
	switch {
	case strings.Contains(rule, "constraints"):
		return analysisCategoryMissingConstraints
	case strings.Contains(rule, "goals"), strings.Contains(rule, "verification"):
		return analysisCategorySuccessGaps
	default:
		return ""
	}
}

func addRiskGapFindings(spec *notes.Note, findings *[]SpecAnalysisFinding) {
	risks := notes.ExtractSection(spec.Content, "Risks / Open Questions")
	switch {
	case strings.TrimSpace(risks) == "":
		*findings = append(*findings, SpecAnalysisFinding{
			Severity:       "warn",
			Category:       analysisCategoryRiskGaps,
			Message:        "Risks / Open Questions is empty, so unresolved edges are not visible before implementation.",
			Recommendation: "Add concrete risks, open questions, or rollback concerns under ## Risks / Open Questions.",
		})
	case sectionLooksThin(risks):
		*findings = append(*findings, SpecAnalysisFinding{
			Severity:       "warn",
			Category:       analysisCategoryRiskGaps,
			Message:        "Risks / Open Questions is present but too thin to pressure-test the spec.",
			Recommendation: "Expand ## Risks / Open Questions with concrete failure modes, ambiguity, and boundary risks.",
		})
	}
}

func addDependencyGapFindings(spec *notes.Note, findings *[]SpecAnalysisFinding) {
	combined := strings.ToLower(strings.Join([]string{
		notes.ExtractSection(spec.Content, "Solution Shape"),
		notes.ExtractSection(spec.Content, "Flows"),
		notes.ExtractSection(spec.Content, "Data / Interfaces"),
		notes.ExtractSection(spec.Content, "Rollout"),
	}, "\n"))
	risks := strings.ToLower(notes.ExtractSection(spec.Content, "Risks / Open Questions"))

	type dependencyHeuristic struct {
		keywords       []string
		riskKeywords   []string
		message        string
		recommendation string
	}

	heuristics := []dependencyHeuristic{
		{
			keywords:       []string{"migration", "schema", "database", "table", "column", "sql", "postgres"},
			riskKeywords:   []string{"migration", "schema", "database", "data"},
			message:        "The spec mentions data-shape changes without calling out migration or data-safety risks.",
			recommendation: "Add migration, backfill, or rollback concerns to ## Risks / Open Questions and ## Rollout.",
		},
		{
			keywords:       []string{"api", "endpoint", "http", "webhook", "oauth", "third-party", "external"},
			riskKeywords:   []string{"api", "webhook", "external", "integration", "oauth"},
			message:        "The spec appears to depend on external interfaces, but the dependency risk is not named explicitly.",
			recommendation: "Call out integration contracts, failure handling, and ownership risks under ## Risks / Open Questions.",
		},
		{
			keywords:       []string{"queue", "worker", "background", "cron", "job"},
			riskKeywords:   []string{"queue", "worker", "background", "cron", "job"},
			message:        "The spec includes background or asynchronous work without surfacing operational risks.",
			recommendation: "Describe retry, observability, and failure handling for background work under ## Risks / Open Questions.",
		},
	}

	for _, heuristic := range heuristics {
		if !containsAny(combined, heuristic.keywords) || containsAny(risks, heuristic.riskKeywords) {
			continue
		}
		*findings = append(*findings, SpecAnalysisFinding{
			Severity:       "warn",
			Category:       analysisCategoryHiddenDeps,
			Message:        heuristic.message,
			Recommendation: heuristic.recommendation,
		})
	}
}

func addNarrativeLeakageFindings(spec *notes.Note, findings *[]SpecAnalysisFinding) {
	narrative := strings.ToLower(strings.Join([]string{
		notes.ExtractSection(spec.Content, "Why"),
		notes.ExtractSection(spec.Content, "Problem"),
		notes.ExtractSection(spec.Content, "Goals"),
		notes.ExtractSection(spec.Content, "Non-Goals"),
	}, "\n"))
	if !containsAny(narrative, []string{
		"api", "endpoint", "schema", "migration", "table", "column", "sql",
		"queue", "worker", "cron", "job", "cli", "flag", "yaml", "json", "cobra", "golang",
	}) {
		return
	}
	*findings = append(*findings, SpecAnalysisFinding{
		Severity:       "warn",
		Category:       analysisCategoryLeakage,
		Message:        "The narrative sections include implementation detail that belongs in Solution Shape or Data / Interfaces.",
		Recommendation: "Keep ## Why, ## Problem, ## Goals, and ## Non-Goals product-facing, then move technical detail into ## Solution Shape or ## Data / Interfaces.",
	})
}

func addRecommendedRevisions(findings *[]SpecAnalysisFinding) {
	var recommendations []string
	for _, finding := range *findings {
		if strings.TrimSpace(finding.Recommendation) == "" {
			continue
		}
		recommendations = append(recommendations, finding.Recommendation)
	}
	recommendations = uniqueStrings(recommendations)
	for _, recommendation := range recommendations {
		*findings = append(*findings, SpecAnalysisFinding{
			Severity:       "warn",
			Category:       analysisCategoryRecommended,
			Message:        recommendation,
			Recommendation: recommendation,
		})
	}
}

func dedupeAnalysisFindings(findings []SpecAnalysisFinding) []SpecAnalysisFinding {
	seen := map[string]struct{}{}
	var out []SpecAnalysisFinding
	for _, finding := range findings {
		key := finding.Severity + "|" + finding.Category + "|" + finding.Message + "|" + finding.Recommendation
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, finding)
	}
	return out
}

func renderSpecAnalysis(report *SpecAnalysisReport) string {
	var lines []string
	for _, category := range specAnalysisCategoryOrder {
		lines = append(lines, "### "+category, "")
		items := report.FindingsFor(category)
		if len(items) == 0 {
			lines = append(lines, "- None.")
		} else {
			for _, item := range items {
				lines = append(lines, "- ["+item.Severity+"] "+item.Message)
			}
		}
		lines = append(lines, "")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

func containsAny(body string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(body, needle) {
			return true
		}
	}
	return false
}

func uniqueStrings(items []string) []string {
	var out []string
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || slices.Contains(out, item) {
			continue
		}
		out = append(out, item)
	}
	return out
}
