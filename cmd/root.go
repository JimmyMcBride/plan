package cmd

import "github.com/spf13/cobra"

var projectDir string

func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "plan",
		Short: "Local-first planning CLI for AI-assisted software projects",
	}
	root.PersistentFlags().StringVar(&projectDir, "project", ".", "project root")

	root.AddCommand(
		newAdoptCommand(),
		newCheckCommand(),
		newImportCommand(),
		newInitCommand(),
		newDoctorCommand(),
		newUpdateCommand(),
		newBrainstormCommand(),
		newEpicCommand(),
		newSpecCommand(),
		newStoryCommand(),
		newRoadmapCommand(),
		newReadyCommand(),
		newStatusCommand(),
		newSkillsCommand(),
	)
	return root
}
