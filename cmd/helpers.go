package cmd

import (
	"github.com/hevinxx/private-skill-repository/internal/config"
	"github.com/hevinxx/private-skill-repository/internal/github"
	"github.com/hevinxx/private-skill-repository/internal/registry"
)

func loadConfig() (*config.Config, error) {
	return config.Load(buildInfo.BinaryName)
}

func newGitHubClient() (*github.Client, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	return github.NewClient(cfg)
}

func fetchIndex() (*registry.Index, error) {
	client, err := newGitHubClient()
	if err != nil {
		return nil, err
	}
	data, err := client.FetchFile("index.yaml")
	if err != nil {
		return nil, err
	}
	return registry.ParseIndex(data)
}

func entryPath(idx *registry.Index, name string) string {
	entry := idx.Find(name)
	if entry != nil {
		return entry.Path
	}
	return name
}
