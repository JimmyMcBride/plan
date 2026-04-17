package planning

import (
	"slices"
	"strings"

	"plan/internal/notes"
)

func appendSubsection(lines *[]string, heading, body string) {
	if len(*lines) > 0 {
		*lines = append(*lines, "")
	}
	*lines = append(*lines, "### "+heading, "")
	body = strings.TrimSpace(body)
	if body == "" {
		return
	}
	*lines = append(*lines, strings.Split(body, "\n")...)
}

func normalizeBulletList(body string) string {
	var items []string
	for _, line := range strings.Split(strings.TrimSpace(body), "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "))
		if line == "" {
			continue
		}
		items = append(items, "- "+line)
	}
	if len(items) == 0 {
		return ""
	}
	return strings.Join(items, "\n")
}

func extractSubsection(content, parentHeading, childHeading string) string {
	if strings.TrimSpace(parentHeading) == "" {
		return notes.ExtractSection(content, childHeading)
	}
	return notes.ExtractNestedSection(content, parentHeading, childHeading)
}

func replaceNamedSubsection(sectionBody, heading, body string) string {
	names := []string{heading}
	for _, name := range allNamedSubsections(sectionBody) {
		if name != heading {
			names = append(names, name)
		}
	}
	slices.Sort(names)

	sections := map[string]string{heading: strings.TrimSpace(body)}
	for _, name := range allNamedSubsections(sectionBody) {
		if name == heading {
			continue
		}
		sections[name] = notes.ExtractSection(sectionBody, name)
	}
	return renderNamedSubsections(names, sections)
}

func renderNamedSubsections(order []string, sections map[string]string) string {
	var lines []string
	for _, name := range order {
		body := strings.TrimSpace(sections[name])
		if body == "" {
			continue
		}
		appendSubsection(&lines, name, body)
	}
	return strings.Join(lines, "\n")
}

func allNamedSubsections(sectionBody string) []string {
	lines := strings.Split(sectionBody, "\n")
	seen := map[string]struct{}{}
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "### ") {
			continue
		}
		name := strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}
