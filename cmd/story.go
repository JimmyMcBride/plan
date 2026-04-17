package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"plan/internal/planning"

	"github.com/spf13/cobra"
)

func newStoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "story",
		Short: "Manage stories",
	}

	var body string
	var criteria []string
	var verification []string
	var resources []string
	create := &cobra.Command{
		Use:   "create <epic-slug> <title>",
		Short: "Create a story from an approved spec",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := planningManager().CreateStory(args[0], args[1], body, criteria, verification, resources)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created story %s\n", note.Path)
			return nil
		},
	}
	create.Flags().StringVarP(&body, "body", "b", "", "story description")
	create.Flags().StringArrayVar(&criteria, "criteria", nil, "acceptance criterion; repeatable")
	create.Flags().StringArrayVar(&verification, "verify", nil, "verification step; repeatable")
	create.Flags().StringArrayVar(&resources, "resource", nil, "resource entry; repeatable")

	var status string
	var addCriteria []string
	var addVerification []string
	var addResources []string
	update := &cobra.Command{
		Use:   "update <story-slug>",
		Short: "Update a story",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := planningManager().UpdateStory(args[0], planning.StoryChanges{
				Status:          status,
				AddCriteria:     addCriteria,
				AddVerification: addVerification,
				AddResources:    addResources,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated story %s\n", note.Path)
			return nil
		},
	}
	update.Flags().StringVar(&status, "status", "", "new status: todo, in_progress, blocked, done")
	update.Flags().StringArrayVar(&addCriteria, "criteria", nil, "acceptance criterion to append; repeatable")
	update.Flags().StringArrayVar(&addVerification, "verify", nil, "verification step to append; repeatable")
	update.Flags().StringArrayVar(&addResources, "resource", nil, "resource entry to append; repeatable")

	var epicFilter string
	var statusFilter string
	list := &cobra.Command{
		Use:   "list",
		Short: "List stories",
		RunE: func(cmd *cobra.Command, args []string) error {
			items, err := planningManager().ListStories(epicFilter, statusFilter)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No stories found.")
				return nil
			}
			printed := 0
			for _, item := range items {
				fmt.Fprintf(cmd.OutOrStdout(), "%s [%s] epic=%s spec=%s\n", item.Title, item.Status, item.Epic, item.Spec)
				printed++
			}
			if printed == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No stories found.")
			}
			return nil
		},
	}
	list.Flags().StringVar(&epicFilter, "epic", "", "filter by epic slug")
	list.Flags().StringVar(&statusFilter, "status", "", "filter by story status")

	show := &cobra.Command{
		Use:   "show <story-slug>",
		Short: "Show a story note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := planningManager().ReadStory(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n%s", note.Path, note.Content)
			return nil
		},
	}

	var apply bool
	slice := &cobra.Command{
		Use:   "slice <epic-slug>",
		Short: "Preview or apply first-pass story slices from an approved spec",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			preview, err := planningManager().PreviewStorySlices(args[0])
			if err != nil {
				return err
			}
			printStorySlicePreview(cmd.OutOrStdout(), preview)
			if !apply {
				return nil
			}
			ok, err := confirmStorySliceApply(bufio.NewReader(cmd.InOrStdin()), cmd.OutOrStdout())
			if err != nil {
				return err
			}
			if !ok {
				fmt.Fprintln(cmd.OutOrStdout(), "Apply canceled.")
				return nil
			}
			result, err := planningManager().ApplyStorySlices(args[0])
			if err != nil {
				return err
			}
			printStorySliceApplyResult(cmd.OutOrStdout(), result)
			return nil
		},
	}
	slice.Flags().BoolVar(&apply, "apply", false, "write missing story notes after preview and confirmation")

	critique := &cobra.Command{
		Use:   "critique <story-slug>",
		Short: "Interactively critique a story for execution readiness",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(cmd.InOrStdin())
			out := cmd.OutOrStdout()

			state, err := planningManager().ReadStoryCritique(args[0])
			if err != nil {
				return err
			}

			if err := runRefinementCluster(
				reader,
				out,
				"cluster 1/3: scope fit and vertical slice check",
				state.ScopeFit == "",
				"Scope Fit",
				"Describe whether this story is the right size and shape for the intended implementation slice.",
				&state.ScopeFit,
				state.VerticalSliceCheck == "",
				"Vertical Slice Check",
				"Describe whether this story delivers a vertical slice of value rather than only horizontal plumbing.",
				&state.VerticalSliceCheck,
			); err != nil {
				return err
			}
			if _, err := planningManager().UpdateStoryCritique(args[0], planning.StoryCritiqueInput{
				ScopeFit:           state.ScopeFit,
				VerticalSliceCheck: state.VerticalSliceCheck,
			}); err != nil {
				return err
			}

			if err := runRefinementCluster(
				reader,
				out,
				"cluster 2/3: hidden prerequisites and verification gaps",
				state.HiddenPrerequisites == "",
				"Hidden Prerequisites",
				"List prerequisites or missing inputs this story depends on. Enter one per line.",
				&state.HiddenPrerequisites,
				state.VerificationGaps == "",
				"Verification Gaps",
				"List any verification gaps or weak checks. Enter one per line.",
				&state.VerificationGaps,
			); err != nil {
				return err
			}
			if _, err := planningManager().UpdateStoryCritique(args[0], planning.StoryCritiqueInput{
				HiddenPrerequisites: state.HiddenPrerequisites,
				VerificationGaps:    state.VerificationGaps,
			}); err != nil {
				return err
			}

			if state.RewriteRecommendation == "" {
				value, err := promptSectionValue(reader, out, "Rewrite Recommendation", "Enter one of: keep, rewrite, reslice.")
				if err != nil {
					return err
				}
				state.RewriteRecommendation = value
			} else {
				fmt.Fprintf(out, "Skipping Rewrite Recommendation; already captured.\n")
			}
			if _, err := planningManager().UpdateStoryCritique(args[0], planning.StoryCritiqueInput{
				RewriteRecommendation: state.RewriteRecommendation,
			}); err != nil {
				return err
			}

			if state.HasGaps() {
				fmt.Fprintf(out, "Critique saved for %s with remaining gaps.\n", state.Path)
				return nil
			}
			fmt.Fprintf(out, "Critique saved for %s\n", state.Path)
			if state.HasBlockingRecommendation() {
				return fmt.Errorf("story critique recommends %s", state.RecommendationAction())
			}
			return nil
		},
	}

	cmd.AddCommand(create, update, list, show, slice, critique)
	return cmd
}

func printStorySlicePreview(out io.Writer, preview *planning.StorySlicePreview) {
	fmt.Fprintf(out, "story_slice_preview: %s\n", preview.SpecPath)
	fmt.Fprintf(out, "candidates: %d\n", len(preview.Candidates))
	for _, candidate := range preview.Candidates {
		status := "new"
		if candidate.StoryPath != "" {
			status = "existing:" + candidate.Status
		}
		fmt.Fprintf(out, "- %s [%s]\n", candidate.Title, status)
		fmt.Fprintf(out, "  description: %s\n", candidate.Description)
		for _, item := range candidate.AcceptanceCriteria {
			fmt.Fprintf(out, "  criteria: %s\n", item)
		}
		for _, item := range candidate.Verification {
			fmt.Fprintf(out, "  verify: %s\n", item)
		}
	}
}

func printStorySliceApplyResult(out io.Writer, result *planning.StorySliceApplyResult) {
	fmt.Fprintf(out, "story_slice_apply: %s\n", result.SpecPath)
	fmt.Fprintf(out, "created: %d\n", len(result.CreatedPaths))
	fmt.Fprintf(out, "reused: %d\n", len(result.SkippedPaths))
	for _, path := range result.CreatedPaths {
		fmt.Fprintf(out, "- created %s\n", path)
	}
	for _, path := range result.SkippedPaths {
		fmt.Fprintf(out, "- reused %s\n", path)
	}
}

func confirmStorySliceApply(reader *bufio.Reader, out io.Writer) (bool, error) {
	fmt.Fprint(out, "Apply these slices? [y/N]: ")
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}
