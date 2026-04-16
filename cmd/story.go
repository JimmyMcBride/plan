package cmd

import (
	"fmt"

	"plan/internal/planning"

	"github.com/spf13/cobra"
)

func newStoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "story",
		Short: "Manage stories",
	}

	var body string
	var criteria []string
	var verification []string
	var resources []string
	create := &cobra.Command{
		Use:   "create <epic-slug> <title>",
		Short: "Create a story from an approved spec",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := planningManager().CreateStory(args[0], args[1], body, criteria, verification, resources)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created story %s\n", note.Path)
			return nil
		},
	}
	create.Flags().StringVarP(&body, "body", "b", "", "story description")
	create.Flags().StringArrayVar(&criteria, "criteria", nil, "acceptance criterion; repeatable")
	create.Flags().StringArrayVar(&verification, "verify", nil, "verification step; repeatable")
	create.Flags().StringArrayVar(&resources, "resource", nil, "resource entry; repeatable")

	var status string
	var addCriteria []string
	var addVerification []string
	var addResources []string
	update := &cobra.Command{
		Use:   "update <story-slug>",
		Short: "Update a story",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := planningManager().UpdateStory(args[0], planning.StoryChanges{
				Status:          status,
				AddCriteria:     addCriteria,
				AddVerification: addVerification,
				AddResources:    addResources,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated story %s\n", note.Path)
			return nil
		},
	}
	update.Flags().StringVar(&status, "status", "", "new status: todo, in_progress, blocked, done")
	update.Flags().StringArrayVar(&addCriteria, "criteria", nil, "acceptance criterion to append; repeatable")
	update.Flags().StringArrayVar(&addVerification, "verify", nil, "verification step to append; repeatable")
	update.Flags().StringArrayVar(&addResources, "resource", nil, "resource entry to append; repeatable")

	var epicFilter string
	var statusFilter string
	var versionFilter string
	list := &cobra.Command{
		Use:   "list",
		Short: "List stories",
		RunE: func(cmd *cobra.Command, args []string) error {
			items, err := planningManager().ListStories(epicFilter, statusFilter)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No stories found.")
				return nil
			}
			printed := 0
			for _, item := range items {
				if versionFilter != "" && item.TargetVersion != versionFilter {
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s [%s] epic=%s spec=%s\n", item.Title, item.Status, item.Epic, item.Spec)
				printed++
			}
			if printed == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No stories found.")
			}
			return nil
		},
	}
	list.Flags().StringVar(&epicFilter, "epic", "", "filter by epic slug")
	list.Flags().StringVar(&statusFilter, "status", "", "filter by story status")
	list.Flags().StringVar(&versionFilter, "version", "", "filter by target roadmap version")

	show := &cobra.Command{
		Use:   "show <story-slug>",
		Short: "Show a story note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := planningManager().ReadStory(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n%s", note.Path, note.Content)
			return nil
		},
	}

	cmd.AddCommand(create, update, list, show)
	return cmd
}
