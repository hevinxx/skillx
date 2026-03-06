package installer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hevinxx/private-skill-repository/internal/github"
	"github.com/hevinxx/private-skill-repository/internal/registry"
	"github.com/hevinxx/private-skill-repository/internal/skillrc"
)

// installPathMap maps skill types to their target directories.
var installPathMap = map[string]string{
	"command": ".claude/commands",
	"skill":  ".claude/skills",
	"agent":  ".claude/agents",
}

// Installer handles installing and removing skills.
type Installer struct {
	client     *github.Client
	binaryName string
}

// New creates a new Installer.
func New(client *github.Client, binaryName string) *Installer {
	return &Installer{client: client, binaryName: binaryName}
}

// baseDir returns the base directory for the given scope.
func baseDir(scope string) (string, error) {
	if scope == "global" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return home, nil
	}
	return ".", nil
}

// Install downloads and installs a skill's files.
func (inst *Installer) Install(entry *registry.SkillEntry, scope string) ([]string, error) {
	// Fetch skill.yaml to get file list
	metaPath := entry.Path + "/skill.yaml"
	metaData, err := inst.client.FetchFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("fetching skill metadata: %w", err)
	}
	meta, err := registry.ParseSkillMeta(metaData)
	if err != nil {
		return nil, err
	}

	base, err := baseDir(scope)
	if err != nil {
		return nil, err
	}

	typeDir, ok := installPathMap[meta.Type]
	if !ok {
		return nil, fmt.Errorf("unknown skill type: %s", meta.Type)
	}

	var installed []string
	for _, file := range meta.Files {
		// Fetch the file from the skill repo
		remotePath := entry.Path + "/" + file
		data, err := inst.client.FetchFile(remotePath)
		if err != nil {
			return installed, fmt.Errorf("fetching %s: %w", file, err)
		}

		// Determine local path: <base>/<typeDir>/<skill-name>/<file>
		localPath := filepath.Join(base, typeDir, meta.Name, file)
		if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
			return installed, fmt.Errorf("creating directory for %s: %w", file, err)
		}
		if err := os.WriteFile(localPath, data, 0644); err != nil {
			return installed, fmt.Errorf("writing %s: %w", file, err)
		}
		installed = append(installed, localPath)
	}

	return installed, nil
}

// Remove deletes a skill's installed files.
func (inst *Installer) Remove(name, skillType, scope string) error {
	base, err := baseDir(scope)
	if err != nil {
		return err
	}

	typeDir, ok := installPathMap[skillType]
	if !ok {
		return fmt.Errorf("unknown skill type: %s", skillType)
	}

	skillDir := filepath.Join(base, typeDir, name)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return fmt.Errorf("skill directory not found: %s", skillDir)
	}
	return os.RemoveAll(skillDir)
}

// TrackInstall records the installation in .skillrc.
func TrackInstall(name, skillType, commit, scope, binaryName string) error {
	rcPath, err := skillrc.Path(scope, binaryName)
	if err != nil {
		return err
	}
	rc, err := skillrc.Load(rcPath)
	if err != nil {
		return err
	}
	rc.Add(name, skillType, commit)
	return skillrc.Save(rcPath, rc)
}

// TrackRemove removes the skill from .skillrc.
func TrackRemove(name, scope, binaryName string) error {
	rcPath, err := skillrc.Path(scope, binaryName)
	if err != nil {
		return err
	}
	rc, err := skillrc.Load(rcPath)
	if err != nil {
		return err
	}
	rc.Remove(name)
	return skillrc.Save(rcPath, rc)
}
