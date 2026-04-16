package planning

import "fmt"

type ReadyWork struct {
	Ready   []StoryInfo
	Blocked []BlockedStory
}

type BlockedStory struct {
	Story   StoryInfo
	Reasons []string
}

func (m *Manager) ReadyWork() (*ReadyWork, error) {
	stories, err := m.ListStories("", "")
	if err != nil {
		return nil, err
	}
	index := make(map[string]StoryInfo, len(stories))
	for _, story := range stories {
		if slug := slugFromPath(story.Path); slug != "" {
			index[slug] = story
		}
	}

	work := &ReadyWork{}
	for _, story := range stories {
		switch story.Status {
		case "done", "in_progress":
			continue
		}

		reasons := readyBlockerReasons(story, index)
		if len(reasons) == 0 {
			work.Ready = append(work.Ready, story)
			continue
		}
		work.Blocked = append(work.Blocked, BlockedStory{
			Story:   story,
			Reasons: reasons,
		})
	}
	return work, nil
}

func readyBlockerReasons(story StoryInfo, index map[string]StoryInfo) []string {
	var reasons []string
	if story.Status == "blocked" {
		reasons = append(reasons, "story status is blocked")
	}
	for _, blocker := range story.Blockers {
		blockerStory, ok := index[slugify(blocker)]
		if !ok {
			reasons = append(reasons, fmt.Sprintf("missing blocker story %q", blocker))
			continue
		}
		if blockerStory.Status != "done" {
			reasons = append(reasons, fmt.Sprintf("blocked by %s [%s]", blocker, blockerStory.Status))
		}
	}
	return reasons
}
