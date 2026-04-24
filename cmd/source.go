package cmd

import (
	"fmt"

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
			return nil
		},
	}

	set := &cobra.Command{
		Use:   "set <local|github|hybrid>",
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
