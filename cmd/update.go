package cmd

import (
	"fmt"

	"plan/internal/workspace"

	"github.com/spf13/cobra"
)

func newUpdateCommand() *cobra.Command {
	var archiveLegacy bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Repair or normalize the local .plan workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := workspaceManager().UpdateWithOptions(workspace.UpdateOptions{
				ArchiveLegacy: archiveLegacy,
			})
			if err != nil {
				return err
			}
			fmt.Printf("Updated plan workspace at %s\n", result.Info.PlanDir)
			for _, item := range result.Created {
				fmt.Printf("  created %s\n", item)
			}
			for _, item := range result.Updated {
				fmt.Printf("  updated %s\n", item)
			}
			for _, move := range result.Archived {
				fmt.Printf("  archived %s -> %s\n", move.From, move.To)
			}
			if len(result.Created) == 0 && len(result.Updated) == 0 && len(result.Archived) == 0 {
				fmt.Println("  no changes")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&archiveLegacy, "archive-legacy", false, "move legacy epic/story hierarchy into .plan/archive/ and keep active specs in place")
	return cmd
}
