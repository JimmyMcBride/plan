package cmd

import (
	"encoding/json"
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
		promoteBrainstorm string
		promoteDiscussion string
		promoteFormat     string
		promoteApply      bool
		promoteConfirm    bool
		promoteTarget     string
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
				BrainstormSlug: promoteBrainstorm,
				DiscussionRef:  promoteDiscussion,
				Confirm:        promoteConfirm,
				TargetMode:     planning.SourceOfTruthMode(strings.TrimSpace(promoteTarget)),
			})
			if err != nil {
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

	cmd.AddCommand(assess, promote)
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
