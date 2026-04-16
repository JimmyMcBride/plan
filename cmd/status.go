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
			out := cmd.OutOrStdout()
			printStatus(out, status)
			return nil
		},
	}
}

func printStatus(out io.Writer, status *planning.ProjectStatus) {
	fmt.Fprintf(out, "project: %s\n", status.Project)
	fmt.Fprintf(out, "planning_model: %s\n", status.PlanningModel)
	fmt.Fprintf(out, "stories: %d total, %d done, %d in_progress, %d blocked\n",
		status.TotalStories,
		status.DoneStories,
		status.InProgressStories,
		status.BlockedStories,
	)
	if hasRoadmapVersionAssignments(status) {
		printVersionStatus(out, status)
		return
	}
	if len(status.Epics) == 0 {
		fmt.Fprintln(out, "epics: none")
		return
	}
	fmt.Fprintln(out, "epics:")
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

func hasRoadmapVersionAssignments(status *planning.ProjectStatus) bool {
	for _, version := range status.Versions {
		if len(version.Epics) > 0 {
			return true
		}
	}
	return false
}

func printVersionStatus(out io.Writer, status *planning.ProjectStatus) {
	if status.RoadmapPath != "" {
		fmt.Fprintf(out, "roadmap: %s\n", status.RoadmapPath)
	}
	fmt.Fprintln(out, "versions:")
	for _, version := range status.Versions {
		fmt.Fprintf(out, "  %s: %s (%d stories, %d done, %d in_progress, %d blocked)\n",
			version.Key,
			version.Title,
			version.TotalStories,
			version.DoneStories,
			version.InProgressStories,
			version.BlockedStories,
		)
		if version.Goal != "" {
			fmt.Fprintf(out, "    goal: %s\n", version.Goal)
		}
		if len(version.Epics) == 0 {
			fmt.Fprintln(out, "    epics: none")
			continue
		}
		for _, epic := range version.Epics {
			fmt.Fprintf(out, "    - %s [%s] (%d/%d done, %d in progress, %d blocked)\n",
				epic.Title,
				epic.SpecStatus,
				epic.DoneStories,
				epic.TotalStories,
				epic.InProgressStories,
				epic.BlockedStories,
			)
		}
	}
	if len(status.UnassignedEpics) == 0 {
		if status.ParkingLotCount > 0 {
			fmt.Fprintf(out, "parking_lot: %d item(s)\n", status.ParkingLotCount)
		}
		return
	}
	fmt.Fprintln(out, "unassigned_epics:")
	for _, epic := range status.UnassignedEpics {
		fmt.Fprintf(out, "  - %s [%s] (%d/%d done, %d in progress, %d blocked)\n",
			epic.Title,
			epic.SpecStatus,
			epic.DoneStories,
			epic.TotalStories,
			epic.InProgressStories,
			epic.BlockedStories,
		)
	}
	if status.ParkingLotCount > 0 {
		fmt.Fprintf(out, "parking_lot: %d item(s)\n", status.ParkingLotCount)
	}
}
