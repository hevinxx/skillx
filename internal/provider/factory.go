package provider

import "fmt"

// Supported provider type constants.
const (
	TypeGitHub = "github"
	TypeGitLab = "gitlab"
	TypeGitea  = "gitea"
)

// New creates a Provider based on the given type, host, org, and repo.
func New(providerType, host, org, repo string) (Provider, error) {
	switch providerType {
	case TypeGitHub, "":
		return NewGitHub(host, org, repo)
	case TypeGitLab:
		return NewGitLab(host, org, repo)
	case TypeGitea:
		return NewGitea(host, org, repo)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s (supported: github, gitlab, gitea)", providerType)
	}
}
