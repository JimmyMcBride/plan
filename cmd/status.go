package cmd

import (
	"fmt"
	"io"

	"plan/internal/planning"

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
			printStatus(cmd.OutOrStdout(), status)
			return nil
		},
	}
}

func printStatus(out io.Writer, status *planning.ProjectStatus) {
	fmt.Fprintf(out, "project: %s\n", status.Project)
	fmt.Fprintf(out, "planning_model: %s\n", status.PlanningModel)
	fmt.Fprintf(out, "specs: %d total, %d draft, %d approved, %d implementing, %d done\n",
		status.TotalSpecs,
		status.DraftSpecs,
		status.ApprovedSpecs,
		status.ImplementingSpecs,
		status.DoneSpecs,
	)
	if len(status.ReadySpecs) > 0 {
		fmt.Fprintf(out, "ready_specs: %d\n", len(status.ReadySpecs))
		for _, spec := range status.ReadySpecs {
			initiativeRef := ""
			if spec.Initiative != "" {
				initiativeRef = fmt.Sprintf(" initiative=%s", spec.Initiative)
			}
			fmt.Fprintf(out, "  - %s%s status=%s\n", spec.Title, initiativeRef, spec.Status)
		}
	}
	if status.TotalStories > 0 {
		fmt.Fprintf(out, "legacy_stories: %d total, %d done, %d in_progress, %d blocked\n",
			status.TotalStories,
			status.DoneStories,
			status.InProgressStories,
			status.BlockedStories,
		)
	}
	if len(status.Epics) > 0 {
		fmt.Fprintln(out, "legacy_epics:")
		for _, epic := range status.Epics {
			fmt.Fprintf(out, "  - %s [%s] (%d/%d done, %d in progress, %d blocked)\n",
				epic.Title,
				epic.SpecStatus,
				epic.DoneStories,
				epic.TotalStories,
				epic.InProgressStories,
				epic.BlockedStories,
			)
		}
	}
}
