package cmd

import (
	"github.com/hevinxx/skillx/internal/config"
	"github.com/hevinxx/skillx/internal/provider"
	"github.com/hevinxx/skillx/internal/registry"
)

func loadConfig() (*config.Config, error) {
	return config.Load(buildInfo.BinaryName)
}

func newProvider() (provider.Provider, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	return provider.New(cfg.Provider.Type, cfg.Provider.Host, cfg.Provider.Org, cfg.Provider.Repo)
}

func fetchIndex() (*registry.Index, error) {
	client, err := newProvider()
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
