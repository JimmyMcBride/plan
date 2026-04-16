package cmd

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"plan/internal/planning"

	"github.com/spf13/cobra"
)

func newStatusCommand() *cobra.Command {
	var versionFilter string
	var epicFilter string
	var storyStatusFilter string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show overall planning status",
		RunE: func(cmd *cobra.Command, args []string) error {
			status, err := planningManager().Status()
			if err != nil {
				return err
			}
			if hasStatusFilters(versionFilter, epicFilter, storyStatusFilter) {
				stories, err := planningManager().ListStories("", storyStatusFilter)
				if err != nil {
					return err
				}
				readyWork, err := planningManager().ReadyWork()
				if err != nil {
					return err
				}
				status = filterProjectStatus(status, stories, readyWork, versionFilter, epicFilter)
				fmt.Fprintf(cmd.OutOrStdout(), "filters: %s\n", formatStatusFilters(versionFilter, epicFilter, storyStatusFilter))
			}
			out := cmd.OutOrStdout()
			printStatus(out, status)
			return nil
		},
	}
	cmd.Flags().StringVar(&versionFilter, "version", "", "filter output to a roadmap version such as v2")
	cmd.Flags().StringVar(&epicFilter, "epic", "", "filter output to an epic slug")
	cmd.Flags().StringVar(&storyStatusFilter, "story-status", "", "filter output to stories with this lifecycle status")
	return cmd
}

func printStatus(out io.Writer, status *planning.ProjectStatus) {
	fmt.Fprintf(out, "project: %s\n", status.Project)
	fmt.Fprintf(out, "planning_model: %s\n", status.PlanningModel)
	fmt.Fprintf(out, "stories: %d total, %d done, %d in_progress, %d blocked\n",
		status.TotalStories,
		status.DoneStories,
		status.InProgressStories,
		status.BlockedStories,
	)
	fmt.Fprintf(out, "ready_work: %d ready, %d blocked_by_dependencies\n",
		status.ReadyStories,
		status.DependencyBlocked,
	)
	if hasRoadmapVersionAssignments(status) {
		printVersionStatus(out, status)
		return
	}
	if len(status.Epics) == 0 {
		fmt.Fprintln(out, "epics: none")
		return
	}
	fmt.Fprintln(out, "epics:")
	for _, epic := range status.Epics {
		fmt.Fprintf(out, "  - %s [%s] (%d/%d done, %d in progress, %d blocked)\n",
			epic.Title,
			epic.SpecStatus,
			epic.DoneStories,
			epic.TotalStories,
			epic.InProgressStories,
			epic.BlockedStories,
		)
	}
}

func hasRoadmapVersionAssignments(status *planning.ProjectStatus) bool {
	for _, version := range status.Versions {
		if len(version.Epics) > 0 {
			return true
		}
	}
	return false
}

func printVersionStatus(out io.Writer, status *planning.ProjectStatus) {
	if status.RoadmapPath != "" {
		fmt.Fprintf(out, "roadmap: %s\n", status.RoadmapPath)
	}
	fmt.Fprintln(out, "versions:")
	for _, version := range status.Versions {
		fmt.Fprintf(out, "  %s: %s (%d stories, %d done, %d in_progress, %d blocked)\n",
			version.Key,
			version.Title,
			version.TotalStories,
			version.DoneStories,
			version.InProgressStories,
			version.BlockedStories,
		)
		if version.Goal != "" {
			fmt.Fprintf(out, "    goal: %s\n", version.Goal)
		}
		if len(version.Epics) == 0 {
			fmt.Fprintln(out, "    epics: none")
			continue
		}
		for _, epic := range version.Epics {
			fmt.Fprintf(out, "    - %s [%s] (%d/%d done, %d in progress, %d blocked)\n",
				epic.Title,
				epic.SpecStatus,
				epic.DoneStories,
				epic.TotalStories,
				epic.InProgressStories,
				epic.BlockedStories,
			)
		}
	}
	if len(status.UnassignedEpics) == 0 {
		if status.ParkingLotCount > 0 {
			fmt.Fprintf(out, "parking_lot: %d item(s)\n", status.ParkingLotCount)
		}
		return
	}
	fmt.Fprintln(out, "unassigned_epics:")
	for _, epic := range status.UnassignedEpics {
		fmt.Fprintf(out, "  - %s [%s] (%d/%d done, %d in progress, %d blocked)\n",
			epic.Title,
			epic.SpecStatus,
			epic.DoneStories,
			epic.TotalStories,
			epic.InProgressStories,
			epic.BlockedStories,
		)
	}
	if status.ParkingLotCount > 0 {
		fmt.Fprintf(out, "parking_lot: %d item(s)\n", status.ParkingLotCount)
	}
}

func hasStatusFilters(version, epic, storyStatus string) bool {
	return strings.TrimSpace(version) != "" || strings.TrimSpace(epic) != "" || strings.TrimSpace(storyStatus) != ""
}

func formatStatusFilters(version, epic, storyStatus string) string {
	var parts []string
	if strings.TrimSpace(version) != "" {
		parts = append(parts, "version="+version)
	}
	if strings.TrimSpace(epic) != "" {
		parts = append(parts, "epic="+epic)
	}
	if strings.TrimSpace(storyStatus) != "" {
		parts = append(parts, "story_status="+storyStatus)
	}
	return strings.Join(parts, ", ")
}

func filterProjectStatus(status *planning.ProjectStatus, stories []planning.StoryInfo, ready *planning.ReadyWork, versionFilter, epicFilter string) *planning.ProjectStatus {
	versionFilter = strings.TrimSpace(versionFilter)
	epicFilter = strings.TrimSpace(epicFilter)
	filteredStories := filterStoriesForDisplay(stories, versionFilter, epicFilter)

	copy := *status
	copy.TotalStories = len(filteredStories)
	copy.DoneStories = 0
	copy.InProgressStories = 0
	copy.BlockedStories = 0
	for _, story := range filteredStories {
		switch story.Status {
		case "done":
			copy.DoneStories++
		case "in_progress":
			copy.InProgressStories++
		case "blocked":
			copy.BlockedStories++
		}
	}

	copy.ReadyStories = 0
	copy.DependencyBlocked = 0
	for _, story := range filterStoriesForDisplay(ready.Ready, versionFilter, epicFilter) {
		copy.ReadyStories++
		_ = story
	}
	for _, blocked := range ready.Blocked {
		if storyMatchesFilters(blocked.Story, versionFilter, epicFilter) {
			copy.DependencyBlocked++
		}
	}

	copy.Epics = filterEpicsForDisplay(status.Epics, filteredStories, versionFilter, epicFilter)
	copy.UnassignedEpics = filterEpicsForDisplay(status.UnassignedEpics, filteredStories, versionFilter, epicFilter)
	copy.Versions = filterVersionsForDisplay(status.Versions, filteredStories, versionFilter, epicFilter)
	return &copy
}

func filterStoriesForDisplay(stories []planning.StoryInfo, versionFilter, epicFilter string) []planning.StoryInfo {
	var out []planning.StoryInfo
	for _, story := range stories {
		if !storyMatchesFilters(story, versionFilter, epicFilter) {
			continue
		}
		out = append(out, story)
	}
	return out
}

func storyMatchesFilters(story planning.StoryInfo, versionFilter, epicFilter string) bool {
	if strings.TrimSpace(versionFilter) != "" && story.TargetVersion != strings.TrimSpace(versionFilter) {
		return false
	}
	if strings.TrimSpace(epicFilter) != "" && story.Epic != strings.TrimSpace(epicFilter) {
		return false
	}
	return true
}

func filterEpicsForDisplay(epics []planning.EpicInfo, stories []planning.StoryInfo, versionFilter, epicFilter string) []planning.EpicInfo {
	storyCounts := map[string][]planning.StoryInfo{}
	for _, story := range stories {
		storyCounts[story.Epic] = append(storyCounts[story.Epic], story)
	}
	var out []planning.EpicInfo
	for _, epic := range epics {
		epicSlug := slugFromRelPath(epic.Path)
		if strings.TrimSpace(epicFilter) != "" && epicSlug != strings.TrimSpace(epicFilter) {
			continue
		}
		epicStories := storyCounts[epicSlug]
		if strings.TrimSpace(versionFilter) != "" {
			matchesVersion := false
			for _, story := range epicStories {
				if story.TargetVersion == strings.TrimSpace(versionFilter) {
					matchesVersion = true
					break
				}
			}
			if !matchesVersion {
				continue
			}
		}
		copy := epic
		copy.TotalStories = len(epicStories)
		copy.DoneStories = 0
		copy.InProgressStories = 0
		copy.BlockedStories = 0
		for _, story := range epicStories {
			switch story.Status {
			case "done":
				copy.DoneStories++
			case "in_progress":
				copy.InProgressStories++
			case "blocked":
				copy.BlockedStories++
			}
		}
		out = append(out, copy)
	}
	return out
}

func filterVersionsForDisplay(versions []planning.VersionStatus, stories []planning.StoryInfo, versionFilter, epicFilter string) []planning.VersionStatus {
	var out []planning.VersionStatus
	for _, version := range versions {
		if strings.TrimSpace(versionFilter) != "" && version.Key != strings.TrimSpace(versionFilter) {
			continue
		}
		filteredEpics := filterEpicsForDisplay(version.Epics, stories, version.Key, epicFilter)
		if strings.TrimSpace(versionFilter) == "" && len(filteredEpics) == 0 {
			continue
		}
		copy := version
		copy.Epics = filteredEpics
		copy.TotalStories = 0
		copy.DoneStories = 0
		copy.InProgressStories = 0
		copy.BlockedStories = 0
		for _, epic := range copy.Epics {
			copy.TotalStories += epic.TotalStories
			copy.DoneStories += epic.DoneStories
			copy.InProgressStories += epic.InProgressStories
			copy.BlockedStories += epic.BlockedStories
		}
		out = append(out, copy)
	}
	return out
}

func slugFromRelPath(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
