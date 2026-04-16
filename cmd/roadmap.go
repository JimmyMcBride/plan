package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"plan/internal/planning"

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

	var versionFilter string
	versions := &cobra.Command{
		Use:   "versions",
		Short: "Show parsed roadmap versions and epics",
		RunE: func(cmd *cobra.Command, args []string) error {
			roadmap, err := planningManager().ReadRoadmap()
			if err != nil {
				return err
			}
			return printRoadmapVersions(cmd.OutOrStdout(), roadmap, versionFilter)
		},
	}
	versions.Flags().StringVar(&versionFilter, "version", "", "filter to a roadmap version key such as v1")

	cmd.AddCommand(show, edit, versions)
	return cmd
}

func printRoadmapVersions(out io.Writer, roadmap *planning.Roadmap, versionFilter string) error {
	fmt.Fprintf(out, "%s\n\n", roadmap.Path)

	filter := strings.TrimSpace(strings.ToLower(versionFilter))
	var matched int
	for _, version := range roadmap.Versions {
		if filter != "" && strings.ToLower(version.Key) != filter {
			continue
		}
		matched++
		fmt.Fprintf(out, "%s: %s\n", version.Key, version.Title)
		if version.Goal != "" {
			fmt.Fprintf(out, "goal: %s\n", version.Goal)
		}
		if len(version.Epics) > 0 {
			fmt.Fprintln(out, "epics:")
			for _, epic := range version.Epics {
				check := " "
				if epic.Done {
					check = "x"
				}
				fmt.Fprintf(out, "  - [%s] %s\n", check, epic.Title)
			}
		}
		if len(version.Summary) > 0 {
			fmt.Fprintln(out, "summary:")
			for _, item := range version.Summary {
				fmt.Fprintf(out, "  - %s\n", item)
			}
		}
		fmt.Fprintln(out)
	}

	if matched > 0 {
		return nil
	}
	if filter != "" {
		return fmt.Errorf("roadmap version %q not found", versionFilter)
	}
	fmt.Fprintln(out, "versions: none")
	return nil
}
