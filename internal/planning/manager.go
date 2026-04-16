package planning

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"plan/internal/notes"
	"plan/internal/templates"
	"plan/internal/workspace"
)

var validSpecStatuses = map[string]struct{}{
	"draft":        {},
	"approved":     {},
	"implementing": {},
	"done":         {},
}

var validStoryStatuses = map[string]struct{}{
	"todo":        {},
	"in_progress": {},
	"blocked":     {},
	"done":        {},
}

type BrainstormInfo struct {
	Path  string
	Title string
}

type BrainstormCreateInput struct {
	Topic         string
	FocusQuestion string
	Ideas         []string
}

type brainstormSectionSpec struct {
	Heading string
	List    bool
}

type EpicInfo struct {
	Path              string
	Title             string
	Spec              string
	SpecStatus        string
	TargetVersion     string
	SourceBrainstorm  string
	TotalStories      int
	DoneStories       int
	InProgressStories int
	BlockedStories    int
}

type StoryInfo struct {
	Path          string
	Title         string
	Status        string
	Epic          string
	Spec          string
	TargetVersion string
	Blockers      []string
}

type VersionStatus struct {
	Key               string
	Title             string
	Goal              string
	Epics             []EpicInfo
	TotalStories      int
	DoneStories       int
	InProgressStories int
	BlockedStories    int
}

type ProjectStatus struct {
	Project           string
	PlanningModel     string
	RoadmapPath       string
	Versions          []VersionStatus
	UnassignedEpics   []EpicInfo
	ParkingLotCount   int
	Epics             []EpicInfo
	TotalStories      int
	DoneStories       int
	BlockedStories    int
	InProgressStories int
	ReadyStories      int
	DependencyBlocked int
}

type StoryChanges struct {
	Status          string
	AddCriteria     []string
	AddVerification []string
	AddResources    []string
	SetBlockers     []string
}

type EpicBundle struct {
	Epic *notes.Note
	Spec *notes.Note
}

type Manager struct {
	workspace *workspace.Manager
}

func New(workspaceManager *workspace.Manager) *Manager {
	return &Manager{workspace: workspaceManager}
}

func (m *Manager) CreateBrainstorm(topic string) (*notes.Note, error) {
	return m.CreateBrainstormWithInput(BrainstormCreateInput{Topic: topic})
}

func (m *Manager) CreateBrainstormWithInput(input BrainstormCreateInput) (*notes.Note, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Topic) == "" {
		return nil, fmt.Errorf("brainstorm topic is required")
	}
	body, err := templates.Render("brainstorm.md", map[string]any{
		"Title": input.Topic,
		"Now":   time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.FocusQuestion) != "" {
		body = notes.AppendUnderHeading(body, "Focus Question", strings.TrimSpace(input.FocusQuestion))
	}
	if block := formatBrainstormEntry(brainstormIdeasSection, strings.Join(input.Ideas, "\n")); block != "" {
		body = notes.AppendUnderHeading(body, brainstormIdeasSection.Heading, block)
	}

	slug := slugify(input.Topic)
	path := filepath.Join(info.BrainstormsDir, slug+".md")
	note, err := notes.Create(path, input.Topic, "brainstorm", body, map[string]any{
		"slug":    slug,
		"status":  "active",
		"project": info.ProjectName,
	})
	if err != nil {
		return nil, err
	}
	return m.relNote(note, info.ProjectDir), nil
}

func (m *Manager) ReadBrainstorm(slug string) (*notes.Note, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	note, err := notes.Read(filepath.Join(info.BrainstormsDir, slugify(slug)+".md"))
	if err != nil {
		return nil, err
	}
	return m.relNote(note, info.ProjectDir), nil
}

func (m *Manager) AddIdea(brainstormSlug, body string) (*notes.Note, error) {
	return m.AddBrainstormEntry(brainstormSlug, brainstormIdeasSection.Heading, body)
}

func (m *Manager) AddBrainstormEntry(brainstormSlug, section, body string) (*notes.Note, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(info.BrainstormsDir, slugify(brainstormSlug)+".md")
	note, err := notes.Read(path)
	if err != nil {
		return nil, err
	}
	if note.Type != "brainstorm" {
		return nil, fmt.Errorf("%s is not a brainstorm note", note.Path)
	}
	spec, err := resolveBrainstormSection(section)
	if err != nil {
		return nil, err
	}
	entry := formatBrainstormEntry(spec, body)
	if entry == "" {
		return nil, fmt.Errorf("brainstorm entry is required")
	}
	updated, err := notes.Update(path, notes.UpdateInput{
		Body: ptr(notes.AppendUnderHeading(note.Content, spec.Heading, entry)),
	})
	if err != nil {
		return nil, err
	}
	return m.relNote(updated, info.ProjectDir), nil
}

func (m *Manager) CreateEpic(title, sourceBrainstorm string) (*EpicBundle, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	epicSlug := slugify(title)
	epicBody, err := templates.Render("epic.md", map[string]any{
		"Title": title,
		"Now":   time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return nil, err
	}
	epicMeta := map[string]any{
		"project": info.ProjectName,
		"slug":    epicSlug,
		"spec":    epicSlug,
	}
	if sourceBrainstorm != "" {
		epicMeta["source_brainstorm"] = sourceBrainstorm
	}
	epic, err := notes.Create(filepath.Join(info.EpicsDir, epicSlug+".md"), title, "epic", epicBody, epicMeta)
	if err != nil {
		return nil, err
	}
	spec, err := m.createSpecForEpic(info, epic)
	if err != nil {
		return nil, err
	}
	return &EpicBundle{
		Epic: m.relNote(epic, info.ProjectDir),
		Spec: m.relNote(spec, info.ProjectDir),
	}, nil
}

func (m *Manager) PromoteBrainstorm(brainstormSlug string) (*EpicBundle, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	brainstormPath := filepath.Join(info.BrainstormsDir, slugify(brainstormSlug)+".md")
	brainstorm, err := notes.Read(brainstormPath)
	if err != nil {
		return nil, err
	}
	if brainstorm.Type != "brainstorm" {
		return nil, fmt.Errorf("%s is not a brainstorm note", brainstorm.Path)
	}
	bundle, err := m.CreateEpic(brainstorm.Title, rel(info.ProjectDir, brainstormPath))
	if err != nil {
		return nil, err
	}
	epicAbs := filepath.Join(info.ProjectDir, filepath.FromSlash(bundle.Epic.Path))
	epic, err := notes.Read(epicAbs)
	if err != nil {
		return nil, err
	}
	seededEpic, err := m.seedEpicFromBrainstorm(info, epic, brainstorm)
	if err != nil {
		return nil, err
	}
	specAbs := filepath.Join(info.ProjectDir, filepath.FromSlash(bundle.Spec.Path))
	spec, err := notes.Read(specAbs)
	if err != nil {
		return nil, err
	}
	seeded, err := m.seedSpecFromBrainstorm(info, spec, brainstorm)
	if err != nil {
		return nil, err
	}
	if _, err := notes.Update(brainstormPath, notes.UpdateInput{
		Metadata: map[string]any{
			"status": "promoted",
			"epic":   slugFromPath(seededEpic.Path),
			"spec":   slugFromPath(seeded.Path),
		},
	}); err != nil {
		return nil, err
	}
	bundle.Epic = m.relNote(seededEpic, info.ProjectDir)
	bundle.Spec = m.relNote(seeded, info.ProjectDir)
	return bundle, nil
}

func (m *Manager) ListEpics() ([]EpicInfo, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	epicNotes, err := readNotesInDir(info.EpicsDir)
	if err != nil {
		return nil, err
	}
	stories, err := m.ListStories("", "")
	if err != nil {
		return nil, err
	}
	out := make([]EpicInfo, 0, len(epicNotes))
	for _, epic := range epicNotes {
		epicSlug := slugFromPath(epic.Path)
		specSlug := stringValue(epic.Metadata["spec"])
		if specSlug == "" {
			specSlug = epicSlug
		}
		specStatus := "draft"
		targetVersion := ""
		if spec, err := notes.Read(filepath.Join(info.SpecsDir, specSlug+".md")); err == nil {
			if status := stringValue(spec.Metadata["status"]); status != "" {
				specStatus = status
			}
			targetVersion = stringValue(spec.Metadata["target_version"])
		}
		item := EpicInfo{
			Path:             rel(info.ProjectDir, epic.Path),
			Title:            epic.Title,
			Spec:             specSlug,
			SpecStatus:       specStatus,
			TargetVersion:    targetVersion,
			SourceBrainstorm: stringValue(epic.Metadata["source_brainstorm"]),
		}
		var epicStories []StoryInfo
		for _, story := range stories {
			if story.Epic != epicSlug {
				continue
			}
			epicStories = append(epicStories, story)
			item.TotalStories++
			switch story.Status {
			case "done":
				item.DoneStories++
			case "in_progress":
				item.InProgressStories++
			case "blocked":
				item.BlockedStories++
			}
		}
		item.SpecStatus = deriveSpecStatus(item.SpecStatus, epicStories)
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Title < out[j].Title
	})
	return out, nil
}

func (m *Manager) ReadEpic(epicSlug string) (*notes.Note, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	note, err := notes.Read(filepath.Join(info.EpicsDir, slugify(epicSlug)+".md"))
	if err != nil {
		return nil, err
	}
	return m.relNote(note, info.ProjectDir), nil
}

func (m *Manager) ReadSpec(epicSlug string) (*notes.Note, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	note, err := notes.Read(filepath.Join(info.SpecsDir, m.specSlugForEpic(epicSlug)+".md"))
	if err != nil {
		return nil, err
	}
	return m.relNote(note, info.ProjectDir), nil
}

func (m *Manager) ReadStory(storySlug string) (*notes.Note, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	note, err := notes.Read(filepath.Join(info.StoriesDir, slugify(storySlug)+".md"))
	if err != nil {
		return nil, err
	}
	return m.relNote(note, info.ProjectDir), nil
}

func (m *Manager) UpdateSpec(epicSlug string, input notes.UpdateInput) (*notes.Note, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	if status := stringValue(input.Metadata["status"]); status != "" && !isValidSpecStatus(status) {
		return nil, fmt.Errorf("invalid spec status %q", status)
	}
	note, err := notes.Update(filepath.Join(info.SpecsDir, m.specSlugForEpic(epicSlug)+".md"), input)
	if err != nil {
		return nil, err
	}
	return m.relNote(note, info.ProjectDir), nil
}

func (m *Manager) SetSpecStatus(epicSlug, status string) (*notes.Note, error) {
	if !isValidSpecStatus(status) {
		return nil, fmt.Errorf("invalid spec status %q", status)
	}
	return m.UpdateSpec(epicSlug, notes.UpdateInput{
		Metadata: map[string]any{"status": status},
	})
}

func (m *Manager) CreateStory(epicSlug, title, description string, criteria, verification, resources []string) (*notes.Note, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	if len(trimmedItems(criteria)) == 0 {
		return nil, fmt.Errorf("at least one acceptance criterion is required")
	}
	if len(trimmedItems(verification)) == 0 {
		return nil, fmt.Errorf("at least one verification step is required")
	}
	epic, err := notes.Read(filepath.Join(info.EpicsDir, slugify(epicSlug)+".md"))
	if err != nil {
		return nil, err
	}
	specSlug := m.specSlugFromEpic(epic)
	spec, err := notes.Read(filepath.Join(info.SpecsDir, specSlug+".md"))
	if err != nil {
		return nil, err
	}
	if status := stringValue(spec.Metadata["status"]); status != "approved" {
		if status == "" {
			status = "draft"
		}
		return nil, fmt.Errorf("spec %s is %q; approve the spec before creating stories", rel(info.ProjectDir, spec.Path), status)
	}
	storySlug := slugify(title)
	body, err := templates.Render("story.md", map[string]any{
		"Title": title,
		"Now":   time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return nil, err
	}
	story, err := notes.Create(filepath.Join(info.StoriesDir, storySlug+".md"), title, "story", body, map[string]any{
		"project": info.ProjectName,
		"slug":    storySlug,
		"status":  "todo",
		"epic":    slugFromPath(epic.Path),
		"spec":    specSlug,
	})
	if err != nil {
		return nil, err
	}
	resourceLinks := append([]string{
		fmt.Sprintf("- [Canonical Spec](%s)", relativeLinkPath(filepath.Dir(story.Path), spec.Path)),
	}, resources...)
	body = composeStoryBody(story, StoryChanges{
		AddCriteria:     criteria,
		AddVerification: verification,
		AddResources:    resourceLinks,
	}, description)
	updated, err := notes.Update(story.Path, notes.UpdateInput{
		Body: &body,
	})
	if err != nil {
		return nil, err
	}
	return m.relNote(updated, info.ProjectDir), nil
}

func (m *Manager) UpdateStory(storySlug string, changes StoryChanges) (*notes.Note, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	if changes.Status != "" && !isValidStoryStatus(changes.Status) {
		return nil, fmt.Errorf("invalid story status %q", changes.Status)
	}
	path := filepath.Join(info.StoriesDir, slugify(storySlug)+".md")
	note, err := notes.Read(path)
	if err != nil {
		return nil, err
	}
	body := composeStoryBody(note, changes, "")
	nextStatus := stringValue(note.Metadata["status"])
	if changes.Status != "" {
		nextStatus = changes.Status
	}
	if requiresExecutionExpectations(nextStatus) && !storyBodyHasExecutionExpectations(body) {
		return nil, fmt.Errorf("story %s needs acceptance criteria and verification steps before it can be marked %q", rel(info.ProjectDir, path), nextStatus)
	}

	input := notes.UpdateInput{Body: &body}
	if changes.Status != "" || changes.SetBlockers != nil {
		input.Metadata = map[string]any{}
	}
	if changes.Status != "" {
		input.Metadata["status"] = changes.Status
	}
	if changes.SetBlockers != nil {
		input.Metadata["blockers"] = normalizeStoryRefs(changes.SetBlockers)
	}
	updated, err := notes.Update(path, input)
	if err != nil {
		return nil, err
	}
	if err := m.syncSpecStatusForStory(info, updated); err != nil {
		return nil, err
	}
	return m.relNote(updated, info.ProjectDir), nil
}

func (m *Manager) ListStories(filterEpic, filterStatus string) ([]StoryInfo, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	items, err := readNotesInDir(info.StoriesDir)
	if err != nil {
		return nil, err
	}
	specVersions, err := loadSpecTargetVersions(info)
	if err != nil {
		return nil, err
	}
	filterEpic = slugify(filterEpic)
	out := make([]StoryInfo, 0, len(items))
	for _, story := range items {
		specSlug := stringValue(story.Metadata["spec"])
		item := StoryInfo{
			Path:          rel(info.ProjectDir, story.Path),
			Title:         story.Title,
			Status:        stringValue(story.Metadata["status"]),
			Epic:          stringValue(story.Metadata["epic"]),
			Spec:          specSlug,
			TargetVersion: specVersions[specSlug],
			Blockers:      stringSliceValue(story.Metadata["blockers"]),
		}
		if filterEpic != "" && item.Epic != filterEpic {
			continue
		}
		if filterStatus != "" && item.Status != filterStatus {
			continue
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Title < out[j].Title
	})
	return out, nil
}

func (m *Manager) Status() (*ProjectStatus, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	epics, err := m.ListEpics()
	if err != nil {
		return nil, err
	}
	stories, err := m.ListStories("", "")
	if err != nil {
		return nil, err
	}
	status := &ProjectStatus{
		Project:       info.ProjectName,
		PlanningModel: workspace.PlanningModel,
		Epics:         epics,
		TotalStories:  len(stories),
	}
	for _, story := range stories {
		switch story.Status {
		case "done":
			status.DoneStories++
		case "blocked":
			status.BlockedStories++
		case "in_progress":
			status.InProgressStories++
		}
	}
	if ready, err := m.ReadyWork(); err == nil {
		status.ReadyStories = len(ready.Ready)
		status.DependencyBlocked = len(ready.Blocked)
	}
	if roadmap, err := m.ReadRoadmap(); err == nil {
		status.RoadmapPath = roadmap.Path
		status.ParkingLotCount = len(roadmap.ParkingLot)
		status.Versions, status.UnassignedEpics = buildVersionStatuses(epics, roadmap)
	}
	return status, nil
}

func buildVersionStatuses(epics []EpicInfo, roadmap *Roadmap) ([]VersionStatus, []EpicInfo) {
	versionEpics := make(map[string][]EpicInfo)
	var unassigned []EpicInfo
	for _, epic := range epics {
		key := strings.TrimSpace(epic.TargetVersion)
		if key == "" {
			unassigned = append(unassigned, epic)
			continue
		}
		versionEpics[key] = append(versionEpics[key], epic)
	}

	versions := make([]VersionStatus, 0, len(roadmap.Versions))
	for _, version := range roadmap.Versions {
		status := VersionStatus{
			Key:   strings.TrimSpace(version.Key),
			Title: version.Title,
			Goal:  version.Goal,
			Epics: orderEpicsForVersion(versionEpics[strings.TrimSpace(version.Key)], version),
		}
		accumulateVersionCounts(&status)
		versions = append(versions, status)
		delete(versionEpics, status.Key)
	}

	extraKeys := make([]string, 0, len(versionEpics))
	for key := range versionEpics {
		extraKeys = append(extraKeys, key)
	}
	sort.Strings(extraKeys)
	for _, key := range extraKeys {
		status := VersionStatus{
			Key:   key,
			Title: key,
			Epics: append([]EpicInfo(nil), versionEpics[key]...),
		}
		accumulateVersionCounts(&status)
		versions = append(versions, status)
	}

	return versions, unassigned
}

func orderEpicsForVersion(epics []EpicInfo, version RoadmapVersion) []EpicInfo {
	if len(epics) == 0 {
		return nil
	}
	ordered := make([]EpicInfo, 0, len(epics))
	used := make([]bool, len(epics))
	for _, roadmapEpic := range version.Epics {
		for idx, epic := range epics {
			if used[idx] || epic.Title != roadmapEpic.Title {
				continue
			}
			ordered = append(ordered, epic)
			used[idx] = true
			break
		}
	}
	var extras []EpicInfo
	for idx, epic := range epics {
		if used[idx] {
			continue
		}
		extras = append(extras, epic)
	}
	sort.Slice(extras, func(i, j int) bool {
		return extras[i].Title < extras[j].Title
	})
	return append(ordered, extras...)
}

func accumulateVersionCounts(version *VersionStatus) {
	for _, epic := range version.Epics {
		version.TotalStories += epic.TotalStories
		version.DoneStories += epic.DoneStories
		version.InProgressStories += epic.InProgressStories
		version.BlockedStories += epic.BlockedStories
	}
}

func (m *Manager) createSpecForEpic(info *workspace.Info, epic *notes.Note) (*notes.Note, error) {
	specSlug := m.specSlugFromEpic(epic)
	body, err := templates.Render("spec.md", map[string]any{
		"Title": epic.Title + " Spec",
		"Now":   time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return nil, err
	}
	metadata := map[string]any{
		"project": info.ProjectName,
		"slug":    specSlug,
		"epic":    slugFromPath(epic.Path),
		"status":  "draft",
	}
	if sourceBrainstorm := stringValue(epic.Metadata["source_brainstorm"]); sourceBrainstorm != "" {
		metadata["source_brainstorm"] = sourceBrainstorm
	}
	return notes.Create(filepath.Join(info.SpecsDir, specSlug+".md"), epic.Title+" Spec", "spec", body, metadata)
}

func (m *Manager) seedEpicFromBrainstorm(info *workspace.Info, epic *notes.Note, brainstorm *notes.Note) (*notes.Note, error) {
	body := epic.Content
	if desiredOutcome := notes.ExtractSection(brainstorm.Content, "Desired Outcome"); desiredOutcome != "" {
		body = notes.AppendUnderHeading(body, "Outcome", desiredOutcome)
	}
	if focus := notes.ExtractSection(brainstorm.Content, "Focus Question"); focus != "" {
		body = notes.AppendUnderHeading(body, "Why Now", focus)
	}
	body = notes.AppendUnderHeading(body, "Resources", fmt.Sprintf("- [Source Brainstorm](%s)", relativeLinkPath(filepath.Dir(epic.Path), brainstorm.Path)))
	return notes.Update(epic.Path, notes.UpdateInput{
		Body: ptr(body),
	})
}

func (m *Manager) seedSpecFromBrainstorm(info *workspace.Info, spec *notes.Note, brainstorm *notes.Note) (*notes.Note, error) {
	body := spec.Content
	if focus := notes.ExtractSection(brainstorm.Content, "Focus Question"); focus != "" {
		body = notes.AppendUnderHeading(body, "Problem", focus)
	}
	if desiredOutcome := notes.ExtractSection(brainstorm.Content, "Desired Outcome"); desiredOutcome != "" {
		body = notes.AppendUnderHeading(body, "Goals", desiredOutcome)
	}
	if ideas := notes.ExtractSection(brainstorm.Content, "Ideas"); ideas != "" {
		body = notes.AppendUnderHeading(body, "Goals", ideas)
	}
	if constraints := notes.ExtractSection(brainstorm.Content, "Constraints"); constraints != "" {
		body = notes.AppendUnderHeading(body, "Constraints", constraints)
	}
	if questions := notes.ExtractSection(brainstorm.Content, "Open Questions"); questions != "" {
		body = notes.AppendUnderHeading(body, "Risks / Open Questions", questions)
	}
	body = notes.AppendUnderHeading(body, "Story Breakdown", "- [ ] Split approved spec into execution-ready stories")
	body = notes.AppendUnderHeading(body, "Resources", fmt.Sprintf("- [Source Brainstorm](%s)", relativeLinkPath(filepath.Dir(spec.Path), brainstorm.Path)))
	updated, err := notes.Update(spec.Path, notes.UpdateInput{
		Body: ptr(body),
		Metadata: map[string]any{
			"source_brainstorm": rel(info.ProjectDir, brainstorm.Path),
		},
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func composeStoryBody(note *notes.Note, changes StoryChanges, description string) string {
	body := note.Content
	if strings.TrimSpace(description) != "" {
		body = notes.AppendUnderHeading(body, "Description", strings.TrimSpace(description))
	}
	for _, item := range changes.AddCriteria {
		if strings.TrimSpace(item) == "" {
			continue
		}
		body = notes.AppendUnderHeading(body, "Acceptance Criteria", checklist(item))
	}
	for _, item := range changes.AddVerification {
		if strings.TrimSpace(item) == "" {
			continue
		}
		body = notes.AppendUnderHeading(body, "Verification", bullet(item))
	}
	for _, item := range changes.AddResources {
		if strings.TrimSpace(item) == "" {
			continue
		}
		body = notes.AppendUnderHeading(body, "Resources", bullet(item))
	}
	return body
}

func (m *Manager) specSlugForEpic(epicSlug string) string {
	return slugify(epicSlug)
}

func (m *Manager) specSlugFromEpic(epic *notes.Note) string {
	specSlug := stringValue(epic.Metadata["spec"])
	if specSlug != "" {
		return specSlug
	}
	return slugFromPath(epic.Path)
}

func (m *Manager) relNote(note *notes.Note, projectDir string) *notes.Note {
	copy := *note
	copy.Path = rel(projectDir, note.Path)
	return &copy
}

func readNotesInDir(dir string) ([]*notes.Note, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []*notes.Note
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		note, err := notes.Read(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, note)
	}
	return out, nil
}

func isValidSpecStatus(status string) bool {
	_, ok := validSpecStatuses[status]
	return ok
}

func isValidStoryStatus(status string) bool {
	_, ok := validStoryStatuses[status]
	return ok
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastDash := false
	for _, ch := range value {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			b.WriteRune(ch)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func slugFromPath(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func rel(root, path string) string {
	r, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(r)
}

func relativeLinkPath(fromDir, toFile string) string {
	r, err := filepath.Rel(fromDir, toFile)
	if err != nil {
		return filepath.ToSlash(toFile)
	}
	return filepath.ToSlash(r)
}

func stringValue(v any) string {
	s, _ := v.(string)
	return s
}

func ptr[T any](v T) *T {
	return &v
}

func checklist(item string) string {
	item = strings.TrimSpace(item)
	if strings.HasPrefix(item, "- [ ]") {
		return item
	}
	return "- [ ] " + item
}

func bullet(item string) string {
	item = strings.TrimSpace(item)
	if strings.HasPrefix(item, "- ") || strings.HasPrefix(item, "* ") {
		return item
	}
	return "- " + item
}

var brainstormIdeasSection = brainstormSectionSpec{
	Heading: "Ideas",
	List:    true,
}

func resolveBrainstormSection(value string) (brainstormSectionSpec, error) {
	switch slugify(value) {
	case "", "idea", "ideas":
		return brainstormIdeasSection, nil
	case "focus", "focus-question":
		return brainstormSectionSpec{Heading: "Focus Question"}, nil
	case "desired-outcome", "outcome":
		return brainstormSectionSpec{Heading: "Desired Outcome"}, nil
	case "constraints", "constraint":
		return brainstormSectionSpec{Heading: "Constraints", List: true}, nil
	case "open-questions", "questions", "question":
		return brainstormSectionSpec{Heading: "Open Questions", List: true}, nil
	case "raw-notes", "notes":
		return brainstormSectionSpec{Heading: "Raw Notes"}, nil
	default:
		return brainstormSectionSpec{}, fmt.Errorf("unsupported brainstorm section %q", value)
	}
}

func formatBrainstormEntry(section brainstormSectionSpec, body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	if !section.List {
		return body
	}

	var items []string
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		if line == "" {
			continue
		}
		items = append(items, bullet(line))
	}
	return strings.Join(items, "\n")
}

func requiresExecutionExpectations(status string) bool {
	switch status {
	case "in_progress", "done":
		return true
	default:
		return false
	}
}

func storyBodyHasExecutionExpectations(body string) bool {
	return notes.ExtractSection(body, "Acceptance Criteria") != "" && notes.ExtractSection(body, "Verification") != ""
}

func trimmedItems(items []string) []string {
	var out []string
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func normalizeStoryRefs(items []string) []string {
	var out []string
	seen := map[string]struct{}{}
	for _, item := range items {
		slug := slugify(item)
		if slug == "" {
			continue
		}
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}
		out = append(out, slug)
	}
	return out
}

func stringSliceValue(v any) []string {
	switch value := v.(type) {
	case []string:
		return append([]string(nil), value...)
	case []any:
		var out []string
		for _, item := range value {
			s, ok := item.(string)
			if !ok || strings.TrimSpace(s) == "" {
				continue
			}
			out = append(out, s)
		}
		return out
	default:
		return nil
	}
}

func loadSpecTargetVersions(info *workspace.Info) (map[string]string, error) {
	specs, err := readNotesInDir(info.SpecsDir)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(specs))
	for _, spec := range specs {
		out[slugFromPath(spec.Path)] = stringValue(spec.Metadata["target_version"])
	}
	return out, nil
}

func deriveSpecStatus(current string, stories []StoryInfo) string {
	if len(stories) == 0 {
		return current
	}

	nextStatus := "approved"
	allDone := true
	anyStarted := false
	for _, item := range stories {
		switch item.Status {
		case "done":
			anyStarted = true
		case "in_progress", "blocked":
			anyStarted = true
			allDone = false
		default:
			allDone = false
		}
	}
	if allDone {
		return "done"
	}
	if anyStarted {
		return "implementing"
	}
	return nextStatus
}

func (m *Manager) syncSpecStatusForStory(info *workspace.Info, story *notes.Note) error {
	epicSlug := stringValue(story.Metadata["epic"])
	specSlug := stringValue(story.Metadata["spec"])
	if epicSlug == "" || specSlug == "" {
		return nil
	}

	stories, err := m.ListStories(epicSlug, "")
	if err != nil {
		return err
	}
	if len(stories) == 0 {
		return nil
	}

	specPath := filepath.Join(info.SpecsDir, specSlug+".md")
	spec, err := notes.Read(specPath)
	if err != nil {
		return err
	}
	nextStatus := deriveSpecStatus(stringValue(spec.Metadata["status"]), stories)
	if stringValue(spec.Metadata["status"]) == nextStatus {
		return nil
	}
	_, err = notes.Update(specPath, notes.UpdateInput{
		Metadata: map[string]any{"status": nextStatus},
	})
	return err
}
