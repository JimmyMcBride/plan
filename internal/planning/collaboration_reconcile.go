package planning

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"plan/internal/workspace"
)

type promotionIssueCandidate struct {
	issue        *GitHubIssue
	record       *workspace.GitHubPlanningRecord
	recordSource bool
	score        int
	identity     []string
}

type gitHubIssueRelationshipReader interface {
	GetIssueRelationships(projectDir, repo string, issueNumber int) (*GitHubIssueRelationships, error)
}

func (m *Manager) reconcilePromotionDraft(data *collaborationSourceData, draft *PromotionDraft) error {
	if draft == nil {
		return nil
	}
	initializePromotionRelationshipPlan(draft)
	if data == nil || data.source.Mode != CollaborationSourceGitHubDiscussion {
		return nil
	}

	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return err
	}
	context, err := m.github.CurrentContext(info.ProjectDir)
	if err != nil {
		return err
	}
	state, err := m.workspace.ReadGitHubState()
	if err != nil {
		return err
	}
	initiatives, err := m.promotionCandidates(
		info.ProjectDir,
		context.Repo.Repo,
		state,
		"initiative",
		planIssueInitiativeLabel,
		data.source,
	)
	if err != nil {
		return err
	}
	specs, err := m.promotionCandidates(
		info.ProjectDir,
		context.Repo.Repo,
		state,
		"spec",
		planIssueSpecLabel,
		data.source,
	)
	if err != nil {
		return err
	}

	used := map[int]struct{}{}
	var initiativeRelationships *GitHubIssueRelationships
	if draft.ProposedInitiativeIssue != nil {
		existing, identity, err := resolvePromotionIssue(*draft.ProposedInitiativeIssue, initiatives, nil, -1, used)
		if err != nil {
			return err
		}
		if existing != nil {
			assignPromotionIssueIdentity(draft.ProposedInitiativeIssue, existing, identity)
			used[existing.Number] = struct{}{}
			initiativeRelationships, err = m.getIssueRelationships(info.ProjectDir, context.Repo.Repo, existing.Number)
			if err != nil {
				return fmt.Errorf("inspect initiative #%d relationships: %w", existing.Number, err)
			}
		}
	}

	parentPositions := map[int]int{}
	if initiativeRelationships != nil {
		for i, number := range initiativeRelationships.SubIssues {
			parentPositions[number] = i
		}
	}
	for i := range draft.ProposedSpecIssues {
		existing, identity, err := resolvePromotionIssue(
			draft.ProposedSpecIssues[i],
			specs,
			parentPositions,
			i,
			used,
		)
		if err != nil {
			return err
		}
		if existing == nil {
			continue
		}
		assignPromotionIssueIdentity(&draft.ProposedSpecIssues[i], existing, identity)
		used[existing.Number] = struct{}{}
	}

	if err := m.reconcilePromotionMilestone(info.ProjectDir, context, state, data, draft); err != nil {
		return err
	}
	classifyPromotionIssueActions(draft)
	return m.reconcilePromotionRelationships(info.ProjectDir, context.Repo.Repo, draft, initiativeRelationships)
}

func initializePromotionRelationshipPlan(draft *PromotionDraft) {
	draft.RelationshipPlan = nil
	if draft.ProposedInitiativeIssue != nil {
		for _, spec := range draft.ProposedSpecIssues {
			draft.RelationshipPlan = append(draft.RelationshipPlan, PromotionRelationshipPlan{
				Kind:         "parent_sub_issue",
				Action:       PromotionActionCreate,
				IssueTitle:   draft.ProposedInitiativeIssue.Title,
				RelatedTitle: spec.Title,
			})
		}
	}
	for _, spec := range draft.ProposedSpecIssues {
		for _, dependency := range spec.BlockedBy {
			draft.RelationshipPlan = append(draft.RelationshipPlan, PromotionRelationshipPlan{
				Kind:         "blocked_by",
				Action:       PromotionActionCreate,
				IssueTitle:   spec.Title,
				RelatedTitle: dependency,
			})
		}
	}
}

func (m *Manager) promotionCandidates(
	projectDir, repo string,
	state *workspace.GitHubState,
	kind, label string,
	source CollaborationSourceRef,
) ([]promotionIssueCandidate, error) {
	listed, err := m.github.ListIssuesByLabel(projectDir, repo, []string{label})
	if err != nil {
		return nil, err
	}
	byNumber := make(map[int]*GitHubIssue, len(listed))
	for i := range listed {
		copy := listed[i]
		byNumber[copy.Number] = &copy
	}
	if state != nil {
		for _, record := range state.Planning {
			if record.Kind != kind || record.IssueNumber <= 0 || !planningRecordMatchesSource(record, source) {
				continue
			}
			if _, ok := byNumber[record.IssueNumber]; ok {
				continue
			}
			issue, err := m.github.GetIssue(projectDir, repo, record.IssueNumber)
			if err != nil {
				return nil, fmt.Errorf("inspect Plan metadata issue #%d: %w", record.IssueNumber, err)
			}
			byNumber[issue.Number] = issue
		}
	}

	recordByIssue := map[int]*workspace.GitHubPlanningRecord{}
	if state != nil {
		for _, record := range state.Planning {
			copy := record
			if copy.Kind == kind && copy.IssueNumber > 0 {
				recordByIssue[copy.IssueNumber] = &copy
			}
		}
	}
	numbers := make([]int, 0, len(byNumber))
	for number := range byNumber {
		numbers = append(numbers, number)
	}
	sort.Ints(numbers)
	out := make([]promotionIssueCandidate, 0, len(numbers))
	for _, number := range numbers {
		out = append(out, promotionIssueCandidate{
			issue:        byNumber[number],
			record:       recordByIssue[number],
			recordSource: recordByIssue[number] != nil && planningRecordMatchesSource(*recordByIssue[number], source),
		})
	}
	return out, nil
}

func resolvePromotionIssue(
	draft PromotionIssueDraft,
	candidates []promotionIssueCandidate,
	parentPositions map[int]int,
	expectedPosition int,
	used map[int]struct{},
) (*GitHubIssue, []string, error) {
	var ranked []promotionIssueCandidate
	for _, candidate := range candidates {
		if candidate.issue == nil {
			continue
		}
		if _, ok := used[candidate.issue.Number]; ok {
			continue
		}
		score, identity, strong := scorePromotionIssueCandidate(draft, candidate, parentPositions, expectedPosition)
		if !strong {
			continue
		}
		candidate.score = score
		candidate.identity = identity
		ranked = append(ranked, candidate)
	}
	if len(ranked) == 0 {
		return nil, nil, nil
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].score == ranked[j].score {
			return ranked[i].issue.Number < ranked[j].issue.Number
		}
		return ranked[i].score > ranked[j].score
	})
	if len(ranked) > 1 && ranked[0].score == ranked[1].score {
		return nil, nil, fmt.Errorf(
			"ambiguous Plan identity for %s %q: issues #%d and #%d match equally; repair Plan metadata or source links before promotion",
			draft.Kind,
			draft.Title,
			ranked[0].issue.Number,
			ranked[1].issue.Number,
		)
	}
	return ranked[0].issue, ranked[0].identity, nil
}

func scorePromotionIssueCandidate(
	draft PromotionIssueDraft,
	candidate promotionIssueCandidate,
	parentPositions map[int]int,
	expectedPosition int,
) (int, []string, bool) {
	issue := candidate.issue
	if issue == nil {
		return 0, nil, false
	}
	score := 0
	var identity []string
	recordSource := candidate.recordSource
	sourceLink := issueBodyMatchesSource(issue.Body, draft.SourceLinks)
	exactSlug := slugify(issue.Title) == draft.Slug
	similarity := promotionTitleSimilarity(issue.Title, draft.Title)
	planLabel := containsString(issue.Labels, planIssueInitiativeLabel) || containsString(issue.Labels, planIssueSpecLabel)

	if candidate.record != nil && candidate.record.Slug == draft.Slug && recordSource {
		score += 1000
		identity = append(identity, "plan_metadata")
	}
	if recordSource {
		score += 300
		identity = appendIdentity(identity, "source_identity")
	}
	if sourceLink {
		score += 250
		identity = appendIdentity(identity, "source_link")
	}
	if exactSlug {
		score += 200
		identity = appendIdentity(identity, "stable_slug")
	}
	if similarity >= 0.35 {
		score += int(similarity * 100)
		identity = appendIdentity(identity, "title_similarity")
	}
	parentPosition, hasParent := parentPositions[issue.Number]
	if hasParent {
		score += 100
		identity = appendIdentity(identity, "parent_sub_issue")
	}
	positionMatch := hasParent && expectedPosition >= 0 && parentPosition == expectedPosition
	if positionMatch {
		score += 75
		identity = appendIdentity(identity, "parent_sub_issue_order")
	}
	if candidate.record != nil && candidate.record.ParentIssueNumber > 0 {
		score += 50
		identity = appendIdentity(identity, "parent_metadata")
	}
	strong := (recordSource || sourceLink) && (exactSlug || similarity >= 0.35 || positionMatch)
	strong = strong || (exactSlug && planLabel)
	strong = strong || (candidate.record != nil && candidate.record.Slug == draft.Slug && recordSource)
	return score, identity, strong
}

func planningRecordMatchesSource(record workspace.GitHubPlanningRecord, source CollaborationSourceRef) bool {
	if source.Discussion == nil {
		return false
	}
	if record.SourceMode != "" && record.SourceMode != string(CollaborationSourceGitHubDiscussion) {
		return false
	}
	if record.DiscussionNumber > 0 && record.DiscussionNumber == source.Discussion.Number {
		return true
	}
	return strings.TrimSpace(record.DiscussionURL) != "" &&
		strings.EqualFold(strings.TrimSpace(record.DiscussionURL), strings.TrimSpace(source.Discussion.URL))
}

func issueBodyMatchesSource(body string, sourceLinks []string) bool {
	for _, link := range sourceLinks {
		if strings.TrimSpace(link) != "" && strings.Contains(strings.ToLower(body), strings.ToLower(strings.TrimSpace(link))) {
			return true
		}
	}
	return false
}

func assignPromotionIssueIdentity(draft *PromotionIssueDraft, existing *GitHubIssue, identity []string) {
	draft.IssueNumber = existing.Number
	draft.IssueURL = existing.URL
	draft.Identity = strings.Join(dedupeExactStrings(identity), ",")
	draft.existing = existing
}

func (m *Manager) reconcilePromotionMilestone(
	projectDir string,
	context *GitHubContext,
	state *workspace.GitHubState,
	data *collaborationSourceData,
	draft *PromotionDraft,
) error {
	if draft.MilestonePlan == nil {
		return nil
	}
	var existing *GitHubMilestone
	identity := ""
	if draft.ProposedInitiativeIssue != nil && draft.ProposedInitiativeIssue.existing != nil &&
		draft.ProposedInitiativeIssue.existing.Milestone != nil {
		copy := *draft.ProposedInitiativeIssue.existing.Milestone
		existing = &copy
		identity = "initiative_milestone"
	}
	if existing == nil {
		var common *GitHubMilestone
		for i := range draft.ProposedSpecIssues {
			issue := draft.ProposedSpecIssues[i].existing
			if issue == nil || issue.Milestone == nil {
				continue
			}
			if common == nil {
				copy := *issue.Milestone
				common = &copy
				continue
			}
			if common.Number != issue.Milestone.Number {
				return fmt.Errorf("ambiguous milestone identity: reconciled spec issues use milestones #%d and #%d", common.Number, issue.Milestone.Number)
			}
		}
		if common != nil {
			existing = common
			identity = "spec_milestone"
		}
	}
	if existing == nil && state != nil && draft.ProposedInitiativeIssue != nil {
		if record, ok := state.Planning[draft.ProposedInitiativeIssue.Slug]; ok &&
			planningRecordMatchesSource(record, data.source) &&
			strings.TrimSpace(record.MilestoneTitle) != "" {
			found, err := m.github.FindMilestone(projectDir, context.Repo.Repo, record.MilestoneTitle)
			if err != nil {
				return err
			}
			if found != nil {
				existing = found
				identity = "plan_metadata"
			}
		}
	}
	if existing == nil && strings.TrimSpace(data.milestoneTitle) != "" {
		found, err := m.github.FindMilestone(projectDir, context.Repo.Repo, data.milestoneTitle)
		if err != nil {
			return err
		}
		if found != nil {
			existing = found
			identity = "source_identity"
		}
	}
	if existing == nil {
		draft.MilestonePlan.Action = PromotionActionCreate
		draft.MilestonePlan.Create = true
		return nil
	}
	draft.MilestonePlan.Action = PromotionActionReuse
	draft.MilestonePlan.Create = false
	draft.MilestonePlan.Title = existing.Title
	draft.MilestonePlan.Number = existing.Number
	draft.MilestonePlan.URL = strings.TrimRight(context.Repo.RepoURL, "/") + "/milestone/" + fmt.Sprint(existing.Number)
	draft.MilestonePlan.Identity = identity
	draft.MilestonePlan.existing = existing
	return nil
}

func classifyPromotionIssueActions(draft *PromotionDraft) {
	var milestone *GitHubMilestone
	if draft.MilestonePlan != nil {
		milestone = draft.MilestonePlan.existing
	}
	if draft.ProposedInitiativeIssue != nil {
		classifyPromotionIssueAction(draft.ProposedInitiativeIssue, milestone)
	}
	for i := range draft.ProposedSpecIssues {
		classifyPromotionIssueAction(&draft.ProposedSpecIssues[i], milestone)
	}
}

func classifyPromotionIssueAction(draft *PromotionIssueDraft, milestone *GitHubMilestone) {
	if draft == nil || draft.existing == nil {
		draft.Action = PromotionActionCreate
		return
	}
	existing := draft.existing
	unchanged := strings.TrimSpace(existing.Title) == strings.TrimSpace(draft.Title) &&
		strings.TrimSpace(existing.Body) == strings.TrimSpace(draft.Body) &&
		promotionLabelsMatch(existing.Labels, draft.Labels)
	if milestone != nil {
		unchanged = unchanged && existing.Milestone != nil && existing.Milestone.Number == milestone.Number
	}
	if unchanged {
		draft.Action = PromotionActionUnchanged
	} else {
		draft.Action = PromotionActionUpdate
	}
}

func (m *Manager) applyPromotionIssue(
	projectDir, repo string,
	draft *PromotionIssueDraft,
	milestone *GitHubMilestone,
) (*GitHubIssue, error) {
	if draft == nil {
		return nil, fmt.Errorf("promotion issue draft is required")
	}
	if draft.Action == PromotionActionUnchanged || draft.Action == PromotionActionReuse {
		if draft.existing == nil {
			return nil, fmt.Errorf("promotion %s %q is classified %s without an existing issue", draft.Kind, draft.Title, draft.Action)
		}
		copy := *draft.existing
		copy.Labels = append([]string(nil), draft.existing.Labels...)
		return &copy, nil
	}
	input := GitHubIssueInput{
		Title:  draft.Title,
		Body:   draft.Body,
		State:  "open",
		Labels: append([]string(nil), draft.Labels...),
	}
	if milestone != nil {
		input.Milestone = &milestone.Number
	}
	if draft.Action == PromotionActionUpdate {
		if draft.existing == nil {
			return nil, fmt.Errorf("promotion %s %q is classified update without an existing issue", draft.Kind, draft.Title)
		}
		input.State = draft.existing.State
		input.Labels = reconcilePromotionLabels(draft.existing.Labels, draft.Labels)
		return m.github.UpdateIssue(projectDir, repo, draft.existing.Number, input)
	}
	return m.github.CreateIssue(projectDir, repo, input)
}

func (m *Manager) reconcilePromotionRelationships(
	projectDir, repo string,
	draft *PromotionDraft,
	initiativeRelationships *GitHubIssueRelationships,
) error {
	relationships := map[int]*GitHubIssueRelationships{}
	if draft.ProposedInitiativeIssue != nil && draft.ProposedInitiativeIssue.IssueNumber > 0 && initiativeRelationships != nil {
		relationships[draft.ProposedInitiativeIssue.IssueNumber] = initiativeRelationships
	}
	for i := range draft.ProposedSpecIssues {
		number := draft.ProposedSpecIssues[i].IssueNumber
		if number <= 0 {
			continue
		}
		current, err := m.getIssueRelationships(projectDir, repo, number)
		if err != nil {
			return fmt.Errorf("inspect spec #%d relationships: %w", number, err)
		}
		relationships[number] = current
	}
	specByTitle := map[string]*PromotionIssueDraft{}
	for i := range draft.ProposedSpecIssues {
		specByTitle[normalizePromotionTitle(draft.ProposedSpecIssues[i].Title)] = &draft.ProposedSpecIssues[i]
	}
	for i := range draft.RelationshipPlan {
		plan := &draft.RelationshipPlan[i]
		switch plan.Kind {
		case "parent_sub_issue":
			if draft.ProposedInitiativeIssue == nil {
				continue
			}
			spec := specByTitle[normalizePromotionTitle(plan.RelatedTitle)]
			plan.IssueNumber = draft.ProposedInitiativeIssue.IssueNumber
			if spec != nil {
				plan.RelatedIssueNumber = spec.IssueNumber
			}
			if containsInt(relationships[plan.IssueNumber], plan.RelatedIssueNumber, true) {
				plan.Action = PromotionActionReuse
			}
		case "blocked_by":
			spec := specByTitle[normalizePromotionTitle(plan.IssueTitle)]
			dependency := specByTitle[normalizePromotionTitle(plan.RelatedTitle)]
			if spec != nil {
				plan.IssueNumber = spec.IssueNumber
			}
			if dependency != nil {
				plan.RelatedIssueNumber = dependency.IssueNumber
			}
			if containsInt(relationships[plan.IssueNumber], plan.RelatedIssueNumber, false) {
				plan.Action = PromotionActionReuse
			}
		}
	}
	return nil
}

func (m *Manager) getIssueRelationships(projectDir, repo string, issueNumber int) (*GitHubIssueRelationships, error) {
	reader, ok := m.github.(gitHubIssueRelationshipReader)
	if !ok {
		return nil, fmt.Errorf("GitHub client does not support read-only issue relationship inspection")
	}
	return reader.GetIssueRelationships(projectDir, repo, issueNumber)
}

func containsInt(relationships *GitHubIssueRelationships, number int, subIssue bool) bool {
	if relationships == nil || number <= 0 {
		return false
	}
	items := relationships.BlockedBy
	if subIssue {
		items = relationships.SubIssues
	}
	for _, item := range items {
		if item == number {
			return true
		}
	}
	return false
}

func promotionLabelsMatch(existing, desired []string) bool {
	effective := reconcilePromotionLabels(existing, desired)
	if len(effective) != len(existing) {
		return false
	}
	for _, item := range effective {
		if !containsString(existing, item) {
			return false
		}
	}
	return true
}

func reconcilePromotionLabels(existing, desired []string) []string {
	out := make([]string, 0, len(existing)+len(desired))
	for _, label := range existing {
		if label == planIssueInitiativeLabel ||
			label == planIssueSpecLabel ||
			label == planIssueReadyLabel ||
			label == planIssueBlockedLabel {
			continue
		}
		out = append(out, label)
	}
	return mergeLabels(out, desired)
}

func appendIdentity(items []string, item string) []string {
	for _, existing := range items {
		if existing == item {
			return items
		}
	}
	return append(items, item)
}

var promotionTitleTokenPattern = regexp.MustCompile(`[a-z0-9]+`)

func promotionTitleSimilarity(left, right string) float64 {
	leftTokens := promotionTitleTokenPattern.FindAllString(strings.ToLower(left), -1)
	rightTokens := promotionTitleTokenPattern.FindAllString(strings.ToLower(right), -1)
	if len(leftTokens) == 0 || len(rightTokens) == 0 {
		return 0
	}
	leftSet := map[string]struct{}{}
	rightSet := map[string]struct{}{}
	for _, token := range leftTokens {
		leftSet[token] = struct{}{}
	}
	for _, token := range rightTokens {
		rightSet[token] = struct{}{}
	}
	intersection := 0
	for token := range leftSet {
		if _, ok := rightSet[token]; ok {
			intersection++
		}
	}
	return float64(2*intersection) / float64(len(leftSet)+len(rightSet))
}
