package cmd

import (
	"fmt"

	"plan/internal/planning"

	"github.com/spf13/cobra"
)

func newReadyCommand() *cobra.Command {
	var versionFilter string
	var epicFilter string
	cmd := &cobra.Command{
		Use:   "ready",
		Short: "Show ready and dependency-blocked stories",
		RunE: func(cmd *cobra.Command, args []string) error {
			work, err := planningManager().ReadyWork()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			filteredReady := filterStoriesForDisplay(work.Ready, versionFilter, epicFilter)
			filteredBlocked := filterBlockedStoriesForDisplay(work.Blocked, versionFilter, epicFilter)
			if versionFilter != "" || epicFilter != "" {
				fmt.Fprintf(out, "filters: %s\n", formatStatusFilters(versionFilter, epicFilter, ""))
			}
			if len(filteredReady) == 0 {
				fmt.Fprintln(out, "ready: none")
			} else {
				fmt.Fprintf(out, "ready: %d\n", len(filteredReady))
				for _, story := range filteredReady {
					fmt.Fprintf(out, "  - %s [%s] epic=%s\n", story.Title, story.Status, story.Epic)
				}
			}
			if len(filteredBlocked) == 0 {
				fmt.Fprintln(out, "blocked: none")
				return nil
			}
			fmt.Fprintf(out, "blocked: %d\n", len(filteredBlocked))
			for _, item := range filteredBlocked {
				fmt.Fprintf(out, "  - %s [%s] epic=%s\n", item.Story.Title, item.Story.Status, item.Story.Epic)
				for _, reason := range item.Reasons {
					fmt.Fprintf(out, "    %s\n", reason)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&versionFilter, "version", "", "filter by target roadmap version")
	cmd.Flags().StringVar(&epicFilter, "epic", "", "filter by epic slug")
	return cmd
}

func filterBlockedStoriesForDisplay(items []planning.BlockedStory, versionFilter, epicFilter string) []planning.BlockedStory {
	var out []planning.BlockedStory
	for _, item := range items {
		if !storyMatchesFilters(item.Story, versionFilter, epicFilter) {
			continue
		}
		out = append(out, item)
	}
	return out
}
