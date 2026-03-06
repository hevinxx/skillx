package github

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/hevinxx/skillx/internal/config"
)

// Client interacts with the GitHub API to fetch files from the skill repository.
type Client struct {
	httpClient *http.Client
	token      string
	apiBase    string
	org        string
	repo       string
}

// NewClient creates a GitHub client from the given config.
func NewClient(cfg *config.Config) (*Client, error) {
	token, err := resolveToken()
	if err != nil {
		return nil, err
	}

	apiBase := "https://api.github.com"
	if cfg.GitHub.Host != "" && cfg.GitHub.Host != "github.com" {
		apiBase = fmt.Sprintf("https://%s/api/v3", cfg.GitHub.Host)
	}

	return &Client{
		httpClient: &http.Client{},
		token:      token,
		apiBase:    apiBase,
		org:        cfg.GitHub.Org,
		repo:       cfg.GitHub.Repo,
	}, nil
}

// resolveToken finds a GitHub token from environment or gh CLI.
func resolveToken() (string, error) {
	if t := os.Getenv("SKILLX_GITHUB_TOKEN"); t != "" {
		return t, nil
	}
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t, nil
	}
	out, err := exec.Command("gh", "auth", "token").Output()
	if err == nil {
		t := strings.TrimSpace(string(out))
		if t != "" {
			return t, nil
		}
	}
	return "", fmt.Errorf("no GitHub token found. Set SKILLX_GITHUB_TOKEN, GITHUB_TOKEN, or install gh CLI and run 'gh auth login'")
}

// contentsResponse represents the GitHub Contents API response.
type contentsResponse struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

// FetchFile fetches a single file from the skill repository at the given path.
func (c *Client) FetchFile(path string) ([]byte, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.apiBase, c.org, c.repo, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetching %s: HTTP %d: %s", path, resp.StatusCode, string(body))
	}

	var cr contentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, fmt.Errorf("decoding response for %s: %w", path, err)
	}

	if cr.Encoding != "base64" {
		return nil, fmt.Errorf("unexpected encoding for %s: %s", path, cr.Encoding)
	}

	decoded, err := base64.StdEncoding.DecodeString(cr.Content)
	if err != nil {
		return nil, fmt.Errorf("decoding content for %s: %w", path, err)
	}
	return decoded, nil
}

// commitResponse represents a GitHub commit from the commits API.
type commitResponse struct {
	SHA string `json:"sha"`
}

// GetLatestCommit returns the latest commit SHA on the default branch.
func (c *Client) GetLatestCommit() (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits?per_page=1", c.apiBase, c.org, c.repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching latest commit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("fetching latest commit: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var commits []commitResponse
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return "", fmt.Errorf("decoding commits response: %w", err)
	}
	if len(commits) == 0 {
		return "", fmt.Errorf("no commits found in repository")
	}
	return commits[0].SHA, nil
}

// HasFileChanged checks if a file has changed between two commits.
func (c *Client) HasFileChanged(path, sinceCommit, untilCommit string) (bool, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/compare/%s...%s", c.apiBase, c.org, c.repo, sinceCommit, untilCommit)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("comparing commits: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("comparing commits: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var compareResp struct {
		Files []struct {
			Filename string `json:"filename"`
		} `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&compareResp); err != nil {
		return false, fmt.Errorf("decoding compare response: %w", err)
	}

	for _, f := range compareResp.Files {
		if strings.HasPrefix(f.Filename, path) {
			return true, nil
		}
	}
	return false, nil
}
