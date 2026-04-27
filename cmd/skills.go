package cmd

import (
	"fmt"

	"plan/internal/skills"

	"github.com/spf13/cobra"
)

func newSkillsCommand() *cobra.Command {
	var scope string
	var agents []string
	var project string

	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Install Plan skills into global or project-local skill roots",
	}

	install := &cobra.Command{
		Use:   "install",
		Short: "Install the Plan skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := skills.NewInstaller("").Install(skills.InstallRequest{
				Scope:      skills.Scope(scope),
				Agents:     agents,
				ProjectDir: project,
			})
			if err != nil {
				return err
			}
			for _, result := range results {
				fmt.Fprintf(cmd.OutOrStdout(), "%s [%s] %s %s -> %s\n", result.Agent, result.Scope, result.Skill, result.Method, result.Path)
			}
			return nil
		},
	}

	targets := &cobra.Command{
		Use:   "targets",
		Short: "Show skill install targets without writing anything",
		RunE: func(cmd *cobra.Command, args []string) error {
			items, err := skills.NewInstaller("").ResolveTargets(skills.InstallRequest{
				Scope:      skills.Scope(scope),
				Agents:     agents,
				ProjectDir: project,
			})
			if err != nil {
				return err
			}
			for _, item := range items {
				fmt.Fprintf(cmd.OutOrStdout(), "%s [%s] %s\n", item.Agent, item.Scope, item.Path)
			}
			return nil
		},
	}

	for _, sub := range []*cobra.Command{install, targets} {
		sub.Flags().StringVar(&scope, "scope", string(skills.ScopeGlobal), "scope: global, local, or both")
		sub.Flags().StringArrayVarP(&agents, "agent", "a", nil, "target agent name; repeatable")
		sub.Flags().StringVar(&project, "project", ".", "project root used for local installs")
	}

	cmd.AddCommand(install, targets)
	return cmd
}
