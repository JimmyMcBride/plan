package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"plan/internal/planning"

	"github.com/spf13/cobra"
)

func newEpicCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epic",
		Short: "Manage epics",
	}

	create := &cobra.Command{
		Use:   "create <title>",
		Short: "Create an epic and its draft spec",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundle, err := planningManager().CreateEpic(args[0], "")
			if err != nil {
				return err
			}
			fmt.Printf("Created epic %s\n", bundle.Epic.Path)
			fmt.Printf("Created spec %s\n", bundle.Spec.Path)
			return nil
		},
	}

	promote := &cobra.Command{
		Use:   "promote <brainstorm-slug>",
		Short: "Promote a brainstorm into an epic and seeded spec",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundle, err := planningManager().PromoteBrainstorm(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Created epic %s\n", bundle.Epic.Path)
			fmt.Printf("Created seeded spec %s\n", bundle.Spec.Path)
			return nil
		},
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "List epics",
		RunE: func(cmd *cobra.Command, args []string) error {
			epics, err := planningManager().ListEpics()
			if err != nil {
				return err
			}
			if len(epics) == 0 {
				fmt.Println("No epics found.")
				return nil
			}
			for _, epic := range epics {
				fmt.Printf("%s [%s] (%d/%d stories done)\n",
					epic.Title,
					epic.SpecStatus,
					epic.DoneStories,
					epic.TotalStories,
				)
			}
			return nil
		},
	}

	show := &cobra.Command{
		Use:   "show <epic-slug>",
		Short: "Show an epic note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := planningManager().ReadEpic(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("%s\n\n%s", note.Path, note.Content)
			return nil
		},
	}

	shape := &cobra.Command{
		Use:   "shape <epic-slug>",
		Short: "Interactively shape an epic with appetite and scope boundaries",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(cmd.InOrStdin())
			out := cmd.OutOrStdout()

			state, err := planningManager().ReadEpicShape(args[0])
			if err != nil {
				return err
			}

			if err := runRefinementCluster(
				reader,
				out,
				"cluster 1/3: appetite and outcome",
				state.Appetite == "",
				"Appetite",
				"Describe how big this epic should be. Keep it as a boundary, not a schedule.",
				&state.Appetite,
				state.Outcome == "",
				"Outcome",
				"Describe the concrete outcome this epic should achieve if it succeeds.",
				&state.Outcome,
			); err != nil {
				return err
			}
			if _, err := planningManager().UpdateEpicShape(args[0], planning.EpicShapeInput{
				Appetite: state.Appetite,
				Outcome:  state.Outcome,
			}); err != nil {
				return err
			}

			if err := runRefinementCluster(
				reader,
				out,
				"cluster 2/3: scope boundary and out of scope",
				state.ScopeBoundary == "",
				"Scope Boundary",
				"Describe the boundary of the work this epic should cover.",
				&state.ScopeBoundary,
				state.OutOfScope == "",
				"Out of Scope",
				"List the work this epic will explicitly not do. Enter one per line.",
				&state.OutOfScope,
			); err != nil {
				return err
			}
			if _, err := planningManager().UpdateEpicShape(args[0], planning.EpicShapeInput{
				ScopeBoundary: state.ScopeBoundary,
				OutOfScope:    state.OutOfScope,
			}); err != nil {
				return err
			}

			if state.SuccessSignal == "" {
				value, err := promptSectionValue(reader, out, "Success Signal", "Describe how you will know this epic is successful.")
				if err != nil {
					return err
				}
				state.SuccessSignal = value
			} else {
				fmt.Fprintf(out, "Skipping Success Signal; already captured.\n")
			}
			if _, err := planningManager().UpdateEpicShape(args[0], planning.EpicShapeInput{
				SuccessSignal: state.SuccessSignal,
			}); err != nil {
				return err
			}

			if state.HasGaps() {
				fmt.Fprintf(out, "Shape saved for %s with remaining gaps.\n", state.Path)
				return nil
			}
			fmt.Fprintf(out, "Shape saved for %s\n", state.Path)
			return nil
		},
	}

	handoff := &cobra.Command{
		Use:   "handoff <epic-slug>",
		Short: "Continue a guided epic into the spec stage at a stable checkpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(cmd.InOrStdin())
			out := cmd.OutOrStdout()
			session, err := planningManager().ReadGuidedSessionByEpic(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Epic recap:\nCurrent understanding: %s\nRecommended next stage: continue into spec.\nProceed? [y/N]\n", session.Summary)
			confirm, err := reader.ReadString('\n')
			if err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
				fmt.Fprintf(out, "Canceled epic handoff for %s\n", args[0])
				return nil
			}
			updated, spec, err := planningManager().AdvanceGuidedSessionToSpec(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Using spec %s\nNext: %s\n", spec.Path, updated.NextAction)
			return nil
		},
	}

	cmd.AddCommand(create, promote, list, show, shape, handoff)
	return cmd
}
