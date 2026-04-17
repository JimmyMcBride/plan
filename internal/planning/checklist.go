package planning

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"plan/internal/notes"
)

const (
	checklistProfileGeneral        = "general"
	checklistProfileUIFlow         = "ui-flow"
	checklistProfileAPIIntegration = "api-integration"
	checklistProfileDataMigration  = "data-migration"
)

var specChecklistProfileOrder = []string{
	checklistProfileGeneral,
	checklistProfileUIFlow,
	checklistProfileAPIIntegration,
	checklistProfileDataMigration,
}

type SpecChecklistFinding struct {
	Severity       string
	Area           string
	Message        string
	Recommendation string
}

type SpecChecklistReport struct {
	Project  string
	SpecPath string
	Title    string
	Profile  string
	Findings []SpecChecklistFinding
}

func (r *SpecChecklistReport) BlockingCount() int {
	count := 0
	for _, finding := range r.Findings {
		if finding.Severity == "error" {
			count++
		}
	}
	return count
}

func (r *SpecChecklistReport) WarningCount() int {
	count := 0
	for _, finding := range r.Findings {
		if finding.Severity == "warn" {
			count++
		}
	}
	return count
}

func (r *SpecChecklistReport) HasBlockingFindings() bool {
	return r.BlockingCount() > 0
}

func SpecChecklistProfiles() []string {
	return append([]string(nil), specChecklistProfileOrder...)
}

func (m *Manager) RunSpecChecklist(epicSlug, profile string) (*SpecChecklistReport, error) {
	profile = strings.TrimSpace(profile)
	if !slices.Contains(specChecklistProfileOrder, profile) {
		return nil, fmt.Errorf("invalid checklist profile %q", profile)
	}
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(info.SpecsDir, m.specSlugForEpic(epicSlug)+".md")
	spec, err := notes.Read(path)
	if err != nil {
		return nil, err
	}

	report := &SpecChecklistReport{
		Project:  info.ProjectName,
		SpecPath: rel(info.ProjectDir, path),
		Title:    spec.Title,
		Profile:  profile,
	}
	report.Findings = buildSpecChecklistFindings(profile, spec)

	existing := notes.ExtractSection(spec.Content, "Checklist")
	sections := map[string]string{}
	for _, name := range specChecklistProfileOrder {
		if body := notes.ExtractSection(existing, name); strings.TrimSpace(body) != "" {
			sections[name] = body
		}
	}
	sections[profile] = renderSpecChecklistReport(report)
	body := notes.SetSection(spec.Content, "Checklist", renderNamedSubsections(specChecklistProfileOrder, sections))
	updated, err := notes.Update(path, notes.UpdateInput{Body: &body})
	if err != nil {
		return nil, err
	}
	report.SpecPath = rel(info.ProjectDir, updated.Path)
	return report, nil
}

func buildSpecChecklistFindings(profile string, spec *notes.Note) []SpecChecklistFinding {
	switch profile {
	case checklistProfileGeneral:
		return dedupeChecklistFindings(generalChecklistFindings(spec))
	case checklistProfileUIFlow:
		return dedupeChecklistFindings(uiFlowChecklistFindings(spec))
	case checklistProfileAPIIntegration:
		return dedupeChecklistFindings(apiIntegrationChecklistFindings(spec))
	case checklistProfileDataMigration:
		return dedupeChecklistFindings(dataMigrationChecklistFindings(spec))
	default:
		return nil
	}
}

func generalChecklistFindings(spec *notes.Note) []SpecChecklistFinding {
	var findings []SpecChecklistFinding
	if sectionLooksThin(notes.ExtractSection(spec.Content, "Goals")) {
		findings = append(findings, SpecChecklistFinding{
			Severity:       "warn",
			Area:           "Goals",
			Message:        "The general checklist expects more concrete outcome framing under ## Goals.",
			Recommendation: "Expand ## Goals with the clearest user-facing or execution-facing outcomes for this spec.",
		})
	}
	if sectionLooksThin(notes.ExtractSection(spec.Content, "Non-Goals")) {
		findings = append(findings, SpecChecklistFinding{
			Severity:       "warn",
			Area:           "Non-Goals",
			Message:        "The general checklist expects explicit non-goals so scope stays bounded.",
			Recommendation: "Add concrete exclusions under ## Non-Goals.",
		})
	}
	if sectionLooksThin(notes.ExtractSection(spec.Content, "Solution Shape")) {
		findings = append(findings, SpecChecklistFinding{
			Severity:       "warn",
			Area:           "Solution Shape",
			Message:        "The general checklist expects enough solution-shape detail to guide later story slicing.",
			Recommendation: "Expand ## Solution Shape with the intended approach and major boundary decisions.",
		})
	}
	verification := notes.ExtractSection(spec.Content, "Verification")
	if strings.TrimSpace(verification) == "" || sectionLooksThin(verification) {
		findings = append(findings, SpecChecklistFinding{
			Severity:       "error",
			Area:           "Verification",
			Message:        "The general checklist requires concrete verification so the spec can be executed safely.",
			Recommendation: "Add explicit checks, tests, or validation flows under ## Verification.",
		})
	}
	return findings
}

func uiFlowChecklistFindings(spec *notes.Note) []SpecChecklistFinding {
	var findings []SpecChecklistFinding
	flows := notes.ExtractSection(spec.Content, "Flows")
	if strings.TrimSpace(flows) == "" || sectionLooksThin(flows) {
		findings = append(findings, SpecChecklistFinding{
			Severity:       "error",
			Area:           "Flows",
			Message:        "The UI flow checklist requires a concrete user journey under ## Flows.",
			Recommendation: "Describe the primary UI journey step by step under ## Flows.",
		})
	}
	if !containsAny(strings.ToLower(flows), []string{"screen", "page", "modal", "form", "click", "select", "submit", "view"}) {
		findings = append(findings, SpecChecklistFinding{
			Severity:       "warn",
			Area:           "Flows",
			Message:        "The UI flow checklist expects user-facing interaction detail in ## Flows.",
			Recommendation: "Call out the user-visible screens, inputs, or transitions the flow depends on.",
		})
	}
	verification := strings.ToLower(notes.ExtractSection(spec.Content, "Verification"))
	if !containsAny(verification, []string{"manual", "screen", "ui", "click", "form", "visual", "journey"}) {
		findings = append(findings, SpecChecklistFinding{
			Severity:       "warn",
			Area:           "Verification",
			Message:        "The UI flow checklist expects at least one user-visible verification step.",
			Recommendation: "Add manual or UI-facing verification language under ## Verification.",
		})
	}
	if sectionLooksThin(notes.ExtractSection(spec.Content, "Non-Goals")) {
		findings = append(findings, SpecChecklistFinding{
			Severity:       "warn",
			Area:           "Non-Goals",
			Message:        "UI-heavy work benefits from explicit exclusions so the flow does not balloon.",
			Recommendation: "State which adjacent UI work is intentionally out of scope under ## Non-Goals.",
		})
	}
	return findings
}

func apiIntegrationChecklistFindings(spec *notes.Note) []SpecChecklistFinding {
	var findings []SpecChecklistFinding
	interfaces := strings.ToLower(notes.ExtractSection(spec.Content, "Data / Interfaces"))
	if !containsAny(interfaces, []string{"api", "endpoint", "http", "webhook", "oauth", "external", "integration", "contract"}) {
		findings = append(findings, SpecChecklistFinding{
			Severity:       "error",
			Area:           "Data / Interfaces",
			Message:        "The API integration checklist requires the external contract to be described under ## Data / Interfaces.",
			Recommendation: "Document the relevant endpoint, payload, auth, or contract details under ## Data / Interfaces.",
		})
	}
	risks := strings.ToLower(notes.ExtractSection(spec.Content, "Risks / Open Questions"))
	if !containsAny(risks, []string{"timeout", "retry", "failure", "auth", "ownership", "contract", "rate", "limit", "webhook"}) {
		findings = append(findings, SpecChecklistFinding{
			Severity:       "warn",
			Area:           "Risks / Open Questions",
			Message:        "The API integration checklist expects dependency and failure-handling risks to be called out.",
			Recommendation: "Add external-contract, auth, timeout, retry, or ownership risks under ## Risks / Open Questions.",
		})
	}
	rollout := strings.ToLower(notes.ExtractSection(spec.Content, "Rollout"))
	if !containsAny(rollout, []string{"flag", "fallback", "monitor", "observe", "rollback", "backward"}) {
		findings = append(findings, SpecChecklistFinding{
			Severity:       "warn",
			Area:           "Rollout",
			Message:        "The API integration checklist expects rollout safety for external dependency changes.",
			Recommendation: "Describe flags, fallback paths, monitoring, or backward-compatibility concerns under ## Rollout.",
		})
	}
	return findings
}

func dataMigrationChecklistFindings(spec *notes.Note) []SpecChecklistFinding {
	var findings []SpecChecklistFinding
	interfaces := strings.ToLower(notes.ExtractSection(spec.Content, "Data / Interfaces"))
	if !containsAny(interfaces, []string{"migration", "schema", "table", "column", "backfill", "data", "database"}) {
		findings = append(findings, SpecChecklistFinding{
			Severity:       "error",
			Area:           "Data / Interfaces",
			Message:        "The data migration checklist requires the schema or data-shape change to be described under ## Data / Interfaces.",
			Recommendation: "Document the migration, backfill, or data-shape changes under ## Data / Interfaces.",
		})
	}
	rollout := strings.ToLower(notes.ExtractSection(spec.Content, "Rollout"))
	if !containsAny(rollout, []string{"rollback", "backfill", "batch", "monitor", "flag", "rehearse"}) {
		findings = append(findings, SpecChecklistFinding{
			Severity:       "warn",
			Area:           "Rollout",
			Message:        "The data migration checklist expects rollback, backfill, or rollout-safety detail.",
			Recommendation: "Add rollout safety details such as rollback, batching, or backfill sequencing under ## Rollout.",
		})
	}
	verification := strings.ToLower(notes.ExtractSection(spec.Content, "Verification"))
	if !containsAny(verification, []string{"data", "migration", "backfill", "integrity", "row", "record", "schema"}) {
		findings = append(findings, SpecChecklistFinding{
			Severity:       "warn",
			Area:           "Verification",
			Message:        "The data migration checklist expects data validation beyond generic testing language.",
			Recommendation: "Describe how migrated data, backfills, or schema changes will be validated under ## Verification.",
		})
	}
	return findings
}

func renderSpecChecklistReport(report *SpecChecklistReport) string {
	var lines []string
	status := "ok"
	if report.HasBlockingFindings() {
		status = "action_needed"
	} else if len(report.Findings) > 0 {
		status = "guidance"
	}
	lines = append(lines,
		"status: "+status,
		fmt.Sprintf("blocking_findings: %d", report.BlockingCount()),
		fmt.Sprintf("guidance_findings: %d", report.WarningCount()),
	)
	if len(report.Findings) == 0 {
		lines = append(lines, "", "- [ok] No findings.")
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "")
	for _, finding := range report.Findings {
		lines = append(lines, fmt.Sprintf("- [%s] %s: %s", finding.Severity, finding.Area, finding.Message))
		if strings.TrimSpace(finding.Recommendation) != "" {
			lines = append(lines, "  fix: "+finding.Recommendation)
		}
	}
	return strings.Join(lines, "\n")
}

func dedupeChecklistFindings(findings []SpecChecklistFinding) []SpecChecklistFinding {
	seen := map[string]struct{}{}
	var out []SpecChecklistFinding
	for _, finding := range findings {
		key := finding.Severity + "|" + finding.Area + "|" + finding.Message + "|" + finding.Recommendation
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, finding)
	}
	return out
}
