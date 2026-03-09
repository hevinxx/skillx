package provider

// Provider defines the interface for interacting with a Git hosting service.
// Implementations exist for GitHub, GitLab, and Gitea.
type Provider interface {
	// FetchFile fetches a single file from the skill repository at the given path.
	FetchFile(path string) ([]byte, error)

	// GetLatestCommit returns the latest commit SHA on the default branch.
	GetLatestCommit() (string, error)

	// HasFileChanged checks if a file has changed between two commits.
	HasFileChanged(path, sinceCommit, untilCommit string) (bool, error)
}
