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

func newBrainstormCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "brainstorm",
		Short: "Manage brainstorm notes",
	}

	var focusQuestion string
	var seedIdeas []string
	start := &cobra.Command{
		Use:   "start <topic>",
		Short: "Start a new brainstorm",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := planningManager().CreateBrainstormWithInput(planning.BrainstormCreateInput{
				Topic:         args[0],
				FocusQuestion: focusQuestion,
				Ideas:         seedIdeas,
			})
			if err != nil {
				return err
			}
			fmt.Printf("Created brainstorm %s\n", note.Path)
			return nil
		},
	}
	start.Flags().StringVar(&focusQuestion, "focus", "", "focus question to seed into the brainstorm")
	start.Flags().StringArrayVar(&seedIdeas, "idea", nil, "initial idea to capture; repeatable")

	var ideaBody string
	var ideaStdin bool
	var section string
	idea := &cobra.Command{
		Use:   "idea <brainstorm-slug>",
		Short: "Append an idea to a brainstorm",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readBody(cmd.InOrStdin(), ideaBody, ideaStdin)
			if err != nil {
				return err
			}
			if body == "" {
				return fmt.Errorf("idea body is required")
			}
			note, err := planningManager().AddBrainstormEntry(args[0], section, body)
			if err != nil {
				return err
			}
			fmt.Printf("Updated brainstorm %s\n", note.Path)
			return nil
		},
	}
	idea.Flags().StringVarP(&ideaBody, "body", "b", "", "idea body")
	idea.Flags().BoolVar(&ideaStdin, "stdin", false, "read idea body from stdin")
	idea.Flags().StringVar(&section, "section", "ideas", "brainstorm section: ideas, focus-question, desired-outcome, constraints, open-questions, raw-notes")

	show := &cobra.Command{
		Use:   "show <brainstorm-slug>",
		Short: "Show a brainstorm note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := planningManager().ReadBrainstorm(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("%s\n\n%s", note.Path, note.Content)
			return nil
		},
	}

	refine := &cobra.Command{
		Use:   "refine <brainstorm-slug>",
		Short: "Interactively refine a brainstorm into a clearer planning input",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(cmd.InOrStdin())
			out := cmd.OutOrStdout()

			state, err := planningManager().ReadBrainstormRefinement(args[0])
			if err != nil {
				return err
			}

			if err := runRefinementCluster(
				reader,
				out,
				"cluster 1/4: problem and user/value",
				state.Problem == "",
				"Problem",
				"Describe the core problem this brainstorm is trying to solve.",
				&state.Problem,
				state.UserValue == "",
				"User / Value",
				"Describe who benefits and what value they get if this works.",
				&state.UserValue,
			); err != nil {
				return err
			}
			if _, err := planningManager().UpdateBrainstormRefinement(args[0], planning.BrainstormRefinementInput{
				Problem:   state.Problem,
				UserValue: state.UserValue,
			}); err != nil {
				return err
			}

			if err := runRefinementCluster(
				reader,
				out,
				"cluster 2/4: constraints and appetite",
				state.Constraints == "",
				"Constraints",
				"List the constraints that should shape this work. Enter one per line.",
				&state.Constraints,
				state.Appetite == "",
				"Appetite",
				"Describe how big this should be. Keep it as a crisp boundary, not a schedule.",
				&state.Appetite,
			); err != nil {
				return err
			}
			if _, err := planningManager().UpdateBrainstormRefinement(args[0], planning.BrainstormRefinementInput{
				Constraints: state.Constraints,
				Appetite:    state.Appetite,
			}); err != nil {
				return err
			}

			if err := runRefinementCluster(
				reader,
				out,
				"cluster 3/4: open questions and candidate approaches",
				state.RemainingOpenQuestions == "",
				"Remaining Open Questions",
				"Capture unresolved questions that still matter. Enter one per line.",
				&state.RemainingOpenQuestions,
				state.CandidateApproaches == "",
				"Candidate Approaches",
				"List the strongest approaches worth carrying forward. Enter one per line.",
				&state.CandidateApproaches,
			); err != nil {
				return err
			}
			if _, err := planningManager().UpdateBrainstormRefinement(args[0], planning.BrainstormRefinementInput{
				RemainingOpenQuestions: state.RemainingOpenQuestions,
				CandidateApproaches:    state.CandidateApproaches,
			}); err != nil {
				return err
			}

			if state.DecisionSnapshot == "" {
				value, err := promptSectionValue(reader, out, "Decision Snapshot", "Summarize the current best direction or next decision in one short block.")
				if err != nil {
					return err
				}
				state.DecisionSnapshot = value
			} else {
				fmt.Fprintf(out, "Skipping Decision Snapshot; already captured.\n")
			}
			if _, err := planningManager().UpdateBrainstormRefinement(args[0], planning.BrainstormRefinementInput{
				DecisionSnapshot: state.DecisionSnapshot,
			}); err != nil {
				return err
			}

			if state.HasGaps() {
				fmt.Fprintf(out, "Refinement saved for %s with remaining gaps.\n", state.Path)
				return nil
			}
			fmt.Fprintf(out, "Refinement saved for %s\n", state.Path)
			return nil
		},
	}

	challenge := &cobra.Command{
		Use:   "challenge <brainstorm-slug>",
		Short: "Interactively challenge a brainstorm before it hardens into an epic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(cmd.InOrStdin())
			out := cmd.OutOrStdout()

			state, err := planningManager().ReadBrainstormChallenge(args[0])
			if err != nil {
				return err
			}

			if err := runRefinementCluster(
				reader,
				out,
				"cluster 1/3: rabbit holes and no-gos",
				state.RabbitHoles == "",
				"Rabbit Holes",
				"List the traps or rabbit holes that could make this idea sprawl. Enter one per line.",
				&state.RabbitHoles,
				state.NoGos == "",
				"No-Gos",
				"List the explicit no-gos or exclusions for this idea. Enter one per line.",
				&state.NoGos,
			); err != nil {
				return err
			}
			if _, err := planningManager().UpdateBrainstormChallenge(args[0], planning.BrainstormChallengeInput{
				RabbitHoles: state.RabbitHoles,
				NoGos:       state.NoGos,
			}); err != nil {
				return err
			}

			if err := runRefinementCluster(
				reader,
				out,
				"cluster 2/3: assumptions and likely overengineering",
				state.Assumptions == "",
				"Assumptions",
				"Capture the assumptions this idea depends on. Enter one per line.",
				&state.Assumptions,
				state.LikelyOverengineering == "",
				"Likely Overengineering",
				"Describe where this idea is most likely to become more complex than it should.",
				&state.LikelyOverengineering,
			); err != nil {
				return err
			}
			if _, err := planningManager().UpdateBrainstormChallenge(args[0], planning.BrainstormChallengeInput{
				Assumptions:           state.Assumptions,
				LikelyOverengineering: state.LikelyOverengineering,
			}); err != nil {
				return err
			}

			if state.SimplerAlternative == "" {
				value, err := promptSectionValue(reader, out, "Simpler Alternative", "Describe the simplest credible version of this idea.")
				if err != nil {
					return err
				}
				state.SimplerAlternative = value
			} else {
				fmt.Fprintf(out, "Skipping Simpler Alternative; already captured.\n")
			}
			if _, err := planningManager().UpdateBrainstormChallenge(args[0], planning.BrainstormChallengeInput{
				SimplerAlternative: state.SimplerAlternative,
			}); err != nil {
				return err
			}

			if state.HasGaps() {
				fmt.Fprintf(out, "Challenge saved for %s with remaining gaps.\n", state.Path)
				return nil
			}
			fmt.Fprintf(out, "Challenge saved for %s\n", state.Path)
			return nil
		},
	}

	cmd.AddCommand(start, idea, show, refine, challenge)
	return cmd
}

func runRefinementCluster(reader *bufio.Reader, out io.Writer, label string, askA bool, titleA, helpA string, valueA *string, askB bool, titleB, helpB string, valueB *string) error {
	if !askA && !askB {
		fmt.Fprintf(out, "Skipping %s; already complete.\n", label)
		return nil
	}
	fmt.Fprintf(out, "%s\n", label)
	if askA {
		value, err := promptSectionValue(reader, out, titleA, helpA)
		if err != nil {
			return err
		}
		if value != "" {
			*valueA = value
		}
	}
	if askB {
		value, err := promptSectionValue(reader, out, titleB, helpB)
		if err != nil {
			return err
		}
		if value != "" {
			*valueB = value
		}
	}
	fmt.Fprintf(out, "Saved %s.\n", label)
	return nil
}

func promptSectionValue(reader *bufio.Reader, out io.Writer, heading, help string) (string, error) {
	fmt.Fprintf(out, "%s\n%s\nFinish with a blank line. Leave empty to skip.\n", heading, help)
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			return strings.TrimSpace(strings.Join(lines, "\n")), nil
		}
		lines = append(lines, line)
		if errors.Is(err, io.EOF) {
			return strings.TrimSpace(strings.Join(lines, "\n")), nil
		}
	}
}
