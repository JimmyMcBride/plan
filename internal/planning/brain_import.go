package planning

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"plan/internal/notes"
	"plan/internal/workspace"
)

type BrainImportPreview struct {
	WorkspacePath string
	Brainstorms   []BrainImportCandidate
	Epics         []BrainImportCandidate
	Specs         []BrainImportCandidate
	Stories       []BrainImportCandidate
}

type BrainImportCandidate struct {
	Type  string
	Slug  string
	Title string
	Path  string
}

type BrainImportSelection struct {
	WorkspacePath string
	Brainstorms   []string
	Epics         []string
	Specs         []string
	Stories       []string
}

type BrainImportResult struct {
	WorkspacePath string
	Imported      []BrainImportedArtifact
}

type BrainImportedArtifact struct {
	Type            string
	SourcePath      string
	DestinationPath string
}

func InspectBrainWorkspace(path string) (*BrainImportPreview, error) {
	repoRoot, brainDir, err := resolveBrainWorkspace(path)
	if err != nil {
		return nil, err
	}
	preview := &BrainImportPreview{WorkspacePath: filepath.ToSlash(repoRoot)}

	if preview.Brainstorms, err = inspectBrainNoteDir(repoRoot, filepath.Join(brainDir, "brainstorms"), "brainstorm"); err != nil {
		return nil, err
	}
	planningDir := filepath.Join(brainDir, "planning")
	if preview.Epics, err = inspectBrainNoteDir(repoRoot, filepath.Join(planningDir, "epics"), "epic"); err != nil {
		return nil, err
	}
	if preview.Specs, err = inspectBrainNoteDir(repoRoot, filepath.Join(planningDir, "specs"), "spec"); err != nil {
		return nil, err
	}
	if preview.Stories, err = inspectBrainNoteDir(repoRoot, filepath.Join(planningDir, "stories"), "story"); err != nil {
		return nil, err
	}

	return preview, nil
}

func resolveBrainWorkspace(path string) (string, string, error) {
	if path == "" {
		path = "."
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", "", err
	}
	brainDir := filepath.Join(abs, ".brain")
	if stat, err := os.Stat(brainDir); err == nil && stat.IsDir() {
		return abs, brainDir, nil
	}
	if filepath.Base(abs) == ".brain" {
		if stat, err := os.Stat(abs); err == nil && stat.IsDir() {
			return filepath.Dir(abs), abs, nil
		}
	}
	return "", "", fmt.Errorf("brain workspace not found at %s", abs)
}

func inspectBrainNoteDir(repoRoot, dir, noteType string) ([]BrainImportCandidate, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	candidates := make([]BrainImportCandidate, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		note, err := notes.Read(path)
		if err != nil {
			return nil, err
		}
		if note.Type != noteType {
			continue
		}
		candidates = append(candidates, BrainImportCandidate{
			Type:  noteType,
			Slug:  slugFromPath(path),
			Title: note.Title,
			Path:  rel(repoRoot, path),
		})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Type != candidates[j].Type {
			return candidates[i].Type < candidates[j].Type
		}
		return candidates[i].Title < candidates[j].Title
	})
	return candidates, nil
}

func (m *Manager) ImportBrainPlanning(selection BrainImportSelection) (*BrainImportResult, error) {
	info, err := m.workspace.EnsureInitialized()
	if err != nil {
		return nil, err
	}
	repoRoot, brainDir, err := resolveBrainWorkspace(selection.WorkspacePath)
	if err != nil {
		return nil, err
	}
	result := &BrainImportResult{WorkspacePath: filepath.ToSlash(repoRoot)}

	importGroups := []struct {
		items    []string
		noteType string
		dir      string
	}{
		{items: selection.Brainstorms, noteType: "brainstorm", dir: filepath.Join(brainDir, "brainstorms")},
		{items: selection.Epics, noteType: "epic", dir: filepath.Join(brainDir, "planning", "epics")},
		{items: selection.Specs, noteType: "spec", dir: filepath.Join(brainDir, "planning", "specs")},
		{items: selection.Stories, noteType: "story", dir: filepath.Join(brainDir, "planning", "stories")},
	}
	for _, group := range importGroups {
		for _, slug := range normalizeStoryRefs(group.items) {
			note, err := notes.Read(filepath.Join(group.dir, slug+".md"))
			if err != nil {
				return nil, err
			}
			imported, err := m.importBrainNote(info, repoRoot, note)
			if err != nil {
				return nil, err
			}
			result.Imported = append(result.Imported, imported)
		}
	}
	return result, nil
}

func (m *Manager) importBrainNote(info *workspace.Info, repoRoot string, note *notes.Note) (BrainImportedArtifact, error) {
	slug := slugFromPath(note.Path)
	path, metadata, err := brainImportDestination(info, note, slug, repoRoot)
	if err != nil {
		return BrainImportedArtifact{}, err
	}
	created, err := notes.Create(path, note.Title, note.Type, note.Content, metadata)
	if err != nil {
		return BrainImportedArtifact{}, err
	}
	return BrainImportedArtifact{
		Type:            note.Type,
		SourcePath:      rel(repoRoot, note.Path),
		DestinationPath: rel(info.ProjectDir, created.Path),
	}, nil
}

func brainImportDestination(info *workspace.Info, note *notes.Note, slug, repoRoot string) (string, map[string]any, error) {
	sourcePath := rel(repoRoot, note.Path)
	metadata := map[string]any{
		"project":                info.ProjectName,
		"slug":                   slug,
		"imported_from":          "brain",
		"source_brain_path":      sourcePath,
		"source_brain_workspace": filepath.ToSlash(repoRoot),
	}
	switch note.Type {
	case "brainstorm":
		status := stringValue(note.Metadata["brainstorm_status"])
		if status == "" {
			status = "active"
		}
		metadata["status"] = status
		return filepath.Join(info.BrainstormsDir, slug+".md"), metadata, nil
	case "epic":
		specSlug := stringValue(note.Metadata["spec"])
		if specSlug == "" {
			specSlug = slug
		}
		metadata["spec"] = slugify(specSlug)
		return filepath.Join(info.EpicsDir, slug+".md"), metadata, nil
	case "spec":
		status := stringValue(note.Metadata["status"])
		if _, ok := validSpecStatuses[status]; !ok {
			status = "draft"
		}
		metadata["status"] = status
		metadata["epic"] = slugify(stringValue(note.Metadata["epic"]))
		return filepath.Join(info.SpecsDir, slug+".md"), metadata, nil
	case "story":
		status := stringValue(note.Metadata["status"])
		if _, ok := validStoryStatuses[status]; !ok {
			status = "todo"
		}
		metadata["status"] = status
		metadata["epic"] = slugify(stringValue(note.Metadata["epic"]))
		metadata["spec"] = slugify(stringValue(note.Metadata["spec"]))
		return filepath.Join(info.StoriesDir, slug+".md"), metadata, nil
	default:
		return "", nil, fmt.Errorf("unsupported brain note type %q", note.Type)
	}
}
