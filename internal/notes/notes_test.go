package notes

import (
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
