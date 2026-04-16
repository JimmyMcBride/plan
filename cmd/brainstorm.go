package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newBrainstormCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "brainstorm",
		Short: "Manage brainstorm notes",
	}

	start := &cobra.Command{
		Use:   "start <topic>",
		Short: "Start a new brainstorm",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := planningManager().CreateBrainstorm(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Created brainstorm %s\n", note.Path)
			return nil
		},
	}

	var ideaBody string
	var ideaStdin bool
	idea := &cobra.Command{
		Use:   "idea <brainstorm-slug>",
		Short: "Append an idea to a brainstorm",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readBody(cmd.InOrStdin(), ideaBody, ideaStdin)
			if err != nil {
				return err
			}
			if body == "" {
				return fmt.Errorf("idea body is required")
			}
			note, err := planningManager().AddIdea(args[0], body)
			if err != nil {
				return err
			}
			fmt.Printf("Updated brainstorm %s\n", note.Path)
			return nil
		},
	}
	idea.Flags().StringVarP(&ideaBody, "body", "b", "", "idea body")
	idea.Flags().BoolVar(&ideaStdin, "stdin", false, "read idea body from stdin")

	show := &cobra.Command{
		Use:   "show <brainstorm-slug>",
		Short: "Show a brainstorm note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := planningManager().ReadBrainstorm(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("%s\n\n%s", note.Path, note.Content)
			return nil
		},
	}

	cmd.AddCommand(start, idea, show)
	return cmd
}
