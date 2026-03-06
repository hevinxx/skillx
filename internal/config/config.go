package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the CLI configuration stored in ~/.config/skillx/config.yaml.
type Config struct {
	GitHub   GitHubConfig `yaml:"github"`
	Defaults Defaults     `yaml:"defaults"`
}

type GitHubConfig struct {
	Host string `yaml:"host"`
	Org  string `yaml:"org"`
	Repo string `yaml:"repo"`
}

type Defaults struct {
	Scope string `yaml:"scope"` // "project" or "global"
}

// Dir returns the configuration directory path.
// Uses SKILLX_CONFIG_DIR env var if set, otherwise ~/.config/skillx.
func Dir(binaryName string) (string, error) {
	if d := os.Getenv("SKILLX_CONFIG_DIR"); d != "" {
		return d, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", binaryName), nil
}

// Path returns the full path to the config file.
func Path(binaryName string) (string, error) {
	dir, err := Dir(binaryName)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load reads the config file from disk.
func Load(binaryName string) (*Config, error) {
	p, err := Path(binaryName)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config not found. Run '%s init' first", binaryName)
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// Save writes the config to disk, creating directories as needed.
func Save(binaryName string, cfg *Config) error {
	p, err := Path(binaryName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("serializing config: %w", err)
	}
	if err := os.WriteFile(p, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}
