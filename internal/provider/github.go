package provider

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// GitHubProvider implements Provider for GitHub and GitHub Enterprise.
type GitHubProvider struct {
	httpClient *http.Client
	token      string
	apiBase    string
	org        string
	repo       string
}

// NewGitHub creates a GitHub provider from the given parameters.
func NewGitHub(host, org, repo string) (*GitHubProvider, error) {
	token, err := resolveGitHubToken()
	if err != nil {
		return nil, err
	}

	apiBase := "https://api.github.com"
	if host != "" && host != "github.com" {
		apiBase = fmt.Sprintf("https://%s/api/v3", host)
	}

	return &GitHubProvider{
		httpClient: &http.Client{},
		token:      token,
		apiBase:    apiBase,
		org:        org,
		repo:       repo,
	}, nil
}

func resolveGitHubToken() (string, error) {
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
	return "", fmt.Errorf("no GitHub token found. Set SKILLX_GITHUB_TOKEN, GITHUB_TOKEN, or run 'gh auth login'")
}

func (p *GitHubProvider) doRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	return p.httpClient.Do(req)
}

func (p *GitHubProvider) FetchFile(path string) ([]byte, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", p.apiBase, p.org, p.repo, path)
	resp, err := p.doRequest(url)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetching %s: HTTP %d: %s", path, resp.StatusCode, string(body))
	}

	var cr struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
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

func (p *GitHubProvider) GetLatestCommit() (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits?per_page=1", p.apiBase, p.org, p.repo)
	resp, err := p.doRequest(url)
	if err != nil {
		return "", fmt.Errorf("fetching latest commit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("fetching latest commit: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var commits []struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return "", fmt.Errorf("decoding commits response: %w", err)
	}
	if len(commits) == 0 {
		return "", fmt.Errorf("no commits found in repository")
	}
	return commits[0].SHA, nil
}

func (p *GitHubProvider) HasFileChanged(path, sinceCommit, untilCommit string) (bool, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/compare/%s...%s", p.apiBase, p.org, p.repo, sinceCommit, untilCommit)
	resp, err := p.doRequest(url)
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
