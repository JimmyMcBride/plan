package cmd

import (
	"fmt"

	brainapp "github.com/JimmyMcBride/brain/planning/application"
	"github.com/spf13/cobra"
)

func newDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Inspect local .plan workspace health",
		RunE: func(cmd *cobra.Command, args []string) error {
			run := func() error {
				report, err := workspaceManager().Doctor()
				if err != nil {
					return err
				}
				out := cmd.OutOrStdout()
				fmt.Fprintf(out, "project: %s\n", report.ProjectDir)
				fmt.Fprintf(out, "plan_dir: %s\n", report.PlanDir)
				fmt.Fprintf(out, "initialized: %t\n", report.Initialized)
				if report.PlanningModel != "" {
					fmt.Fprintf(out, "planning_model: %s\n", report.PlanningModel)
				}
				if report.SchemaVersion != 0 {
					fmt.Fprintf(out, "schema_version: %d\n", report.SchemaVersion)
				}
				fmt.Fprintf(out, "workspace_status: %s\n", report.WorkspaceStatus)
				fmt.Fprintf(out, "migration_status: %s\n", report.MigrationStatus)
				if len(report.Missing) > 0 {
					fmt.Fprintln(out, "missing:")
					for _, item := range report.Missing {
						fmt.Fprintf(out, "  - %s\n", item)
					}
				}
				if len(report.Broken) > 0 {
					fmt.Fprintln(out, "broken:")
					for _, item := range report.Broken {
						fmt.Fprintf(out, "  - %s\n", item)
					}
				}
				if len(report.Guidance) > 0 {
					fmt.Fprintln(out, "guidance:")
					for _, item := range report.Guidance {
						fmt.Fprintf(out, "  - %s\n", item)
					}
				}
				return nil
			}
			if handled, err := withSharedPlanning(cmd, "status", false, func(*brainapp.Service) error { return run() }); handled {
				return err
			}
			return run()
		},
	}
}
