package registry

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Index represents the top-level structure of index.yaml.
type Index struct {
	Skills []SkillEntry `yaml:"skills"`
}

// SkillEntry is a single entry in the skill catalog.
type SkillEntry struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author"`
	Tags        []string `yaml:"tags"`
	Path        string   `yaml:"path"`
}

// SkillMeta represents the full skill.yaml metadata.
type SkillMeta struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author"`
	Files       []string `yaml:"files"`
	Tags        []string `yaml:"tags"`
}

// ParseIndex parses raw YAML bytes into an Index.
func ParseIndex(data []byte) (*Index, error) {
	var idx Index
	if err := yaml.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parsing index.yaml: %w", err)
	}
	return &idx, nil
}

// ParseSkillMeta parses raw YAML bytes into a SkillMeta.
func ParseSkillMeta(data []byte) (*SkillMeta, error) {
	var meta SkillMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parsing skill.yaml: %w", err)
	}
	return &meta, nil
}

// Find returns the skill entry with the given name, or nil if not found.
func (idx *Index) Find(name string) *SkillEntry {
	for i := range idx.Skills {
		if idx.Skills[i].Name == name {
			return &idx.Skills[i]
		}
	}
	return nil
}

// FilterByType returns entries matching the given type.
func (idx *Index) FilterByType(skillType string) []SkillEntry {
	var result []SkillEntry
	for _, s := range idx.Skills {
		if s.Type == skillType {
			result = append(result, s)
		}
	}
	return result
}

// Search returns entries where query matches name, description, or tags (case-insensitive).
func (idx *Index) Search(query string) []SkillEntry {
	q := strings.ToLower(query)
	var result []SkillEntry
	for _, s := range idx.Skills {
		if strings.Contains(strings.ToLower(s.Name), q) ||
			strings.Contains(strings.ToLower(s.Description), q) ||
			matchTags(s.Tags, q) {
			result = append(result, s)
		}
	}
	return result
}

func matchTags(tags []string, query string) bool {
	for _, t := range tags {
		if strings.Contains(strings.ToLower(t), query) {
			return true
		}
	}
	return false
}
