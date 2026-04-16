package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Inspect local .plan workspace health",
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := workspaceManager().Doctor()
			if err != nil {
				return err
			}
			fmt.Printf("project: %s\n", report.ProjectDir)
			fmt.Printf("plan_dir: %s\n", report.PlanDir)
			fmt.Printf("initialized: %t\n", report.Initialized)
			if report.PlanningModel != "" {
				fmt.Printf("planning_model: %s\n", report.PlanningModel)
			}
			if report.SchemaVersion != 0 {
				fmt.Printf("schema_version: %d\n", report.SchemaVersion)
			}
			fmt.Printf("workspace_status: %s\n", report.WorkspaceStatus)
			fmt.Printf("migration_status: %s\n", report.MigrationStatus)
			if len(report.Missing) > 0 {
				fmt.Println("missing:")
				for _, item := range report.Missing {
					fmt.Printf("  - %s\n", item)
				}
			}
			if len(report.Broken) > 0 {
				fmt.Println("broken:")
				for _, item := range report.Broken {
					fmt.Printf("  - %s\n", item)
				}
			}
			return nil
		},
	}
}
