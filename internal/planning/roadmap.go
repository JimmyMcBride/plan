package planning

import (
	"os"
	"path/filepath"
	"strings"
)

type Roadmap struct {
	Path          string
	Overview      string
	Versions      []RoadmapVersion
	OrderingNotes []string
	ParkingLot    []string
}

type RoadmapVersion struct {
	Key     string
	Title   string
	Goal    string
	Summary []string
	Epics   []RoadmapEpic
}

type RoadmapEpic struct {
	Title string
	Done  bool
}

func (m *Manager) ReadRoadmap() (*Roadmap, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(info.RoadmapFile)
	if err != nil {
		return nil, err
	}
	roadmap := ParseRoadmap(rel(info.ProjectDir, info.RoadmapFile), string(raw))
	return roadmap, nil
}

func ParseRoadmap(path, body string) *Roadmap {
	sections := splitRoadmapSections(body)
	roadmap := &Roadmap{Path: filepath.ToSlash(path)}
	for _, section := range sections {
		switch {
		case strings.EqualFold(section.heading, "Overview"):
			roadmap.Overview = strings.TrimSpace(strings.Join(section.lines, "\n"))
		case strings.EqualFold(section.heading, "Ordering Notes"):
			roadmap.OrderingNotes = collectLooseList(section.lines)
		case strings.EqualFold(section.heading, "Parking Lot"):
			roadmap.ParkingLot = collectLooseList(section.lines)
		case isRoadmapVersionHeading(section.heading):
			roadmap.Versions = append(roadmap.Versions, parseRoadmapVersion(section.heading, section.lines))
		}
	}
	return roadmap
}

type roadmapSection struct {
	heading string
	lines   []string
}

func splitRoadmapSections(body string) []roadmapSection {
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	var sections []roadmapSection
	var current *roadmapSection
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			sections = appendCurrentRoadmapSection(sections, current)
			current = &roadmapSection{heading: strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))}
			continue
		}
		if current == nil {
			continue
		}
		current.lines = append(current.lines, line)
	}
	return appendCurrentRoadmapSection(sections, current)
}

func appendCurrentRoadmapSection(sections []roadmapSection, current *roadmapSection) []roadmapSection {
	if current == nil {
		return sections
	}
	copy := roadmapSection{
		heading: current.heading,
		lines:   append([]string(nil), current.lines...),
	}
	return append(sections, copy)
}

func isRoadmapVersionHeading(heading string) bool {
	heading = strings.TrimSpace(strings.ToLower(heading))
	return strings.HasPrefix(heading, "v") && strings.Contains(heading, ":")
}

func parseRoadmapVersion(heading string, lines []string) RoadmapVersion {
	version := RoadmapVersion{}
	parts := strings.SplitN(heading, ":", 2)
	version.Key = strings.TrimSpace(parts[0])
	if len(parts) == 2 {
		version.Title = strings.TrimSpace(parts[1])
	}

	inSummary := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "Goal:") {
			version.Goal = strings.TrimSpace(strings.TrimPrefix(trimmed, "Goal:"))
			inSummary = false
			continue
		}
		if trimmed == "Summary:" {
			inSummary = true
			continue
		}
		if epic, ok := parseRoadmapEpic(trimmed); ok {
			version.Epics = append(version.Epics, epic)
			inSummary = false
			continue
		}
		if bullet, ok := parseLooseBullet(trimmed); ok {
			if inSummary {
				version.Summary = append(version.Summary, bullet)
			}
			continue
		}
		if inSummary {
			version.Summary = append(version.Summary, trimmed)
		}
	}

	return version
}

func parseRoadmapEpic(line string) (RoadmapEpic, bool) {
	switch {
	case strings.HasPrefix(line, "- [ ] "):
		return RoadmapEpic{Title: strings.TrimSpace(strings.TrimPrefix(line, "- [ ] ")), Done: false}, true
	case strings.HasPrefix(strings.ToLower(line), "- [x] "):
		return RoadmapEpic{Title: strings.TrimSpace(line[6:]), Done: true}, true
	default:
		return RoadmapEpic{}, false
	}
}

func collectLooseList(lines []string) []string {
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if item, ok := parseLooseBullet(trimmed); ok {
			out = append(out, item)
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func parseLooseBullet(line string) (string, bool) {
	switch {
	case strings.HasPrefix(line, "- "):
		return strings.TrimSpace(strings.TrimPrefix(line, "- ")), true
	case strings.HasPrefix(line, "* "):
		return strings.TrimSpace(strings.TrimPrefix(line, "* ")), true
	default:
		return "", false
	}
}
