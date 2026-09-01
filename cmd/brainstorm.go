package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"plan/internal/planning"
	"plan/internal/workspace"

	brainplanning "github.com/JimmyMcBride/brain/planning"
	brainapp "github.com/JimmyMcBride/brain/planning/application"
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
			if handled, err := withSharedPlanning(cmd, "brainstorm start", false, func(service *brainapp.Service) error {
				reader := bufio.NewReader(cmd.InOrStdin())
				out := cmd.OutOrStdout()
				result, err := service.CreateBrainstorm(cmd.Context(), brainapp.CreateBrainstormInput{Title: args[0], Confirmed: true}, sharedAuthorizer(), sharedEvents())
				if err != nil {
					return err
				}
				id := result.Document.Artifact.ID
				if strings.TrimSpace(focusQuestion) != "" {
					if _, err := service.UpdateBrainstorm(cmd.Context(), brainapp.BrainstormUpdateInput{ID: id, Section: "focus-question", Body: focusQuestion, Confirmed: true}, sharedAuthorizer(), sharedEvents()); err != nil {
						return err
					}
				}
				for _, idea := range seedIdeas {
					if strings.TrimSpace(idea) == "" {
						continue
					}
					if _, err := service.UpdateBrainstorm(cmd.Context(), brainapp.BrainstormUpdateInput{ID: id, Section: "ideas", Body: idea, Confirmed: true}, sharedAuthorizer(), sharedEvents()); err != nil {
						return err
					}
				}
				fmt.Fprintf(out, "Created brainstorm %s\n", result.Document.Path)
				vision, err := promptSectionValue(reader, out, "Vision", "Describe the vision in plain language. What do you see in your head? Focus on the outcome first, not the implementation.")
				if err != nil {
					return err
				}
				if _, _, err := planningManager().UpdateGuidedBrainstormIntake(string(id), planning.GuidedBrainstormIntakeInput{Vision: vision}); err != nil {
					return err
				}
				supportingMaterial, err := promptSectionValue(reader, out, "Supporting Material", "List any relevant docs, links, or research you want to provide for this brainstorm. Enter one per line. Leave empty if none.")
				if err != nil {
					return err
				}
				note, session, err := planningManager().UpdateGuidedBrainstormIntake(string(id), planning.GuidedBrainstormIntakeInput{SupportingMaterial: supportingMaterial})
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "Guided brainstorm session ready for %s\nSummary: %s\nNext: %s\n", note.Path, session.Summary, session.NextAction)
				return nil
			}); handled {
				return err
			}
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
			if handled, err := withSharedPlanning(cmd, "brainstorm resume", false, func(service *brainapp.Service) error {
				reader := bufio.NewReader(cmd.InOrStdin())
				out := cmd.OutOrStdout()
				var session brainapp.GuidedSessionRecord
				var err error
				if len(args) == 1 {
					session, err = service.GetGuidedSession(cmd.Context(), args[0])
				} else {
					session, err = service.CurrentGuidedSession(cmd.Context())
				}
				if err != nil {
					return err
				}
				legacySession := workspaceSessionFromShared(session)
				if legacySession.Brainstorm == "" {
					return fmt.Errorf("guided session %s is not linked to a brainstorm", legacySession.ChainID)
				}
				fmt.Fprintf(out, "Resuming %s\nSummary: %s\nNext: %s\n", legacySession.ChainID, legacySession.Summary, legacySession.NextAction)
				action, err := promptSessionAction(reader, out)
				if err != nil {
					return err
				}
				switch action {
				case "3":
					updated, err := planningManager().UpdateGuidedSession(legacySession.ChainID, planning.GuidedSessionUpdateInput{
						CurrentStage: "brainstorm", StageStatus: "in_progress",
						NextAction: "Resume the brainstorm when you are ready to continue shaping it.",
					})
					if err != nil {
						return err
					}
					fmt.Fprintf(out, "Saved guided session state for %s\nNext: %s\n", updated.ChainID, updated.NextAction)
					return nil
				case "1", "2":
					return runGuidedBrainstormResume(reader, out, planningManager(), legacySession, action == "2")
				default:
					return fmt.Errorf("unsupported menu action %q", action)
				}
			}); handled {
				return err
			}
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

	sessions := &cobra.Command{
		Use:   "sessions",
		Short: "List guided brainstorm sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			if handled, err := withSharedPlanning(cmd, "brainstorm sessions", false, func(service *brainapp.Service) error {
				out := cmd.OutOrStdout()
				list, err := service.ListGuidedSessions(cmd.Context())
				if err != nil {
					return err
				}
				current, err := service.CurrentGuidedSession(cmd.Context())
				if err != nil && !strings.Contains(err.Error(), "no active guided session") {
					return err
				}
				for _, session := range list {
					marker := " "
					if session.ChainID == current.ChainID {
						marker = "*"
					}
					fmt.Fprintf(out, "%s %s stage=%s next=%s\n", marker, session.ChainID, session.CurrentStage, session.NextAction)
				}
				return nil
			}); handled {
				return err
			}
			out := cmd.OutOrStdout()
			list, err := planningManager().ListGuidedSessions()
			if err != nil {
				return err
			}
			last, err := planningManager().ReadLastActiveGuidedSession()
			if err != nil && !strings.Contains(err.Error(), "no active guided session") {
				return err
			}
			lastActive := ""
			if last != nil {
				lastActive = last.ChainID
			}
			for _, session := range list {
				marker := " "
				if session.ChainID == lastActive {
					marker = "*"
				}
				fmt.Fprintf(out, "%s %s stage=%s next=%s\n", marker, session.ChainID, session.CurrentStage, session.NextAction)
			}
			return nil
		},
	}

	switchCmd := &cobra.Command{
		Use:   "switch <brainstorm-slug>",
		Short: "Switch the last-active guided brainstorm session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if handled, err := withSharedPlanning(cmd, "brainstorm switch", false, func(service *brainapp.Service) error {
				result, err := service.SwitchGuidedSession(cmd.Context(), brainapp.GuidedSessionMutationInput{ChainID: "brainstorm/" + strings.TrimSpace(args[0]), Confirmed: true}, sharedAuthorizer(), sharedEvents())
				if err != nil {
					return err
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Switched to %s\nSummary: %s\nNext: %s\n", result.Session.ChainID, result.Session.Summary, result.Session.NextAction)
				return err
			}); handled {
				return err
			}
			out := cmd.OutOrStdout()
			session, err := planningManager().SwitchGuidedSession("brainstorm/" + strings.TrimSpace(args[0]))
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Switched to %s\nSummary: %s\nNext: %s\n", session.ChainID, session.Summary, session.NextAction)
			return nil
		},
	}

	reopen := &cobra.Command{
		Use:   "reopen <brainstorm-slug> <stage>",
		Short: "Reopen a stage and mark downstream stages as needs review",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[1] != "epic" {
				if handled, err := withSharedPlanning(cmd, "brainstorm reopen", false, func(service *brainapp.Service) error {
					reader := bufio.NewReader(cmd.InOrStdin())
					out := cmd.OutOrStdout()
					chainID := "brainstorm/" + strings.TrimSpace(args[0])
					stage := strings.TrimSpace(args[1])
					session, err := service.GetGuidedSession(cmd.Context(), chainID)
					if err != nil {
						return err
					}
					downstream := guidedDownstreamStagesForOutput(stage)
					fmt.Fprintf(out, "Reopen stage: %s\nDownstream stages affected: %s\nRecommended path: reopen this stage and mark downstream work as needs review.\nAlternative 1: stay on the current stage and refine locally.\nAlternative 2: stop for now and revisit after a recap pass.\nProceed? [y/N]\n", stage, strings.Join(downstream, ", "))
					confirm, err := reader.ReadString('\n')
					if err != nil && !errors.Is(err, io.EOF) {
						return err
					}
					if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
						fmt.Fprintf(out, "Canceled reopen for %s\n", session.ChainID)
						return nil
					}
					result, err := service.ReopenGuidedSession(cmd.Context(), brainapp.GuidedSessionMutationInput{ChainID: chainID, Stage: stage, Confirmed: true}, sharedAuthorizer(), sharedEvents())
					if err != nil {
						return err
					}
					fmt.Fprintf(out, "Reopened %s for %s\nMarked needs review: %s\nNext: %s\n", stage, result.Session.ChainID, strings.Join(downstream, ", "), result.Session.NextAction)
					return nil
				}); handled {
					return err
				}
			}
			reader := bufio.NewReader(cmd.InOrStdin())
			out := cmd.OutOrStdout()
			chainID := "brainstorm/" + strings.TrimSpace(args[0])
			stage := strings.TrimSpace(args[1])
			session, err := planningManager().ReadGuidedSession(chainID)
			if err != nil {
				return err
			}
			downstream := guidedDownstreamStagesForOutput(stage)
			fmt.Fprintf(out, "Reopen stage: %s\nDownstream stages affected: %s\nRecommended path: reopen this stage and mark downstream work as needs review.\nAlternative 1: stay on the current stage and refine locally.\nAlternative 2: stop for now and revisit after a recap pass.\nProceed? [y/N]\n", stage, strings.Join(downstream, ", "))
			confirm, err := reader.ReadString('\n')
			if err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
				fmt.Fprintf(out, "Canceled reopen for %s\n", session.ChainID)
				return nil
			}
			updated, impacted, err := planningManager().ReopenGuidedSessionStage(chainID, stage)
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Reopened %s for %s\nMarked needs review: %s\nNext: %s\n", stage, updated.ChainID, strings.Join(impacted, ", "), updated.NextAction)
			return nil
		},
	}

	review := &cobra.Command{
		Use:   "review <brainstorm-slug>",
		Short: "Run lightweight downstream review checkpoints for needs-review stages",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if handled, err := withSharedPlanning(cmd, "brainstorm review", false, func(service *brainapp.Service) error {
				chainID := "brainstorm/" + strings.TrimSpace(args[0])
				result, err := service.ReviewGuidedSession(cmd.Context(), brainapp.GuidedSessionMutationInput{ChainID: chainID, Confirmed: true}, sharedAuthorizer(), sharedEvents())
				if err != nil {
					return err
				}
				legacySession, legacyReviewed, err := planningManager().ReviewGuidedSessionStages(chainID)
				if err != nil {
					return err
				}
				reviewed := append([]string(nil), result.Impacted...)
				for _, stage := range legacyReviewed {
					if !containsValue(reviewed, stage) {
						reviewed = append(reviewed, stage)
					}
				}
				if len(reviewed) == 0 {
					_, err = fmt.Fprintf(cmd.OutOrStdout(), "No downstream review checkpoints needed for %s\n", legacySession.ChainID)
					return err
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Reviewed stages for %s: %s\nNext: %s\n", legacySession.ChainID, strings.Join(reviewed, ", "), legacySession.NextAction)
				return err
			}); handled {
				return err
			}
			out := cmd.OutOrStdout()
			session, reviewed, err := planningManager().ReviewGuidedSessionStages("brainstorm/" + strings.TrimSpace(args[0]))
			if err != nil {
				return err
			}
			if len(reviewed) == 0 {
				fmt.Fprintf(out, "No downstream review checkpoints needed for %s\n", session.ChainID)
				return nil
			}
			fmt.Fprintf(out, "Reviewed stages for %s: %s\nNext: %s\n", session.ChainID, strings.Join(reviewed, ", "), session.NextAction)
			return nil
		},
	}

	var parkValue string
	var parkReason string
	var parkUnlock string
	park := &cobra.Command{
		Use:   "park <brainstorm-slug> <title>",
		Short: "Park a good-but-early idea in ROADMAP.md",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if handled, err := withSharedPlanning(cmd, "brainstorm park", false, func(service *brainapp.Service) error {
				_, err := service.ParkRoadmap(cmd.Context(), brainapp.RoadmapParkingInput{
					BrainstormID: brainplanning.ArtifactID(args[0]), Title: args[1], Value: parkValue,
					Reason: parkReason, Unlock: parkUnlock, Confirmed: true,
				}, sharedAuthorizer(), sharedEvents())
				if err != nil {
					return err
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Parked %s in .plan/ROADMAP.md\n", args[1])
				return err
			}); handled {
				return err
			}
			source := fmt.Sprintf("[Brainstorm](../brainstorms/%s.md)", strings.TrimSpace(args[0]))
			if err := planningManager().AddRoadmapParkingEntry(planning.RoadmapParkingInput{
				Title:  args[1],
				Value:  parkValue,
				Reason: parkReason,
				Unlock: parkUnlock,
				Source: source,
			}); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Parked %s in .plan/ROADMAP.md\n", args[1])
			return nil
		},
	}
	park.Flags().StringVar(&parkValue, "value", "", "value or outcome if the idea lands later")
	park.Flags().StringVar(&parkReason, "reason", "", "why the idea is parked now")
	park.Flags().StringVar(&parkUnlock, "unlock", "", "what needs to happen before the idea becomes active")

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
			if handled, err := withSharedPlanning(cmd, "brainstorm idea", false, func(service *brainapp.Service) error {
				result, err := service.UpdateBrainstorm(cmd.Context(), brainapp.BrainstormUpdateInput{ID: brainplanning.ArtifactID(args[0]), Section: section, Body: body, Confirmed: true}, sharedAuthorizer(), sharedEvents())
				if err != nil {
					return err
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Updated brainstorm %s\n", result.Document.Path)
				return err
			}); handled {
				return err
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
			if handled, err := withSharedPlanning(cmd, "brainstorm show", false, func(service *brainapp.Service) error {
				document, err := service.GetBrainstorm(cmd.Context(), brainplanning.ArtifactID(args[0]))
				if err != nil {
					return err
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n%s", document.Path, document.Body)
				return err
			}); handled {
				return err
			}
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
			service, shared, err := sharedPlanningService(cmd.Context())
			if err != nil {
				return err
			}
			updateRefinement := func(input planning.BrainstormRefinementInput) error {
				_, err := planningManager().UpdateBrainstormRefinement(args[0], input)
				return err
			}
			if shared {
				defer warnSharedPlanning(cmd, "brainstorm refine", false)
				updateRefinement = func(input planning.BrainstormRefinementInput) error {
					_, err := service.RefineBrainstorm(cmd.Context(), brainapp.BrainstormRefinementInput{
						ID: brainplanning.ArtifactID(args[0]), Problem: input.Problem, UserValue: input.UserValue,
						Constraints: input.Constraints, Appetite: input.Appetite,
						RemainingOpenQuestions: input.RemainingOpenQuestions, CandidateApproaches: input.CandidateApproaches,
						DecisionSnapshot: input.DecisionSnapshot, Confirmed: true,
					}, sharedAuthorizer(), sharedEvents())
					return err
				}
			}

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
			if err := updateRefinement(planning.BrainstormRefinementInput{
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
			if err := updateRefinement(planning.BrainstormRefinementInput{
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
			if err := updateRefinement(planning.BrainstormRefinementInput{
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
			if err := updateRefinement(planning.BrainstormRefinementInput{
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
			service, shared, err := sharedPlanningService(cmd.Context())
			if err != nil {
				return err
			}
			updateChallenge := func(input planning.BrainstormChallengeInput) error {
				_, err := planningManager().UpdateBrainstormChallenge(args[0], input)
				return err
			}
			if shared {
				defer warnSharedPlanning(cmd, "brainstorm challenge", false)
				updateChallenge = func(input planning.BrainstormChallengeInput) error {
					_, err := service.ChallengeBrainstorm(cmd.Context(), brainapp.BrainstormChallengeInput{
						ID: brainplanning.ArtifactID(args[0]), RabbitHoles: input.RabbitHoles, NoGos: input.NoGos,
						Assumptions: input.Assumptions, LikelyOverengineering: input.LikelyOverengineering,
						SimplerAlternative: input.SimplerAlternative, Confirmed: true,
					}, sharedAuthorizer(), sharedEvents())
					return err
				}
			}

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
			if err := updateChallenge(planning.BrainstormChallengeInput{
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
			if err := updateChallenge(planning.BrainstormChallengeInput{
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
			if err := updateChallenge(planning.BrainstormChallengeInput{
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

	cmd.AddCommand(start, resume, sessions, switchCmd, reopen, review, park, idea, show, refine, challenge)
	return cmd
}

func containsValue(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func workspaceSessionFromShared(session brainapp.GuidedSessionRecord) *workspace.GuidedSessionRecord {
	return &workspace.GuidedSessionRecord{
		ChainID: session.ChainID, Brainstorm: session.Brainstorm, Spec: session.Spec,
		CurrentStage: session.CurrentStage, CurrentCluster: session.CurrentCluster,
		CurrentClusterLabel: session.CurrentClusterLabel, StageStatuses: session.StageStatuses,
		Summary: session.Summary, NextAction: session.NextAction,
		CreatedAt: session.CreatedAt, UpdatedAt: session.UpdatedAt,
	}
}

func runGuidedBrainstormResume(reader *bufio.Reader, out io.Writer, manager *planning.Manager, session *workspace.GuidedSessionRecord, forceCurrent bool) error {
	cluster := guidedBrainstormClusterForSession(session, forceCurrent)
	if cluster.index == 0 {
		fmt.Fprintf(out, "Brainstorm stage is shaped enough to hand off.\nContinue into epic creation? [y/N]\n")
		confirm, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
			updated, err := manager.UpdateGuidedSession(session.ChainID, planning.GuidedSessionUpdateInput{
				CurrentStage:        "brainstorm",
				CurrentCluster:      session.CurrentCluster,
				CurrentClusterLabel: session.CurrentClusterLabel,
				StageStatus:         "in_progress",
				Summary:             session.Summary,
				NextAction:          "Brainstorm handoff is ready when you want to continue into the epic stage.",
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(out, "Checkpoint saved for %s\nNext: %s\n", updated.ChainID, updated.NextAction)
			return nil
		}
		bundle, updated, err := manager.PromoteGuidedBrainstormSession(session.Brainstorm)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Created epic %s\nCreated seeded spec %s\nNext: %s\n", bundle.Epic.Path, bundle.Spec.Path, updated.NextAction)
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
	recap := renderBrainstormStageRecap(state, cluster.nextAction)
	fmt.Fprintf(out, "%s\n", recap)

	updated, err := manager.UpdateGuidedSession(session.ChainID, planning.GuidedSessionUpdateInput{
		CurrentStage:        "brainstorm",
		CurrentCluster:      cluster.nextIndex,
		CurrentClusterLabel: cluster.nextLabel,
		StageStatus:         "in_progress",
		Summary:             recap,
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
			nextIndex:  5,
			nextLabel:  "handoff-epic",
			nextAction: "Review recap and continue into the epic stage when ready.",
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

func renderBrainstormStageRecap(state *planning.BrainstormRefinement, nextAction string) string {
	currentUnderstanding := "Current understanding: vision shaping is in progress."
	if strings.TrimSpace(state.Problem) != "" && strings.TrimSpace(state.UserValue) != "" {
		currentUnderstanding = fmt.Sprintf("Current understanding: %s / %s", strings.TrimSpace(state.Problem), strings.TrimSpace(state.UserValue))
	}
	keyDecisions := "Key decisions: none yet."
	if strings.TrimSpace(state.Appetite) != "" {
		keyDecisions = "Key decisions: appetite captured."
	}
	unresolved := "Unresolved risks or questions: none captured yet."
	if strings.TrimSpace(state.RemainingOpenQuestions) != "" {
		unresolved = "Unresolved risks or questions: there are still open questions to review."
	}
	return strings.Join([]string{
		"Recap:",
		currentUnderstanding,
		keyDecisions,
		unresolved,
		"Parked items: none.",
		"Recommended next stage: " + nextAction,
	}, "\n")
}

func guidedDownstreamStagesForOutput(stage string) []string {
	switch strings.TrimSpace(stage) {
	case "brainstorm":
		return []string{"epic", "spec", "execution"}
	case "epic":
		return []string{"spec", "execution"}
	case "spec":
		return []string{"execution"}
	default:
		return []string{"none"}
	}
}
