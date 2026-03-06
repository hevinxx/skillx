package template

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed files/*
var templateFS embed.FS

// InitRepo scaffolds a new skill repository at the given directory.
func InitRepo(dir string) ([]string, error) {
	var created []string

	entries, err := walkEmbedDir("files")
	if err != nil {
		return nil, fmt.Errorf("reading template files: %w", err)
	}

	for _, entry := range entries {
		data, err := templateFS.ReadFile(entry)
		if err != nil {
			return nil, fmt.Errorf("reading template %s: %w", entry, err)
		}

		// Strip the "files/" prefix to get relative path
		rel := strings.TrimPrefix(entry, "files/")
		// Remove .tmpl suffix if present
		rel = strings.TrimSuffix(rel, ".tmpl")

		target := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return nil, fmt.Errorf("creating directory for %s: %w", rel, err)
		}
		if err := os.WriteFile(target, data, 0644); err != nil {
			return nil, fmt.Errorf("writing %s: %w", rel, err)
		}
		created = append(created, rel)
	}

	// Create empty type directories
	for _, d := range []string{"commands", "skills", "agents"} {
		p := filepath.Join(dir, d)
		if err := os.MkdirAll(p, 0755); err != nil {
			return nil, fmt.Errorf("creating %s directory: %w", d, err)
		}
		// Write .gitkeep to keep empty dirs in git
		gitkeep := filepath.Join(p, ".gitkeep")
		if err := os.WriteFile(gitkeep, []byte{}, 0644); err != nil {
			return nil, fmt.Errorf("writing .gitkeep in %s: %w", d, err)
		}
	}

	return created, nil
}

// ScaffoldSkill creates a new skill directory with template files.
func ScaffoldSkill(repoDir, name, skillType string) ([]string, error) {
	typeDir := skillType + "s"
	skillDir := filepath.Join(repoDir, typeDir, name)

	if _, err := os.Stat(skillDir); err == nil {
		return nil, fmt.Errorf("skill directory already exists: %s", skillDir)
	}

	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return nil, fmt.Errorf("creating skill directory: %w", err)
	}

	// Write skill.yaml
	skillYaml := fmt.Sprintf(`name: %s
type: %s
description: ""
author: ""
files:
  - %s.md
tags: []
`, name, skillType, name)

	var created []string

	yamlPath := filepath.Join(skillDir, "skill.yaml")
	if err := os.WriteFile(yamlPath, []byte(skillYaml), 0644); err != nil {
		return nil, fmt.Errorf("writing skill.yaml: %w", err)
	}
	created = append(created, yamlPath)

	// Write placeholder .md
	mdContent := fmt.Sprintf("# %s\n\nDescribe your %s here.\n", name, skillType)
	mdPath := filepath.Join(skillDir, name+".md")
	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		return nil, fmt.Errorf("writing %s.md: %w", name, err)
	}
	created = append(created, mdPath)

	return created, nil
}

// walkEmbedDir recursively lists all files in the embedded filesystem.
func walkEmbedDir(dir string) ([]string, error) {
	var files []string
	entries, err := templateFS.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		path := dir + "/" + e.Name()
		if e.IsDir() {
			sub, err := walkEmbedDir(path)
			if err != nil {
				return nil, err
			}
			files = append(files, sub...)
		} else {
			files = append(files, path)
		}
	}
	return files, nil
}
