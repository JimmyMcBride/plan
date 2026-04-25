package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"plan/internal/planning"

	"github.com/spf13/cobra"
)

func newDiscussCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discuss",
		Short: "Assess and promote collaborative planning sources",
	}

	var (
		assessBrainstorm string
		assessDiscussion string
		assessFormat     string
	)
	assess := &cobra.Command{
		Use:   "assess",
		Short: "Assess whether a brainstorm or GitHub Discussion is ready for promotion",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := planningManager().AssessCollaborationSource(planning.CollaborationAssessInput{
				BrainstormSlug: assessBrainstorm,
				DiscussionRef:  assessDiscussion,
			})
			if err != nil {
				return err
			}
			return writeDiscussJSON(cmd, assessFormat, result)
		},
	}
	assess.Flags().StringVar(&assessBrainstorm, "brainstorm", "", "local brainstorm slug to assess")
	assess.Flags().StringVar(&assessDiscussion, "discussion", "", "GitHub Discussion number or URL to assess")
	assess.Flags().StringVar(&assessFormat, "format", "json", "output format: json")

	var (
		promoteBrainstorm      string
		promoteDiscussion      string
		promoteFormat          string
		promoteApply           bool
		promoteConfirm         bool
		promoteTarget          string
		promoteProjectDecision string
	)
	promote := &cobra.Command{
		Use:   "promote",
		Short: "Draft or apply a promotion from a brainstorm or GitHub Discussion",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !promoteApply {
				draft, err := planningManager().BuildPromotionDraft(planning.PromotionDraftInput{
					BrainstormSlug: promoteBrainstorm,
					DiscussionRef:  promoteDiscussion,
				})
				if err != nil {
					return err
				}
				return writeDiscussJSON(cmd, promoteFormat, draft)
			}
			result, err := planningManager().ApplyPromotionDraft(planning.PromotionApplyInput{
				BrainstormSlug:  promoteBrainstorm,
				DiscussionRef:   promoteDiscussion,
				Confirm:         promoteConfirm,
				TargetMode:      planning.SourceOfTruthMode(strings.TrimSpace(promoteTarget)),
				ProjectDecision: strings.TrimSpace(promoteProjectDecision),
			})
			if err != nil {
				var fallback *planning.PromotionApplyManualFallbackError
				if errors.As(err, &fallback) && fallback.Result != nil {
					if writeErr := writeDiscussJSON(cmd, promoteFormat, fallback.Result); writeErr != nil {
						return writeErr
					}
				}
				return err
			}
			return writeDiscussJSON(cmd, promoteFormat, result)
		},
	}
	promote.Flags().StringVar(&promoteBrainstorm, "brainstorm", "", "local brainstorm slug to promote")
	promote.Flags().StringVar(&promoteDiscussion, "discussion", "", "GitHub Discussion number or URL to promote")
	promote.Flags().StringVar(&promoteFormat, "format", "json", "output format: json")
	promote.Flags().BoolVar(&promoteApply, "apply", false, "create the promoted GitHub issue set after review")
	promote.Flags().BoolVar(&promoteConfirm, "confirm", false, "required acknowledgement before apply mutates GitHub")
	promote.Flags().StringVar(&promoteTarget, "target", "", "promotion ownership target: local, github, or hybrid")
	promote.Flags().StringVar(&promoteProjectDecision, "project-decision", "", "project prompt decision for 5+ spec promotions: create or skip")

	var (
		repairBrainstorm string
		repairDiscussion string
		repairFormat     string
		repairSpecs      []string
		repairConfirm    bool
	)
	repair := &cobra.Command{
		Use:     "repair",
		Aliases: []string{"repair-spec-split"},
		Short:   "Repair a collaboration source with a canonical Specs section",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := planningManager().RepairSpecSplit(planning.RepairSpecSplitInput{
				BrainstormSlug: repairBrainstorm,
				DiscussionRef:  repairDiscussion,
				Specs:          repairSpecs,
				Confirm:        repairConfirm,
			})
			if err != nil {
				return err
			}
			return writeDiscussJSON(cmd, repairFormat, result)
		},
	}
	repair.Flags().StringVar(&repairBrainstorm, "brainstorm", "", "local brainstorm slug to repair")
	repair.Flags().StringVar(&repairDiscussion, "discussion", "", "GitHub Discussion number or URL to repair")
	repair.Flags().StringVar(&repairFormat, "format", "json", "output format: json")
	repair.Flags().StringArrayVar(&repairSpecs, "spec", nil, "spec title for the repaired split; repeat for each spec")
	repair.Flags().BoolVar(&repairConfirm, "confirm", false, "required acknowledgement before repairing a GitHub Discussion")

	cmd.AddCommand(assess, promote, repair)
	return cmd
}

func writeDiscussJSON(cmd *cobra.Command, format string, v any) error {
	if strings.TrimSpace(format) != "json" {
		return fmt.Errorf("unsupported discuss output format %q; only json is supported in v1", format)
	}
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}
