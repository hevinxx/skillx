package skillrc

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// SkillRC represents the .skillrc tracking file.
type SkillRC struct {
	Installed []InstalledSkill `yaml:"installed"`
}

// InstalledSkill tracks a single installed skill.
type InstalledSkill struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Commit      string `yaml:"commit"`
	InstalledAt string `yaml:"installed_at"`
}

// Path returns the .skillrc path for the given scope.
// "project" uses .claude/.skillrc in the current directory.
// "global" uses ~/.config/<binaryName>/global.skillrc.
func Path(scope, binaryName string) (string, error) {
	if scope == "global" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		return filepath.Join(home, ".config", binaryName, "global.skillrc"), nil
	}
	return filepath.Join(".claude", ".skillrc"), nil
}

// Load reads the .skillrc file. Returns an empty SkillRC if the file doesn't exist.
func Load(path string) (*SkillRC, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &SkillRC{}, nil
		}
		return nil, fmt.Errorf("reading .skillrc: %w", err)
	}
	var rc SkillRC
	if err := yaml.Unmarshal(data, &rc); err != nil {
		return nil, fmt.Errorf("parsing .skillrc: %w", err)
	}
	return &rc, nil
}

// Save writes the .skillrc to disk.
func Save(path string, rc *SkillRC) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory for .skillrc: %w", err)
	}
	data, err := yaml.Marshal(rc)
	if err != nil {
		return fmt.Errorf("serializing .skillrc: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// Find returns the installed skill with the given name, or nil.
func (rc *SkillRC) Find(name string) *InstalledSkill {
	for i := range rc.Installed {
		if rc.Installed[i].Name == name {
			return &rc.Installed[i]
		}
	}
	return nil
}

// Add adds or updates an installed skill entry.
func (rc *SkillRC) Add(name, skillType, commit string) {
	now := time.Now().UTC().Format(time.RFC3339)
	for i := range rc.Installed {
		if rc.Installed[i].Name == name {
			rc.Installed[i].Commit = commit
			rc.Installed[i].InstalledAt = now
			return
		}
	}
	rc.Installed = append(rc.Installed, InstalledSkill{
		Name:        name,
		Type:        skillType,
		Commit:      commit,
		InstalledAt: now,
	})
}

// Remove removes an installed skill entry. Returns true if found.
func (rc *SkillRC) Remove(name string) bool {
	for i := range rc.Installed {
		if rc.Installed[i].Name == name {
			rc.Installed = append(rc.Installed[:i], rc.Installed[i+1:]...)
			return true
		}
	}
	return false
}
