package engine

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skillyfiy/internal/model"
)

// AllowedExtensions defines the skill file extensions recognized by Skillyfiy.
var AllowedExtensions = map[string]bool{
	".md":   true,
	".py":   true,
	".js":   true,
	".json": true,
	".sh":   true,
}

// ScanSkills crawls the provided skills root directory and constructs AgentItem entries.
func ScanSkills(skillsDir string) ([]model.AgentItem, error) {
	var items []model.AgentItem

	info, err := os.Stat(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return items, nil
		}
		return nil, fmt.Errorf("failed to stat skills directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("skills path is not a directory: %s", skillsDir)
	}

	absBase, err := filepath.Abs(skillsDir)
	if err != nil {
		absBase = skillsDir
	}

	err = filepath.Walk(absBase, func(path string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil // Skip unreadable paths gracefully
		}

		name := fi.Name()
		if fi.IsDir() {
			// Skip hidden directories and vendor/cache directories
			if strings.HasPrefix(name, ".") && path != absBase {
				return filepath.SkipDir
			}
			if name == "node_modules" || name == "__pycache__" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(name, ".") {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !AllowedExtensions[ext] {
			return nil
		}

		rel, err := filepath.Rel(absBase, path)
		if err != nil {
			rel = filepath.Base(path)
		}

		size := fi.Size()
		tokens := int(size / 4)
		if size > 0 && tokens == 0 {
			tokens = 1
		}

		description, rawPreview, lineCount := extractSkillPreview(path)

		item := model.AgentItem{
			ID:          "skill:" + rel,
			Name:        rel,
			Type:        model.TypeSkill,
			SourcePath:  path,
			Description: description,
			Tokens:      tokens,
			Size:        size,
			Details: map[string]string{
				"Category":  "Agent Skill",
				"Relative":  rel,
				"Path":      path,
				"Extension": ext,
				"Size":      model.FormatBytes(size),
				"Lines":     fmt.Sprintf("%d", lineCount),
				"Tokens":    fmt.Sprintf("~%d", tokens),
				"Modified":  fi.ModTime().Format("2006-01-02 15:04:05"),
			},
			RawPreview: rawPreview,
			Selected:   false,
		}

		items = append(items, item)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("skills walk error: %w", err)
	}

	return items, nil
}

// extractSkillPreview extracts YAML frontmatter description (if any),
// first non-empty lines for description, and the first 25 lines for raw preview.
func extractSkillPreview(path string) (description string, rawPreview string, totalLines int) {
	file, err := os.Open(path)
	if err != nil {
		return "Unable to read skill content", "", 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var previewLines []string
	var descLines []string
	inFrontmatter := false
	frontmatterChecked := false
	frontmatterDesc := ""

	lineIdx := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineIdx++

		if lineIdx <= 25 {
			previewLines = append(previewLines, line)
		}

		trimmed := strings.TrimSpace(line)

		// Check for YAML frontmatter
		if lineIdx == 1 && trimmed == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter {
			if trimmed == "---" {
				inFrontmatter = false
				frontmatterChecked = true
				continue
			}
			if strings.HasPrefix(strings.ToLower(trimmed), "description:") {
				frontmatterDesc = strings.TrimSpace(trimmed[len("description:"):])
				frontmatterDesc = strings.Trim(frontmatterDesc, `"'`)
			}
			continue
		}

		if len(descLines) < 5 && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			descLines = append(descLines, trimmed)
		}
	}

	totalLines = lineIdx

	if frontmatterDesc != "" {
		description = frontmatterDesc
	} else if len(descLines) > 0 {
		description = strings.Join(descLines, " ")
		if len(description) > 160 {
			description = description[:157] + "..."
		}
	} else {
		description = "Agent skill definition file: " + filepath.Base(path)
	}

	rawPreview = strings.Join(previewLines, "\n")
	_ = frontmatterChecked
	return description, rawPreview, totalLines
}


