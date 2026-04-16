package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a .plan workspace in the project",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := workspaceManager().Init()
			if err != nil {
				return err
			}
			fmt.Printf("Initialized plan workspace at %s\n", result.Info.PlanDir)
			for _, item := range result.Created {
				fmt.Printf("  created %s\n", item)
			}
			return nil
		},
	}
}
