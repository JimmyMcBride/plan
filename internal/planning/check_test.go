package planning

import (
	"strings"
	"testing"

	"plan/internal/notes"
	"plan/internal/workspace"
)

func TestCheckSpecFindsMissingRequiredSections(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}
	body := strings.Join([]string{
		"# Billing Spec",
		"",
		"Created: now",
		"",
		"## Why",
		"",
		"Keep billing reliable.",
		"",
		"## Problem",
		"",
		"## Goals",
		"",
		"## Non-Goals",
		"",
		"## Constraints",
		"",
		"## Verification",
		"",
	}, "\n")
	if _, err := manager.UpdateSpec("billing", notes.UpdateInput{Body: &body}); err != nil {
		t.Fatal(err)
	}

	report, err := manager.Check(CheckInput{SpecSlug: "billing"})
	if err != nil {
		t.Fatal(err)
	}
	if !report.HasErrors() {
		t.Fatalf("expected missing sections to produce blocking findings: %+v", report.Findings)
	}
	assertHasFinding(t, report.Findings, "spec.missing_problem", "Problem")
	assertHasFinding(t, report.Findings, "spec.missing_goals", "Goals")
	assertHasFinding(t, report.Findings, "spec.missing_non_goals", "Non-Goals")
	assertHasFinding(t, report.Findings, "spec.missing_constraints", "Constraints")
	assertHasFinding(t, report.Findings, "spec.missing_verification", "Verification")
}

func TestCheckSpecFlagsThinRequiredSections(t *testing.T) {
	root := t.TempDir()
	ws := workspace.New(root)
	if _, err := ws.Init(); err != nil {
		t.Fatal(err)
	}
	manager := New(ws)

	if _, err := manager.CreateEpic("Billing", ""); err != nil {
		t.Fatal(err)
	}
	body := strings.Join([]string{
		"# Billing Spec",
		"",
		"Created: now",
		"",
		"## Why",
		"",
		"Keep billing reliable.",
		"",
		"## Problem",
		"",
		"Confusing invoices.",
		"",
		"## Goals",
		"",
		"- clarity",
		"",
		"## Non-Goals",
		"",
		"Not taxes.",
		"",
		"## Constraints",
		"",
		"Local only.",
		"",
		"## Verification",
		"",
		"Run tests.",
		"",
	}, "\n")
	if _, err := manager.UpdateSpec("billing", notes.UpdateInput{Body: &body}); err != nil {
		t.Fatal(err)
	}

	report, err := manager.Check(CheckInput{SpecSlug: "billing"})
	if err != nil {
		t.Fatal(err)
	}
	if report.HasErrors() {
		t.Fatalf("expected thin sections to warn without blocking: %+v", report.Findings)
	}
	assertHasFinding(t, report.Findings, "spec.thin_problem", "Problem")
	assertHasFinding(t, report.Findings, "spec.thin_goals", "Goals")
	assertHasFinding(t, report.Findings, "spec.thin_non_goals", "Non-Goals")
	assertHasFinding(t, report.Findings, "spec.thin_constraints", "Constraints")
	assertHasFinding(t, report.Findings, "spec.thin_verification", "Verification")
}

func assertHasFinding(t *testing.T, findings []CheckFinding, rule, section string) {
	t.Helper()
	for _, finding := range findings {
		if finding.Rule == rule && finding.Section == section && finding.Suggestion != "" {
			return
		}
	}
	t.Fatalf("expected finding %s for %s: %+v", rule, section, findings)
}
