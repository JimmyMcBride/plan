package cmd

import (
	"fmt"
	"io"

	"plan/internal/notes"
	"plan/internal/planning"

	"github.com/spf13/cobra"
)

func newSpecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spec",
		Short: "Manage specs",
	}

	show := &cobra.Command{
		Use:   "show <epic-slug>",
		Short: "Show the canonical spec for an epic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := planningManager().ReadSpec(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("%s\n\n%s", note.Path, note.Content)
			return nil
		},
	}

	var body string
	var useStdin bool
	var editor string
	edit := &cobra.Command{
		Use:   "edit <epic-slug>",
		Short: "Edit a spec via --body, --stdin, or $EDITOR",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := planningManager().ReadSpec(args[0])
			if err != nil {
				return err
			}
			updatedBody, err := readBody(cmd.InOrStdin(), body, useStdin)
			if err != nil {
				return err
			}
			if updatedBody == "" && !useStdin {
				updatedBody, err = editTextInEditor(note.Content, editor)
				if err != nil {
					return err
				}
			}
			updated, err := planningManager().UpdateSpec(args[0], notes.UpdateInput{
				Body: &updatedBody,
			})
			if err != nil {
				return err
			}
			fmt.Printf("Updated spec %s\n", updated.Path)
			return nil
		},
	}
	edit.Flags().StringVarP(&body, "body", "b", "", "replacement body")
	edit.Flags().BoolVar(&useStdin, "stdin", false, "read replacement body from stdin")
	edit.Flags().StringVar(&editor, "editor", "", "editor command")

	var status string
	statusCmd := &cobra.Command{
		Use:   "status <epic-slug>",
		Short: "Set spec status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			updated, err := planningManager().SetSpecStatus(args[0], status)
			if err != nil {
				return err
			}
			fmt.Printf("Set spec %s to %s\n", updated.Path, status)
			return nil
		},
	}
	statusCmd.Flags().StringVar(&status, "set", "", "new status: draft, approved, implementing, done")
	_ = statusCmd.MarkFlagRequired("set")

	analyze := &cobra.Command{
		Use:   "analyze <epic-slug>",
		Short: "Analyze a spec for refinement gaps without rewriting its canonical sections",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := planningManager().AnalyzeSpec(args[0])
			if err != nil {
				return err
			}
			printSpecAnalysis(cmd.OutOrStdout(), report)
			if report.HasBlockingFindings() {
				return fmt.Errorf("spec analysis found %d blocking issue(s)", report.BlockingCount())
			}
			return nil
		},
	}

	cmd.AddCommand(show, edit, statusCmd, analyze)
	return cmd
}

func printSpecAnalysis(out io.Writer, report *planning.SpecAnalysisReport) {
	fmt.Fprintf(out, "spec_analysis: %s\n", report.SpecPath)
	fmt.Fprintf(out, "findings: %d total, %d blocking, %d guidance\n",
		len(report.Findings),
		report.BlockingCount(),
		report.WarningCount(),
	)
	if len(report.Findings) == 0 {
		fmt.Fprintln(out, "status: ok")
		return
	}
	for _, category := range planning.SpecAnalysisCategories() {
		items := report.FindingsFor(category)
		if len(items) == 0 {
			continue
		}
		fmt.Fprintf(out, "%s:\n", category)
		for _, item := range items {
			fmt.Fprintf(out, "- [%s] %s\n", item.Severity, item.Message)
			if item.Recommendation != "" {
				fmt.Fprintf(out, "  fix: %s\n", item.Recommendation)
			}
		}
	}
}
