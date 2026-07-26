package planning

import (
	"regexp"
	"strconv"
	"strings"

	"plan/internal/notes"
)

type promotionSpecBrief struct {
	Title              string
	Problem            string
	Purpose            string
	Scope              string
	AcceptanceCriteria []string
	Verification       []string
	Dependencies       []string
	Readiness          ReadinessState
	ReadinessNote      string
}

var (
	promotionSpecHeadingPattern = regexp.MustCompile(`(?m)^###\s+Spec\s+([0-9]+)\s+(?:—|-)\s+(.+?)\s*$`)
	promotionBriefFieldPattern  = regexp.MustCompile(`(?im)^(?:#{4}[ \t]+)?(problem|purpose|scope|acceptance criteria|verification|dependencies|readiness|approval note)[ \t]*:?[ \t]*(.*)$`)
	promotionMilestonePattern   = regexp.MustCompile(`(?im)^Target milestone:[ \t]*(?:\*\*)?(.+?)(?:\*\*)?[ \t]*$`)
)

func parsePromotionMap(content string) (map[string]promotionSpecBrief, []string, string) {
	section := strings.TrimSpace(notes.ExtractSection(content, "Promotion map"))
	if section == "" {
		return nil, nil, ""
	}
	milestoneTitle := ""
	if match := promotionMilestonePattern.FindStringSubmatch(section); len(match) == 2 {
		milestoneTitle = strings.TrimSpace(match[1])
	}
	matches := promotionSpecHeadingPattern.FindAllStringSubmatchIndex(section, -1)
	if len(matches) == 0 {
		return nil, nil, milestoneTitle
	}

	blocks := make([]string, 0, len(matches))
	titles := make([]string, 0, len(matches))
	for i, match := range matches {
		title := strings.TrimSpace(section[match[4]:match[5]])
		blockEnd := len(section)
		if i+1 < len(matches) {
			blockEnd = matches[i+1][0]
		}
		blocks = append(blocks, strings.TrimSpace(section[match[1]:blockEnd]))
		titles = append(titles, title)
	}

	out := make(map[string]promotionSpecBrief, len(blocks))
	for i, block := range blocks {
		brief := parsePromotionSpecBrief(titles[i], block, titles)
		out[normalizePromotionTitle(titles[i])] = brief
	}
	return out, titles, milestoneTitle
}

func parsePromotionSpecBrief(title, block string, titles []string) promotionSpecBrief {
	fields := extractPromotionBriefFields(block)
	purpose := strings.TrimSpace(fields["purpose"])
	if purpose == "" {
		purpose = strings.TrimSpace(fields["_lead"])
	}
	problem := strings.TrimSpace(fields["problem"])
	scope := strings.TrimSpace(fields["scope"])
	if scope == "" {
		scope = purpose
	}
	acceptance := markdownBulletItems(fields["acceptance criteria"])
	verification := markdownBulletItems(fields["verification"])
	if len(verification) == 0 {
		verification = append([]string(nil), acceptance...)
	}
	readinessNote := firstNonEmpty(fields["readiness"], fields["approval note"])
	return promotionSpecBrief{
		Title:              title,
		Problem:            problem,
		Purpose:            purpose,
		Scope:              scope,
		AcceptanceCriteria: acceptance,
		Verification:       verification,
		Dependencies:       parsePromotionBriefDependencies(fields["dependencies"], titles),
		Readiness:          parsePromotionBriefReadiness(readinessNote),
		ReadinessNote:      readinessNote,
	}
}

func extractPromotionBriefFields(block string) map[string]string {
	fields := map[string]string{}
	matches := promotionBriefFieldPattern.FindAllStringSubmatchIndex(block, -1)
	if len(matches) == 0 {
		fields["_lead"] = strings.TrimSpace(block)
		return fields
	}
	fields["_lead"] = strings.TrimSpace(block[:matches[0][0]])
	for i, match := range matches {
		key := strings.ToLower(strings.TrimSpace(block[match[2]:match[3]]))
		value := strings.TrimSpace(block[match[4]:match[5]])
		end := len(block)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		tail := strings.TrimSpace(block[match[1]:end])
		fields[key] = strings.TrimSpace(strings.Join(nonEmptyStrings(value, tail), "\n"))
	}
	return fields
}

func parsePromotionBriefDependencies(raw string, titles []string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.Contains(strings.ToLower(raw), "none") {
		return nil
	}
	numberPattern := regexp.MustCompile(`[0-9]+`)
	var dependencies []string
	for _, match := range numberPattern.FindAllString(raw, -1) {
		number, err := strconv.Atoi(match)
		if err == nil && number > 0 && number <= len(titles) {
			dependencies = append(dependencies, titles[number-1])
		}
	}
	for _, title := range titles {
		if strings.Contains(strings.ToLower(raw), strings.ToLower(title)) {
			dependencies = append(dependencies, title)
		}
	}
	return dedupeTitles(dependencies)
}

func promotionBriefDependencyGuesses(titles []string, briefs map[string]promotionSpecBrief) []MaturityDependencyGuess {
	out := make([]MaturityDependencyGuess, 0, len(titles))
	for _, title := range titles {
		brief := briefs[normalizePromotionTitle(title)]
		out = append(out, MaturityDependencyGuess{
			Spec:      title,
			BlockedBy: append([]string(nil), brief.Dependencies...),
		})
	}
	return out
}

func parsePromotionBriefReadiness(raw string) ReadinessState {
	lower := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case strings.Contains(lower, "reapproval"),
		strings.Contains(lower, "unapproved"),
		strings.Contains(lower, "needs-refinement"),
		strings.Contains(lower, "needs refinement"):
		return ReadinessNeedsRefinement
	case strings.Contains(lower, "blocked"):
		return ReadinessBlocked
	case strings.Contains(lower, "approved"), strings.Contains(lower, "ready"):
		return ReadinessReady
	case lower != "":
		return ReadinessClarifying
	default:
		return ""
	}
}

func markdownBulletItems(section string) []string {
	var out []string
	for _, line := range strings.Split(strings.ReplaceAll(section, "\r\n", "\n"), "\n") {
		item := strings.TrimSpace(line)
		if !strings.HasPrefix(item, "- ") && !strings.HasPrefix(item, "* ") {
			continue
		}
		item = strings.TrimSpace(item[2:])
		if strings.HasPrefix(item, "[ ] ") {
			item = strings.TrimSpace(item[4:])
		} else if strings.HasPrefix(strings.ToLower(item), "[x] ") {
			item = strings.TrimSpace(item[4:])
		}
		if item != "" {
			out = append(out, item)
		}
	}
	return dedupeExactStrings(out)
}

func dedupeExactStrings(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func normalizePromotionTitle(title string) string {
	return strings.ToLower(strings.TrimSpace(title))
}

func nonEmptyStrings(items ...string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			out = append(out, strings.TrimSpace(item))
		}
	}
	return out
}
