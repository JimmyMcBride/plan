package notes

import (
	"os"
	"strings"
	"testing"
)

func TestAppendUnderHeadingUsesSingleBlankLineForEmptySection(t *testing.T) {
	content := "# Brainstorm\n\n## Ideas\n\n## Notes\n"

	updated := AppendUnderHeading(content, "Ideas", "- First idea")

	if strings.Contains(updated, "## Ideas\n\n\n- First idea") {
		t.Fatalf("expected a single blank line before inserted entry:\n%s", updated)
	}
	if !strings.Contains(updated, "## Ideas\n\n- First idea") {
		t.Fatalf("expected inserted entry under ideas heading:\n%s", updated)
	}
}

func TestReadParsesCRLFFrontmatter(t *testing.T) {
	path := t.TempDir() + "/note.md"
	raw := strings.Join([]string{
		"---",
		"title: Windows Note",
		"type: story",
		"status: todo",
		"---",
		"",
		"## Outcome",
		"",
		"Verify CRLF frontmatter parsing.",
		"",
	}, "\r\n")
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	note, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if note.Title != "Windows Note" {
		t.Fatalf("expected title metadata, got %+v", note)
	}
	if note.Type != "story" {
		t.Fatalf("expected type metadata, got %+v", note)
	}
	if note.Metadata["status"] != "todo" {
		t.Fatalf("expected status metadata, got %+v", note.Metadata)
	}
	if !strings.Contains(note.Content, "Verify CRLF frontmatter parsing.") {
		t.Fatalf("expected preserved note body, got:\n%s", note.Content)
	}
}

func TestSetSectionReplacesExistingSection(t *testing.T) {
	content := strings.Join([]string{
		"# Brainstorm",
		"",
		"## Ideas",
		"",
		"- First idea",
		"",
		"## Notes",
		"",
		"Original notes.",
		"",
	}, "\n")

	updated := SetSection(content, "Ideas", strings.Join([]string{
		"### Candidate Approaches",
		"",
		"- Try a simpler default flow",
	}, "\n"))

	if strings.Contains(updated, "- First idea") {
		t.Fatalf("expected existing section content to be replaced:\n%s", updated)
	}
	if !strings.Contains(updated, "## Ideas\n\n### Candidate Approaches") {
		t.Fatalf("expected replacement section content:\n%s", updated)
	}
	if !strings.Contains(updated, "## Notes\n\nOriginal notes.") {
		t.Fatalf("expected following sections to remain intact:\n%s", updated)
	}
}

func TestSetSectionAppendsMissingSection(t *testing.T) {
	content := strings.Join([]string{
		"# Spec",
		"",
		"## Problem",
		"",
		"Need clearer planning.",
		"",
	}, "\n")

	updated := SetSection(content, "Analysis", strings.Join([]string{
		"### Missing Constraints",
		"",
		"- None.",
	}, "\n"))

	if !strings.Contains(updated, "## Analysis\n\n### Missing Constraints") {
		t.Fatalf("expected missing section to be appended:\n%s", updated)
	}
}
