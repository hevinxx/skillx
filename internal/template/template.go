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

// ciPrefixes maps each provider type to the CI directory prefixes it uses.
var ciPrefixes = map[string][]string{
	"github": {".github/"},
	"gitlab": {".gitlab-ci.yml"},
	"gitea":  {".gitea/"},
}

// allCIPrefixes returns all CI-related prefixes across all providers.
func allCIPrefixes() []string {
	var all []string
	for _, prefixes := range ciPrefixes {
		all = append(all, prefixes...)
	}
	return all
}

// isCIFile checks if a relative path belongs to any provider's CI configuration.
func isCIFile(rel string) bool {
	for _, prefix := range allCIPrefixes() {
		if strings.HasPrefix(rel, prefix) {
			return true
		}
	}
	return false
}

// isCIFileForProvider checks if a relative path belongs to the given provider's CI configuration.
func isCIFileForProvider(rel, providerType string) bool {
	prefixes, ok := ciPrefixes[providerType]
	if !ok {
		return false
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(rel, prefix) {
			return true
		}
	}
	return false
}

// InitRepo scaffolds a new skill repository at the given directory.
// providerType determines which CI workflow files to include ("github", "gitlab", or "gitea").
func InitRepo(dir, providerType string) ([]string, error) {
	var created []string

	entries, err := walkEmbedDir("files")
	if err != nil {
		return nil, fmt.Errorf("reading template files: %w", err)
	}

	for _, entry := range entries {
		// Strip the "files/" prefix to get relative path
		rel := strings.TrimPrefix(entry, "files/")
		// Remove .tmpl suffix if present
		rel = strings.TrimSuffix(rel, ".tmpl")

		// Skip CI files that don't belong to the selected provider
		if isCIFile(rel) && !isCIFileForProvider(rel, providerType) {
			continue
		}

		data, err := templateFS.ReadFile(entry)
		if err != nil {
			return nil, fmt.Errorf("reading template %s: %w", entry, err)
		}

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
