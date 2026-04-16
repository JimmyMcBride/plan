package planning

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"plan/internal/notes"
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
