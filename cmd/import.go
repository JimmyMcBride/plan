package cmd

import (
	"fmt"
	"io"

	"plan/internal/planning"

	"github.com/spf13/cobra"
)

func newImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import planning material from external local workspaces",
	}
	cmd.AddCommand(newImportBrainCommand())
	return cmd
}

func newImportBrainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "brain",
		Short: "Inspect or import planning notes from a Brain workspace",
	}

	var workspacePath string
	inspect := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect importable planning notes in a Brain workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			preview, err := planning.InspectBrainWorkspace(workspacePath)
			if err != nil {
				return err
			}
			printBrainImportPreview(cmd.OutOrStdout(), preview)
			return nil
		},
	}
	inspect.Flags().StringVar(&workspacePath, "workspace", ".", "path to a Brain repo root or .brain directory")

	var brainstorms []string
	var epics []string
	var specs []string
	var stories []string
	apply := &cobra.Command{
		Use:   "apply",
		Short: "Import selected planning notes from a Brain workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := planningManager().ImportBrainPlanning(planning.BrainImportSelection{
				WorkspacePath: workspacePath,
				Brainstorms:   brainstorms,
				Epics:         epics,
				Specs:         specs,
				Stories:       stories,
			})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "brain_workspace: %s\n", result.WorkspacePath)
			if len(result.Imported) == 0 {
				fmt.Fprintln(out, "imported: none")
				return nil
			}
			fmt.Fprintf(out, "imported: %d\n", len(result.Imported))
			for _, item := range result.Imported {
				fmt.Fprintf(out, "  - %s %s -> %s\n", item.Type, item.SourcePath, item.DestinationPath)
			}
			fmt.Fprintln(out, "review: inspect imported notes before execution work")
			return nil
		},
	}
	apply.Flags().StringVar(&workspacePath, "workspace", ".", "path to a Brain repo root or .brain directory")
	apply.Flags().StringArrayVar(&brainstorms, "brainstorm", nil, "brainstorm slug to import; repeatable")
	apply.Flags().StringArrayVar(&epics, "epic", nil, "epic slug to import; repeatable")
	apply.Flags().StringArrayVar(&specs, "spec", nil, "spec slug to import; repeatable")
	apply.Flags().StringArrayVar(&stories, "story", nil, "story slug to import; repeatable")

	cmd.AddCommand(inspect, apply)
	return cmd
}

func printBrainImportPreview(out io.Writer, preview *planning.BrainImportPreview) {
	fmt.Fprintf(out, "brain_workspace: %s\n", preview.WorkspacePath)
	printBrainImportCandidates(out, "brainstorms", preview.Brainstorms)
	printBrainImportCandidates(out, "epics", preview.Epics)
	printBrainImportCandidates(out, "specs", preview.Specs)
	printBrainImportCandidates(out, "stories", preview.Stories)
}

func printBrainImportCandidates(out io.Writer, label string, items []planning.BrainImportCandidate) {
	fmt.Fprintf(out, "%s: %d\n", label, len(items))
	for _, item := range items {
		fmt.Fprintf(out, "  - %s %s (%s)\n", item.Type, item.Slug, item.Path)
	}
}
