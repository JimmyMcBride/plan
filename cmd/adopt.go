package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newAdoptCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "adopt",
		Short: "Adopt an existing repo into a managed .plan workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := workspaceManager().Adopt()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Adopted plan workspace at %s\n", result.Info.PlanDir)
			for _, item := range result.Created {
				fmt.Fprintf(out, "  created %s\n", item)
			}
			for _, item := range result.Updated {
				fmt.Fprintf(out, "  updated %s\n", item)
			}
			if len(result.Created) == 0 && len(result.Updated) == 0 {
				fmt.Fprintln(out, "  no changes")
			}
			return nil
		},
	}
}
