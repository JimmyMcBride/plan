package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newRoadmapCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roadmap",
		Short: "Show or edit ROADMAP.md",
	}

	show := &cobra.Command{
		Use:   "show",
		Short: "Show ROADMAP.md",
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := workspaceManager().EnsureInitialized()
			if err != nil {
				return err
			}
			raw, err := os.ReadFile(info.RoadmapFile)
			if err != nil {
				return err
			}
			fmt.Printf(".plan/ROADMAP.md\n\n%s", string(raw))
			return nil
		},
	}

	var body string
	var useStdin bool
	var editor string
	edit := &cobra.Command{
		Use:   "edit",
		Short: "Edit ROADMAP.md via --body, --stdin, or $EDITOR",
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := workspaceManager().EnsureInitialized()
			if err != nil {
				return err
			}
			current, err := os.ReadFile(info.RoadmapFile)
			if err != nil {
				return err
			}
			updatedBody, err := readBody(cmd.InOrStdin(), body, useStdin)
			if err != nil {
				return err
			}
			if updatedBody == "" && !useStdin {
				updatedBody, err = editTextInEditor(string(current), editor)
				if err != nil {
					return err
				}
			}
			if err := os.WriteFile(info.RoadmapFile, []byte(updatedBody), 0o644); err != nil {
				return err
			}
			fmt.Println("Updated .plan/ROADMAP.md")
			return nil
		},
	}
	edit.Flags().StringVarP(&body, "body", "b", "", "replacement body")
	edit.Flags().BoolVar(&useStdin, "stdin", false, "read replacement body from stdin")
	edit.Flags().StringVar(&editor, "editor", "", "editor command")

	cmd.AddCommand(show, edit)
	return cmd
}
