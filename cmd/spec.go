package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

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
		Use:   "show <spec-slug>",
		Short: "Show a canonical spec",
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
		Use:   "edit <spec-slug>",
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
		Use:   "status <spec-slug>",
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
		Use:   "analyze <spec-slug>",
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

	var profile string
	checklist := &cobra.Command{
		Use:   "checklist <spec-slug>",
		Short: "Run a profile-driven checklist pass against a spec",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := planningManager().RunSpecChecklist(args[0], profile)
			if err != nil {
				return err
			}
			printSpecChecklist(cmd.OutOrStdout(), report)
			if report.HasBlockingFindings() {
				return fmt.Errorf("spec checklist found %d blocking issue(s)", report.BlockingCount())
			}
			return nil
		},
	}
	checklist.Flags().StringVar(&profile, "profile", "general", "checklist profile: general, ui-flow, api-integration, data-migration")

	var initiativeSlug string
	var initiativeTitle string
	var initiativeSummary string
	var clearInitiative bool
	initiative := &cobra.Command{
		Use:   "initiative <spec-slug>",
		Short: "Set or clear lightweight initiative metadata on a spec",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !clearInitiative && strings.TrimSpace(initiativeSlug) == "" {
				return fmt.Errorf("initiative --set value is required unless --clear is used")
			}
			var (
				note *notes.Note
				err  error
			)
			if clearInitiative {
				note, err = planningManager().ClearSpecInitiative(args[0])
			} else {
				note, err = planningManager().SetSpecInitiative(args[0], planning.InitiativeRef{
					Slug:    initiativeSlug,
					Title:   initiativeTitle,
					Summary: initiativeSummary,
				})
			}
			if err != nil {
				return err
			}
			fmt.Printf("Updated initiative metadata for %s\n", note.Path)
			return nil
		},
	}
	initiative.Flags().StringVar(&initiativeSlug, "set", "", "initiative slug or shared reference")
	initiative.Flags().StringVar(&initiativeTitle, "title", "", "human title for the initiative")
	initiative.Flags().StringVar(&initiativeSummary, "summary", "", "optional initiative summary or goal")
	initiative.Flags().BoolVar(&clearInitiative, "clear", false, "clear initiative metadata from the spec")

	var executeBranchPrefix string
	execute := &cobra.Command{
		Use:   "execute <spec-slug>",
		Short: "Start spec execution with a suggested branch and ephemeral slices",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			plan, err := planningManager().BeginSpecExecution(args[0], executeBranchPrefix)
			if err != nil {
				return err
			}
			printSpecExecutionPlan(cmd.OutOrStdout(), plan)
			return nil
		},
	}
	execute.Flags().StringVar(&executeBranchPrefix, "branch-prefix", "feature/", "prefix for the suggested execution branch")

	var handoffBranchPrefix string
	handoff := &cobra.Command{
		Use:   "handoff <spec-slug>",
		Short: "Continue a guided spec into the execution stage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			session, err := planningManager().ReadGuidedSessionBySpec(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Spec recap:\nCurrent understanding: %s\nRecommended next stage: continue into execution.\n", session.Summary)
			preview, err := planningManager().PreviewSpecExecution(args[0], handoffBranchPrefix)
			if err != nil {
				return err
			}
			printSpecExecutionPlan(cmd.OutOrStdout(), preview)
			ok, err := confirmSpecExecutionStart(bufio.NewReader(cmd.InOrStdin()), cmd.OutOrStdout())
			if err != nil {
				return err
			}
			if !ok {
				updated, err := planningManager().UpdateGuidedSession(session.ChainID, planning.GuidedSessionUpdateInput{
					CurrentStage: "spec",
					StageStatus:  "in_progress",
					NextAction:   "Spec execution handoff is ready when you want to start implementation.",
				})
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Checkpoint saved for %s\nNext: %s\n", updated.ChainID, updated.NextAction)
				return nil
			}
			plan, err := planningManager().BeginSpecExecution(args[0], handoffBranchPrefix)
			if err != nil {
				return err
			}
			updated, err := planningManager().AdvanceGuidedSessionToExecutionBySpec(args[0])
			if err != nil {
				return err
			}
			printSpecExecutionPlan(cmd.OutOrStdout(), plan)
			fmt.Fprintf(cmd.OutOrStdout(), "Next: %s\n", updated.NextAction)
			return nil
		},
	}
	handoff.Flags().StringVar(&handoffBranchPrefix, "branch-prefix", "feature/", "prefix for the suggested execution branch")

	cmd.AddCommand(show, edit, statusCmd, analyze, checklist, initiative, execute, handoff)
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

func printSpecChecklist(out io.Writer, report *planning.SpecChecklistReport) {
	fmt.Fprintf(out, "spec_checklist: %s\n", report.SpecPath)
	fmt.Fprintf(out, "profile: %s\n", report.Profile)
	fmt.Fprintf(out, "findings: %d total, %d blocking, %d guidance\n",
		len(report.Findings),
		report.BlockingCount(),
		report.WarningCount(),
	)
	if len(report.Findings) == 0 {
		fmt.Fprintln(out, "status: ok")
		return
	}
	for _, item := range report.Findings {
		fmt.Fprintf(out, "- [%s] %s: %s\n", item.Severity, item.Area, item.Message)
		if item.Recommendation != "" {
			fmt.Fprintf(out, "  fix: %s\n", item.Recommendation)
		}
	}
}

func printSpecExecutionPlan(out io.Writer, plan *planning.SpecExecutionPlan) {
	fmt.Fprintf(out, "spec_execution: %s\n", plan.SpecPath)
	fmt.Fprintf(out, "status: %s\n", plan.Status)
	fmt.Fprintf(out, "branch: %s\n", plan.SuggestedBranch)
	fmt.Fprintf(out, "slices: %d\n", len(plan.Slices))
	for index, slice := range plan.Slices {
		fmt.Fprintf(out, "%d. %s\n", index+1, slice.Title)
		fmt.Fprintf(out, "   goal: %s\n", slice.Goal)
		for _, verify := range slice.Verification {
			fmt.Fprintf(out, "   verify: %s\n", verify)
		}
	}
	fmt.Fprintln(out, "workflow:")
	fmt.Fprintln(out, "- implement one slice at a time")
	fmt.Fprintln(out, "- review and verify each slice before committing it")
	fmt.Fprintln(out, "- open a PR after the full spec is built")
}

func confirmSpecExecutionStart(reader *bufio.Reader, out io.Writer) (bool, error) {
	fmt.Fprint(out, "Start execution from this spec plan? [y/N]: ")
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}
