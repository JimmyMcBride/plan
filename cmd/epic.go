package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newEpicCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epic",
		Short: "Manage epics",
	}

	create := &cobra.Command{
		Use:   "create <title>",
		Short: "Create an epic and its draft spec",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundle, err := planningManager().CreateEpic(args[0], "")
			if err != nil {
				return err
			}
			fmt.Printf("Created epic %s\n", bundle.Epic.Path)
			fmt.Printf("Created spec %s\n", bundle.Spec.Path)
			return nil
		},
	}

	promote := &cobra.Command{
		Use:   "promote <brainstorm-slug>",
		Short: "Promote a brainstorm into an epic and seeded spec",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundle, err := planningManager().PromoteBrainstorm(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Created epic %s\n", bundle.Epic.Path)
			fmt.Printf("Created seeded spec %s\n", bundle.Spec.Path)
			return nil
		},
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "List epics",
		RunE: func(cmd *cobra.Command, args []string) error {
			epics, err := planningManager().ListEpics()
			if err != nil {
				return err
			}
			if len(epics) == 0 {
				fmt.Println("No epics found.")
				return nil
			}
			for _, epic := range epics {
				fmt.Printf("%s [%s] (%d/%d stories done)\n",
					epic.Title,
					epic.SpecStatus,
					epic.DoneStories,
					epic.TotalStories,
				)
			}
			return nil
		},
	}

	show := &cobra.Command{
		Use:   "show <epic-slug>",
		Short: "Show an epic note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := planningManager().ReadEpic(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("%s\n\n%s", note.Path, note.Content)
			return nil
		},
	}

	cmd.AddCommand(create, promote, list, show)
	return cmd
}
