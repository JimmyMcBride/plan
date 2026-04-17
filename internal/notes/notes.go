package notes

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Note struct {
	Path     string
	Title    string
	Type     string
	Metadata map[string]any
	Content  string
}

type UpdateInput struct {
	Title    *string
	Body     *string
	Metadata map[string]any
}

func Read(path string) (*Note, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	meta, body, err := parse(raw)
	if err != nil {
		return nil, err
	}
	title, _ := meta["title"].(string)
	noteType, _ := meta["type"].(string)
	delete(meta, "title")
	delete(meta, "type")
	return &Note{
		Path:     filepath.ToSlash(path),
		Title:    title,
		Type:     noteType,
		Metadata: meta,
		Content:  body,
	}, nil
}

func Create(path, title, noteType, body string, metadata map[string]any) (*Note, error) {
	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("note already exists: %s", path)
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if _, ok := metadata["created_at"]; !ok {
		metadata["created_at"] = now
	}
	metadata["updated_at"] = now
	note := &Note{
		Path:     filepath.ToSlash(path),
		Title:    title,
		Type:     noteType,
		Metadata: metadata,
		Content:  strings.TrimRight(body, "\n") + "\n",
	}
	if err := write(note); err != nil {
		return nil, err
	}
	return Read(path)
}

func Update(path string, input UpdateInput) (*Note, error) {
	note, err := Read(path)
	if err != nil {
		return nil, err
	}
	if input.Title != nil {
		note.Title = *input.Title
	}
	if input.Body != nil {
		note.Content = strings.TrimRight(*input.Body, "\n") + "\n"
	}
	if note.Metadata == nil {
		note.Metadata = map[string]any{}
	}
	for key, value := range input.Metadata {
		note.Metadata[key] = value
	}
	note.Metadata["updated_at"] = time.Now().UTC().Format(time.RFC3339)
	if err := write(note); err != nil {
		return nil, err
	}
	return Read(path)
}

func parse(raw []byte) (map[string]any, string, error) {
	content := strings.ReplaceAll(string(raw), "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	meta := map[string]any{}
	if !strings.HasPrefix(content, "---\n") {
		return meta, strings.TrimLeft(content, "\n"), nil
	}
	rest := content[len("---\n"):]
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		return nil, "", fmt.Errorf("invalid frontmatter format")
	}
	header := rest[:idx]
	body := rest[idx+len("\n---\n"):]
	if err := yaml.Unmarshal([]byte(header), &meta); err != nil {
		return nil, "", err
	}
	return meta, strings.TrimLeft(body, "\n"), nil
}

func write(note *Note) error {
	meta := map[string]any{
		"title": note.Title,
		"type":  note.Type,
	}
	for key, value := range note.Metadata {
		meta[key] = value
	}
	keys := make([]string, 0, len(meta))
	for key := range meta {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	ordered := make(map[string]any, len(meta))
	for _, key := range keys {
		ordered[key] = meta[key]
	}
	header, err := yaml.Marshal(ordered)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(header)
	buf.WriteString("---\n\n")
	buf.WriteString(strings.TrimLeft(note.Content, "\n"))
	if !strings.HasSuffix(note.Content, "\n") {
		buf.WriteString("\n")
	}
	return os.WriteFile(note.Path, buf.Bytes(), 0o644)
}

func AppendUnderHeading(content, heading, entry string) string {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	needle := strings.ToLower(strings.TrimSpace(heading))
	inSection := false
	sectionLevel := 0
	inserted := false
	var out []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			level := 0
			for _, ch := range trimmed {
				if ch == '#' {
					level++
					continue
				}
				break
			}
			title := strings.ToLower(strings.TrimSpace(strings.TrimLeft(trimmed, "#")))
			if inSection && level <= sectionLevel && !inserted {
				out = appendEntryBlock(out, entry)
				inserted = true
				inSection = false
			}
			if title == needle {
				inSection = true
				sectionLevel = level
			}
		}
		out = append(out, line)
	}

	if inSection && !inserted {
		out = appendEntryBlock(out, entry)
		inserted = true
	}
	if !inserted {
		if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, "")
		}
		out = append(out, "## "+heading, "", strings.TrimRight(entry, "\n"))
	}
	return strings.Join(out, "\n") + "\n"
}

func SetSection(content, heading, body string) string {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	needle := strings.ToLower(strings.TrimSpace(heading))
	replacement := renderSectionBlock(heading, body)
	var out []string
	replaced := false
	inSection := false
	sectionLevel := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			level := 0
			for _, ch := range trimmed {
				if ch == '#' {
					level++
					continue
				}
				break
			}
			title := strings.ToLower(strings.TrimSpace(strings.TrimLeft(trimmed, "#")))
			if inSection && level <= sectionLevel {
				inSection = false
			}
			if !inSection && title == needle {
				if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
					out = append(out, "")
				}
				out = append(out, replacement...)
				replaced = true
				inSection = true
				sectionLevel = level
				continue
			}
		}
		if inSection {
			continue
		}
		out = append(out, line)
	}

	if !replaced {
		if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, "")
		}
		out = append(out, replacement...)
	}

	return strings.Join(out, "\n") + "\n"
}

func renderSectionBlock(heading, body string) []string {
	lines := []string{"## " + heading, ""}
	body = strings.TrimRight(body, "\n")
	if body == "" {
		return lines
	}
	lines = append(lines, strings.Split(body, "\n")...)
	return lines
}

func appendEntryBlock(lines []string, entry string) []string {
	if len(lines) == 0 || strings.TrimSpace(lines[len(lines)-1]) != "" {
		lines = append(lines, "")
	}
	return append(lines, strings.TrimRight(entry, "\n"))
}

func ExtractSection(content, heading string) string {
	lines := strings.Split(content, "\n")
	needle := strings.ToLower(strings.TrimSpace(heading))
	inSection := false
	level := 0
	var out []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			currentLevel := 0
			for _, ch := range trimmed {
				if ch == '#' {
					currentLevel++
					continue
				}
				break
			}
			title := strings.ToLower(strings.TrimSpace(strings.TrimLeft(trimmed, "#")))
			if inSection && currentLevel <= level {
				break
			}
			if title == needle {
				inSection = true
				level = currentLevel
				continue
			}
		}
		if inSection {
			out = append(out, line)
		}
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
