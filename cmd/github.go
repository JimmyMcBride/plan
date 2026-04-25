package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
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

	var (
		adoptBrainstorm      string
		adoptDiscussion      string
		adoptIssues          []string
		adoptFormat          string
		adoptProjectDecision string
	)
	adopt := &cobra.Command{
		Use:   "adopt",
		Short: "Adopt existing GitHub planning issues into Plan metadata",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			issueNumbers, err := parseIssueNumbers(adoptIssues)
			if err != nil {
				return err
			}
			result, err := planningManager().AdoptGitHubPromotion(planning.GitHubAdoptInput{
				BrainstormSlug:  adoptBrainstorm,
				DiscussionRef:   adoptDiscussion,
				IssueNumbers:    issueNumbers,
				ProjectDecision: strings.TrimSpace(adoptProjectDecision),
			})
			if err != nil {
				return err
			}
			if strings.TrimSpace(adoptFormat) != "json" {
				return fmt.Errorf("unsupported github adopt output format %q; only json is supported", adoptFormat)
			}
			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", "  ")
			return encoder.Encode(result)
		},
	}
	adopt.Flags().StringVar(&adoptBrainstorm, "brainstorm", "", "local brainstorm slug that produced the promotion draft")
	adopt.Flags().StringVar(&adoptDiscussion, "discussion", "", "GitHub Discussion number or URL that produced the promotion draft")
	adopt.Flags().StringSliceVar(&adoptIssues, "issues", nil, "GitHub issue numbers in draft order, comma-separated or repeated")
	adopt.Flags().StringVar(&adoptFormat, "format", "json", "output format: json")
	adopt.Flags().StringVar(&adoptProjectDecision, "project-decision", "", "project prompt decision for 5+ spec promotions: create or skip")

	cmd.AddCommand(enable, reconcile, adopt)
	return cmd
}

func parseIssueNumbers(values []string) ([]int, error) {
	var out []int
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(strings.TrimPrefix(part, "#"))
			if part == "" {
				continue
			}
			number, err := strconv.Atoi(part)
			if err != nil || number <= 0 {
				return nil, fmt.Errorf("invalid issue number %q", part)
			}
			out = append(out, number)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("github adopt requires --issues")
	}
	return out, nil
}
