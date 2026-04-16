package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newReadyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ready",
		Short: "Show ready and dependency-blocked stories",
		RunE: func(cmd *cobra.Command, args []string) error {
			work, err := planningManager().ReadyWork()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(work.Ready) == 0 {
				fmt.Fprintln(out, "ready: none")
			} else {
				fmt.Fprintf(out, "ready: %d\n", len(work.Ready))
				for _, story := range work.Ready {
					fmt.Fprintf(out, "  - %s [%s] epic=%s\n", story.Title, story.Status, story.Epic)
				}
			}
			if len(work.Blocked) == 0 {
				fmt.Fprintln(out, "blocked: none")
				return nil
			}
			fmt.Fprintf(out, "blocked: %d\n", len(work.Blocked))
			for _, item := range work.Blocked {
				fmt.Fprintf(out, "  - %s [%s] epic=%s\n", item.Story.Title, item.Story.Status, item.Story.Epic)
				for _, reason := range item.Reasons {
					fmt.Fprintf(out, "    %s\n", reason)
				}
			}
			return nil
		},
	}
}
