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

// GiteaProvider implements Provider for Gitea and Forgejo instances.
type GiteaProvider struct {
	httpClient *http.Client
	token      string
	apiBase    string
	org        string
	repo       string
}

// NewGitea creates a Gitea provider from the given parameters.
func NewGitea(host, org, repo string) (*GiteaProvider, error) {
	token, err := resolveGiteaToken()
	if err != nil {
		return nil, err
	}

	if host == "" {
		return nil, fmt.Errorf("gitea host is required (e.g. gitea.example.com)")
	}
	apiBase := fmt.Sprintf("https://%s/api/v1", host)

	return &GiteaProvider{
		httpClient: &http.Client{},
		token:      token,
		apiBase:    apiBase,
		org:        org,
		repo:       repo,
	}, nil
}

func resolveGiteaToken() (string, error) {
	if t := os.Getenv("SKILLX_GITEA_TOKEN"); t != "" {
		return t, nil
	}
	if t := os.Getenv("GITEA_TOKEN"); t != "" {
		return t, nil
	}
	out, err := exec.Command("tea", "login", "default", "--output=simple").Output()
	if err == nil {
		t := strings.TrimSpace(string(out))
		if t != "" {
			return t, nil
		}
	}
	return "", fmt.Errorf("no Gitea token found. Set SKILLX_GITEA_TOKEN, GITEA_TOKEN, or configure the tea CLI")
}

func (p *GiteaProvider) doRequest(reqURL string) (*http.Response, error) {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+p.token)
	return p.httpClient.Do(req)
}

func (p *GiteaProvider) FetchFile(path string) ([]byte, error) {
	reqURL := fmt.Sprintf("%s/repos/%s/%s/contents/%s", p.apiBase, p.org, p.repo, path)
	resp, err := p.doRequest(reqURL)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetching %s: HTTP %d: %s", path, resp.StatusCode, string(body))
	}

	var fr struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&fr); err != nil {
		return nil, fmt.Errorf("decoding response for %s: %w", path, err)
	}

	if fr.Encoding != "base64" {
		return nil, fmt.Errorf("unexpected encoding for %s: %s", path, fr.Encoding)
	}

	decoded, err := base64.StdEncoding.DecodeString(fr.Content)
	if err != nil {
		return nil, fmt.Errorf("decoding content for %s: %w", path, err)
	}
	return decoded, nil
}

func (p *GiteaProvider) GetLatestCommit() (string, error) {
	reqURL := fmt.Sprintf("%s/repos/%s/%s/commits?limit=1", p.apiBase, p.org, p.repo)
	resp, err := p.doRequest(reqURL)
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

func (p *GiteaProvider) HasFileChanged(path, sinceCommit, untilCommit string) (bool, error) {
	// Gitea does not have a direct compare API like GitHub.
	// We use the git compare endpoint available since Gitea 1.17.
	reqURL := fmt.Sprintf("%s/repos/%s/%s/compare/%s...%s",
		p.apiBase, p.org, p.repo, sinceCommit, untilCommit)
	resp, err := p.doRequest(reqURL)
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
