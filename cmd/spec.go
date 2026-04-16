package cmd

import (
	"fmt"

	"plan/internal/notes"

	"github.com/spf13/cobra"
)

func newSpecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spec",
		Short: "Manage specs",
	}

	show := &cobra.Command{
		Use:   "show <epic-slug>",
		Short: "Show the canonical spec for an epic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := planningManager().ReadSpec(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("%s\n\n%s", note.Path, note.Content)
			return nil
		},
	}

	var body string
	var useStdin bool
	var editor string
	edit := &cobra.Command{
		Use:   "edit <epic-slug>",
		Short: "Edit a spec via --body, --stdin, or $EDITOR",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := planningManager().ReadSpec(args[0])
			if err != nil {
				return err
			}
			updatedBody, err := readBody(cmd.InOrStdin(), body, useStdin)
			if err != nil {
				return err
			}
			if updatedBody == "" && !useStdin {
				updatedBody, err = editTextInEditor(note.Content, editor)
				if err != nil {
					return err
				}
			}
			updated, err := planningManager().UpdateSpec(args[0], notes.UpdateInput{
				Body: &updatedBody,
			})
			if err != nil {
				return err
			}
			fmt.Printf("Updated spec %s\n", updated.Path)
			return nil
		},
	}
	edit.Flags().StringVarP(&body, "body", "b", "", "replacement body")
	edit.Flags().BoolVar(&useStdin, "stdin", false, "read replacement body from stdin")
	edit.Flags().StringVar(&editor, "editor", "", "editor command")

	var status string
	statusCmd := &cobra.Command{
		Use:   "status <epic-slug>",
		Short: "Set spec status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			updated, err := planningManager().SetSpecStatus(args[0], status)
			if err != nil {
				return err
			}
			fmt.Printf("Set spec %s to %s\n", updated.Path, status)
			return nil
		},
	}
	statusCmd.Flags().StringVar(&status, "set", "", "new status: draft, approved, implementing, done")
	_ = statusCmd.MarkFlagRequired("set")

	cmd.AddCommand(show, edit, statusCmd)
	return cmd
}
