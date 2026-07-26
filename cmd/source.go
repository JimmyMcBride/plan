package cmd

import (
	"fmt"
	"strings"

	"plan/internal/workspace"

	"github.com/spf13/cobra"
)

func newSourceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "source",
		Short: "Inspect or change the workspace source-of-truth mode",
	}

	show := &cobra.Command{
		Use:   "show",
		Short: "Show the current workspace source-of-truth mode",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mode, err := planningManager().SourceMode()
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "source_mode: %s\n", mode)
			if mode == workspace.SourceOfTruthLinear {
				state, err := planningManager().LinearConfig()
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "linear_team: %s\n", linearTeamSummary(state))
				fmt.Fprintln(cmd.OutOrStdout(), "linear_ownership: durable planning data lives in Linear after promotion")
				if state.TeamID == "" && state.TeamKey == "" {
					fmt.Fprintln(cmd.OutOrStdout(), "linear_guidance: configure .plan/.meta/linear.json with team_id or team_key before Linear promotion")
				}
			}
			return nil
		},
	}

	set := &cobra.Command{
		Use:   "set <local|github|hybrid|linear>",
		Short: "Set the workspace source-of-truth mode",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			meta, err := planningManager().SetSourceMode(workspace.SourceOfTruthMode(args[0]))
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "source_mode: %s\n", meta.SourceMode)
			return nil
		},
	}

	cmd.AddCommand(show, set)
	return cmd
}

func linearTeamSummary(state *workspace.LinearState) string {
	if state == nil || (state.TeamID == "" && state.TeamKey == "") {
		return "not_configured"
	}
	var parts []string
	if state.TeamKey != "" {
		parts = append(parts, state.TeamKey)
	}
	if state.TeamName != "" {
		parts = append(parts, state.TeamName)
	}
	if state.TeamID != "" {
		parts = append(parts, state.TeamID)
	}
	return strings.Join(parts, " / ")
}
