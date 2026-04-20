package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"plan/internal/planning"
	"plan/internal/workspace"

	"github.com/spf13/cobra"
)

type guidedClusterPrompt struct {
	heading string
	help    string
	apply   func(input *planning.BrainstormRefinementInput, value string)
}

type guidedBrainstormCluster struct {
	index      int
	label      string
	prompts    []guidedClusterPrompt
	nextIndex  int
	nextLabel  string
	nextAction string
}

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
			reader := bufio.NewReader(cmd.InOrStdin())
			out := cmd.OutOrStdout()

			note, err := planningManager().CreateBrainstormWithInput(planning.BrainstormCreateInput{
				Topic:         args[0],
				FocusQuestion: focusQuestion,
				Ideas:         seedIdeas,
			})
			if err != nil {
				return err
			}
			if _, err := planningManager().EnsureGuidedBrainstormSession(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(out, "Created brainstorm %s\n", note.Path)

			vision, err := promptSectionValue(reader, out, "Vision", "Describe the vision in plain language. What do you see in your head? Focus on the outcome first, not the implementation.")
			if err != nil {
				return err
			}
			if _, _, err := planningManager().UpdateGuidedBrainstormIntake(args[0], planning.GuidedBrainstormIntakeInput{
				Vision: vision,
			}); err != nil {
				return err
			}

			supportingMaterial, err := promptSectionValue(reader, out, "Supporting Material", "List any relevant docs, links, or research you want to provide for this brainstorm. Enter one per line. Leave empty if none.")
			if err != nil {
				return err
			}
			note, session, err := planningManager().UpdateGuidedBrainstormIntake(args[0], planning.GuidedBrainstormIntakeInput{
				SupportingMaterial: supportingMaterial,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Guided brainstorm session ready for %s\nSummary: %s\nNext: %s\n", note.Path, session.Summary, session.NextAction)
			return nil
		},
	}
	start.Flags().StringVar(&focusQuestion, "focus", "", "focus question to seed into the brainstorm")
	start.Flags().StringArrayVar(&seedIdeas, "idea", nil, "initial idea to capture; repeatable")

	resume := &cobra.Command{
		Use:   "resume [brainstorm-slug]",
		Short: "Resume the active guided brainstorm session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(cmd.InOrStdin())
			out := cmd.OutOrStdout()

			var session *workspace.GuidedSessionRecord
			var err error
			if len(args) == 1 {
				session, err = planningManager().ReadGuidedSession("brainstorm/" + strings.TrimSpace(args[0]))
			} else {
				session, err = planningManager().ReadLastActiveGuidedSession()
			}
			if err != nil {
				return err
			}

			if session.Brainstorm == "" {
				return fmt.Errorf("guided session %s is not linked to a brainstorm", session.ChainID)
			}
			fmt.Fprintf(out, "Resuming %s\nSummary: %s\nNext: %s\n", session.ChainID, session.Summary, session.NextAction)

			action, err := promptSessionAction(reader, out)
			if err != nil {
				return err
			}
			switch action {
			case "3":
				updated, err := planningManager().UpdateGuidedSession(session.ChainID, planning.GuidedSessionUpdateInput{
					CurrentStage: "brainstorm",
					StageStatus:  "in_progress",
					NextAction:   "Resume the brainstorm when you are ready to continue shaping it.",
				})
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "Saved guided session state for %s\nNext: %s\n", updated.ChainID, updated.NextAction)
				return nil
			case "1", "2":
				return runGuidedBrainstormResume(reader, out, planningManager(), session, action == "2")
			default:
				return fmt.Errorf("unsupported menu action %q", action)
			}
		},
	}

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

	cmd.AddCommand(start, resume, idea, show, refine, challenge)
	return cmd
}

func runGuidedBrainstormResume(reader *bufio.Reader, out io.Writer, manager *planning.Manager, session *workspace.GuidedSessionRecord, forceCurrent bool) error {
	cluster := guidedBrainstormClusterForSession(session, forceCurrent)
	if cluster.index == 0 {
		updated, err := manager.UpdateGuidedSession(session.ChainID, planning.GuidedSessionUpdateInput{
			CurrentStage:        "brainstorm",
			CurrentCluster:      session.CurrentCluster,
			CurrentClusterLabel: session.CurrentClusterLabel,
			StageStatus:         "in_progress",
			Summary:             session.Summary,
			NextAction:          "Guided brainstorm clarification is complete for now.",
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Nothing new to resume for %s\nNext: %s\n", updated.ChainID, updated.NextAction)
		return nil
	}

	fmt.Fprintf(out, "%s\n", cluster.label)
	var input planning.BrainstormRefinementInput
	for _, prompt := range cluster.prompts {
		value, err := promptSectionValue(reader, out, prompt.heading, prompt.help)
		if err != nil {
			return err
		}
		prompt.apply(&input, value)
	}
	if _, err := manager.UpdateBrainstormRefinement(session.Brainstorm, input); err != nil {
		return err
	}

	state, err := manager.ReadBrainstormRefinement(session.Brainstorm)
	if err != nil {
		return err
	}
	reflection := renderBrainstormClusterReflection(cluster.index, state)
	fmt.Fprintf(out, "Reflection: %s\n", reflection)
	gap := renderBrainstormGapGuidance(cluster.index, state)
	if gap != "" {
		fmt.Fprintf(out, "%s\n", gap)
	}

	updated, err := manager.UpdateGuidedSession(session.ChainID, planning.GuidedSessionUpdateInput{
		CurrentStage:        "brainstorm",
		CurrentCluster:      cluster.nextIndex,
		CurrentClusterLabel: cluster.nextLabel,
		StageStatus:         "in_progress",
		Summary:             reflection,
		NextAction:          cluster.nextAction,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Saved cluster %d for %s\nNext: %s\n", cluster.index, updated.ChainID, updated.NextAction)
	return nil
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

func promptSessionAction(reader *bufio.Reader, out io.Writer) (string, error) {
	for {
		fmt.Fprintf(out, "Choose next step:\n1. Continue\n2. Refine\n3. Stop for now\n")
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}
		choice := strings.TrimSpace(line)
		if choice == "1" || choice == "2" || choice == "3" {
			return choice, nil
		}
		if errors.Is(err, io.EOF) && choice == "" {
			return "", fmt.Errorf("session action is required")
		}
		fmt.Fprintf(out, "Enter 1, 2, or 3.\n")
		if errors.Is(err, io.EOF) {
			return "", fmt.Errorf("unsupported menu action %q", choice)
		}
	}
}

func guidedBrainstormClusterForSession(session *workspace.GuidedSessionRecord, forceCurrent bool) guidedBrainstormCluster {
	label := session.CurrentClusterLabel
	if label == "" {
		label = "vision-intake"
	}
	if label == "vision-intake" {
		label = "clarify-problem-user-value"
	}
	if forceCurrent && session.CurrentCluster > 0 && session.CurrentClusterLabel != "" && session.CurrentClusterLabel != "vision-intake" {
		label = session.CurrentClusterLabel
	}

	switch label {
	case "clarify-problem-user-value":
		return guidedBrainstormCluster{
			index: 2,
			label: "cluster 1/3: problem and user/value",
			prompts: []guidedClusterPrompt{
				{
					heading: "Problem",
					help:    "Describe the core problem this brainstorm is trying to solve.",
					apply: func(input *planning.BrainstormRefinementInput, value string) {
						input.Problem = value
					},
				},
				{
					heading: "User / Value",
					help:    "Describe who benefits and what value they get if this works.",
					apply: func(input *planning.BrainstormRefinementInput, value string) {
						input.UserValue = value
					},
				},
			},
			nextIndex:  3,
			nextLabel:  "clarify-constraints-appetite",
			nextAction: "Continue with constraints and appetite.",
		}
	case "clarify-constraints-appetite":
		return guidedBrainstormCluster{
			index: 3,
			label: "cluster 2/3: constraints and appetite",
			prompts: []guidedClusterPrompt{
				{
					heading: "Constraints",
					help:    "List the constraints that should shape this work. Enter one per line.",
					apply: func(input *planning.BrainstormRefinementInput, value string) {
						input.Constraints = value
					},
				},
				{
					heading: "Appetite",
					help:    "Describe how big this should be. Keep it as a crisp boundary, not a schedule.",
					apply: func(input *planning.BrainstormRefinementInput, value string) {
						input.Appetite = value
					},
				},
			},
			nextIndex:  4,
			nextLabel:  "clarify-open-approaches",
			nextAction: "Continue with open questions and candidate approaches.",
		}
	case "clarify-open-approaches":
		return guidedBrainstormCluster{
			index: 4,
			label: "cluster 3/3: open questions and candidate approaches",
			prompts: []guidedClusterPrompt{
				{
					heading: "Remaining Open Questions",
					help:    "Capture unresolved questions that still matter. Enter one per line.",
					apply: func(input *planning.BrainstormRefinementInput, value string) {
						input.RemainingOpenQuestions = value
					},
				},
				{
					heading: "Candidate Approaches",
					help:    "List the strongest approaches worth carrying forward. Enter one per line.",
					apply: func(input *planning.BrainstormRefinementInput, value string) {
						input.CandidateApproaches = value
					},
				},
			},
			nextIndex:  4,
			nextLabel:  "clarify-open-approaches",
			nextAction: "Review recap and decide whether to refine more or move to the next stage.",
		}
	default:
		return guidedBrainstormCluster{}
	}
}

func renderBrainstormClusterReflection(index int, state *planning.BrainstormRefinement) string {
	switch index {
	case 2:
		var parts []string
		if strings.TrimSpace(state.Problem) != "" {
			parts = append(parts, "You are trying to solve "+strings.TrimSpace(state.Problem)+".")
		}
		if strings.TrimSpace(state.UserValue) != "" {
			parts = append(parts, "The main value is "+strings.TrimSpace(state.UserValue)+".")
		}
		return strings.Join(parts, " ")
	case 3:
		var parts []string
		if strings.TrimSpace(state.Constraints) != "" {
			parts = append(parts, "The work is bounded by the current constraints.")
		}
		if strings.TrimSpace(state.Appetite) != "" {
			parts = append(parts, "The current appetite is "+strings.TrimSpace(state.Appetite)+".")
		}
		return strings.Join(parts, " ")
	case 4:
		var parts []string
		if strings.TrimSpace(state.RemainingOpenQuestions) != "" {
			parts = append(parts, "There are still a few open questions to resolve.")
		}
		if strings.TrimSpace(state.CandidateApproaches) != "" {
			parts = append(parts, "You have candidate approaches worth carrying forward.")
		}
		return strings.Join(parts, " ")
	default:
		return "Guided brainstorm progress updated."
	}
}

func renderBrainstormGapGuidance(index int, state *planning.BrainstormRefinement) string {
	switch index {
	case 2:
		if looksVague(state.Problem) || looksVague(state.UserValue) {
			return "Potential gap: the problem or value is still vague.\nRecommended path: tighten the user and outcome into one concrete pain point.\nAlternative 1: narrow to one user group.\nAlternative 2: state the smallest credible value signal."
		}
	case 3:
		if looksVague(state.Appetite) || strings.TrimSpace(state.Constraints) == "" {
			return "Potential gap: scope is still too loose.\nRecommended path: set one hard boundary the work will not cross.\nAlternative 1: cap this to one focused implementation pass.\nAlternative 2: move optional depth into the roadmap parking lot."
		}
	case 4:
		if strings.TrimSpace(state.CandidateApproaches) == "" || strings.TrimSpace(state.RemainingOpenQuestions) == "" {
			return "Potential gap: the path forward is not shaped enough yet.\nRecommended path: pick one approach and name the biggest unanswered risk.\nAlternative 1: park the weaker approach for later.\nAlternative 2: rewrite the question list down to the blockers only."
		}
	}
	return ""
}

func looksVague(value string) bool {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return true
	}
	for _, needle := range []string{"not sure", "maybe", "everything", "anything", "somehow", "flexible"} {
		if strings.Contains(trimmed, needle) {
			return true
		}
	}
	return false
}
