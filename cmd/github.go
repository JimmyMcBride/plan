package cmd

import (
	"fmt"
	"strings"

	"plan/internal/planning"

	"github.com/spf13/cobra"
)

func newGitHubCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "github",
		Short: "Manage the GitHub story backend",
	}

	enable := &cobra.Command{
		Use:   "enable",
		Short: "Enable GitHub-backed stories for this repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := planningManager().EnableGitHubBackend()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "github_backend: %s\n", result.Backend)
			fmt.Fprintf(out, "repo: %s\n", result.Repo)
			fmt.Fprintf(out, "default_branch: %s\n", result.DefaultBranch)
			fmt.Fprintf(out, "state: %s\n", result.StatePath)
			return nil
		},
	}

	var updateVisible bool
	reconcile := &cobra.Command{
		Use:   "reconcile",
		Short: "Reconcile GitHub-backed stories after planning changes merge",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := planningManager().ReconcileGitHubStories(planning.GitHubReconcileOptions{
				UpdateVisible: updateVisible,
			})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "repo: %s\n", result.Repo)
			fmt.Fprintf(out, "branch: %s\n", result.CurrentBranch)
			fmt.Fprintf(out, "default_branch: %s\n", result.DefaultBranch)
			fmt.Fprintf(out, "updated_issues: %d\n", len(result.UpdatedIssues))
			if len(result.ReadyStories) > 0 {
				fmt.Fprintf(out, "ready_stories: %s\n", strings.Join(result.ReadyStories, ", "))
			}
			if len(result.BlockedStories) > 0 {
				fmt.Fprintf(out, "blocked_stories: %s\n", strings.Join(result.BlockedStories, ", "))
			}
			return nil
		},
	}
	reconcile.Flags().BoolVar(&updateVisible, "update-visible", false, "update optional GitHub-visible readiness markers while reconciling")

	cmd.AddCommand(enable, reconcile)
	return cmd
}
