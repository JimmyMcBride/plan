package cmd

import (
	"fmt"

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

	cmd.AddCommand(enable)
	return cmd
}
