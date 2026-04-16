package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show overall planning status",
		RunE: func(cmd *cobra.Command, args []string) error {
			status, err := planningManager().Status()
			if err != nil {
				return err
			}
			fmt.Printf("project: %s\n", status.Project)
			fmt.Printf("planning_model: %s\n", status.PlanningModel)
			fmt.Printf("stories: %d total, %d done, %d in_progress, %d blocked\n",
				status.TotalStories,
				status.DoneStories,
				status.InProgressStories,
				status.BlockedStories,
			)
			if len(status.Epics) == 0 {
				fmt.Println("epics: none")
				return nil
			}
			fmt.Println("epics:")
			for _, epic := range status.Epics {
				fmt.Printf("  - %s [%s] (%d/%d stories done)\n",
					epic.Title,
					epic.SpecStatus,
					epic.DoneStories,
					epic.TotalStories,
				)
			}
			return nil
		},
	}
}
