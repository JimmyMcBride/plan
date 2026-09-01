package cmd

import (
	"fmt"
	"io"

	"plan/internal/planning"

	brainplanning "github.com/JimmyMcBride/brain/planning"
	brainapp "github.com/JimmyMcBride/brain/planning/application"
	"github.com/spf13/cobra"
)

func newCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check [project|epic|spec|story] [slug]",
		Short: "Run plan quality checks",
		Args:  cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			input, scopeLabel, err := resolveCheckScope(args)
			if err != nil {
				return err
			}
			if sharedInput, shared := resolveSharedCheckInput(args); shared {
				if handled, err := withSharedPlanning(cmd, "check", false, func(service *brainapp.Service) error {
					report, err := service.Check(cmd.Context(), sharedInput)
					if err != nil {
						return err
					}
					printSharedCheckReport(cmd.OutOrStdout(), report, scopeLabel)
					if report.HasErrors() {
						return fmt.Errorf("plan check found %d blocking issue(s)", report.ErrorCount())
					}
					return nil
				}); handled {
					return err
				}
			}
			report, err := planningManager().Check(input)
			if err != nil {
				return err
			}
			printCheckReport(cmd.OutOrStdout(), report, scopeLabel)
			if report.HasErrors() {
				return fmt.Errorf("plan check found %d blocking issue(s)", report.ErrorCount())
			}
			return nil
		},
	}
}

func resolveSharedCheckInput(args []string) (brainapp.CheckInput, bool) {
	if len(args) == 0 {
		return brainapp.CheckInput{}, true
	}
	if len(args) != 2 {
		return brainapp.CheckInput{}, false
	}
	switch args[0] {
	case "project":
		return brainapp.CheckInput{}, true
	case "spec":
		id := brainplanning.ArtifactID(args[1])
		return brainapp.CheckInput{SpecID: &id}, true
	default:
		return brainapp.CheckInput{}, false
	}
}

func resolveCheckScope(args []string) (planning.CheckInput, string, error) {
	if len(args) == 0 {
		return planning.CheckInput{}, "project", nil
	}
	if len(args) != 2 {
		return planning.CheckInput{}, "", fmt.Errorf("scope and slug must be provided together")
	}
	switch args[0] {
	case "project":
		return planning.CheckInput{}, "project", nil
	case "epic":
		return planning.CheckInput{EpicSlug: args[1]}, "epic:" + args[1], nil
	case "spec":
		return planning.CheckInput{SpecSlug: args[1]}, "spec:" + args[1], nil
	case "story":
		return planning.CheckInput{StorySlug: args[1]}, "story:" + args[1], nil
	default:
		return planning.CheckInput{}, "", fmt.Errorf("unsupported check scope %q", args[0])
	}
}

func printCheckReport(out io.Writer, report *planning.CheckReport, scopeLabel string) {
	fmt.Fprintf(out, "check_scope: %s\n", scopeLabel)
	fmt.Fprintf(out, "findings: %d total, %d blocking, %d guidance\n",
		len(report.Findings),
		report.ErrorCount(),
		report.WarningCount(),
	)
	if len(report.Findings) == 0 {
		fmt.Fprintln(out, "status: ok")
		return
	}
	for _, finding := range report.Findings {
		fmt.Fprintf(out, "- [%s] %s %s :: %s\n",
			finding.Severity,
			finding.ArtifactType,
			finding.ArtifactPath,
			finding.Section,
		)
		fmt.Fprintf(out, "  %s\n", finding.Message)
		if finding.Suggestion != "" {
			fmt.Fprintf(out, "  fix: %s\n", finding.Suggestion)
		}
	}
}

func printSharedCheckReport(out io.Writer, report brainapp.CheckReport, scopeLabel string) {
	fmt.Fprintf(out, "check_scope: %s\n", scopeLabel)
	fmt.Fprintf(out, "findings: %d total, %d blocking, %d guidance\n",
		len(report.Findings),
		report.ErrorCount(),
		report.WarningCount(),
	)
	if len(report.Findings) == 0 {
		fmt.Fprintln(out, "status: ok")
		return
	}
	for _, finding := range report.Findings {
		fmt.Fprintf(out, "- [%s] %s %s :: %s\n",
			finding.Severity,
			finding.ArtifactType,
			finding.ArtifactPath,
			finding.Section,
		)
		fmt.Fprintf(out, "  %s\n", finding.Message)
		if finding.Suggestion != "" {
			fmt.Fprintf(out, "  fix: %s\n", finding.Suggestion)
		}
	}
}
