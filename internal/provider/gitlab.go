package provider

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

// GitLabProvider implements Provider for GitLab and self-hosted GitLab instances.
type GitLabProvider struct {
	httpClient *http.Client
	token      string
	apiBase    string
	projectID  string // "org/repo" URL-encoded as project ID
}

// NewGitLab creates a GitLab provider from the given parameters.
func NewGitLab(host, org, repo string) (*GitLabProvider, error) {
	token, err := resolveGitLabToken()
	if err != nil {
		return nil, err
	}

	if host == "" {
		host = "gitlab.com"
	}
	apiBase := fmt.Sprintf("https://%s/api/v4", host)
	projectID := url.PathEscape(org + "/" + repo)

	return &GitLabProvider{
		httpClient: &http.Client{},
		token:      token,
		apiBase:    apiBase,
		projectID:  projectID,
	}, nil
}

func resolveGitLabToken() (string, error) {
	if t := os.Getenv("SKILLX_GITLAB_TOKEN"); t != "" {
		return t, nil
	}
	if t := os.Getenv("GITLAB_TOKEN"); t != "" {
		return t, nil
	}
	out, err := exec.Command("glab", "auth", "token").Output()
	if err == nil {
		t := strings.TrimSpace(string(out))
		if t != "" {
			return t, nil
		}
	}
	return "", fmt.Errorf("no GitLab token found. Set SKILLX_GITLAB_TOKEN, GITLAB_TOKEN, or run 'glab auth login'")
}

func (p *GitLabProvider) doRequest(reqURL string) (*http.Response, error) {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", p.token)
	return p.httpClient.Do(req)
}

func (p *GitLabProvider) FetchFile(path string) ([]byte, error) {
	encodedPath := url.PathEscape(path)
	reqURL := fmt.Sprintf("%s/projects/%s/repository/files/%s?ref=HEAD", p.apiBase, p.projectID, encodedPath)
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

func (p *GitLabProvider) GetLatestCommit() (string, error) {
	reqURL := fmt.Sprintf("%s/projects/%s/repository/commits?per_page=1", p.apiBase, p.projectID)
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
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return "", fmt.Errorf("decoding commits response: %w", err)
	}
	if len(commits) == 0 {
		return "", fmt.Errorf("no commits found in repository")
	}
	return commits[0].ID, nil
}

func (p *GitLabProvider) HasFileChanged(path, sinceCommit, untilCommit string) (bool, error) {
	reqURL := fmt.Sprintf("%s/projects/%s/repository/compare?from=%s&to=%s",
		p.apiBase, p.projectID, sinceCommit, untilCommit)
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
		Diffs []struct {
			NewPath string `json:"new_path"`
			OldPath string `json:"old_path"`
		} `json:"diffs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&compareResp); err != nil {
		return false, fmt.Errorf("decoding compare response: %w", err)
	}

	for _, d := range compareResp.Diffs {
		if strings.HasPrefix(d.NewPath, path) || strings.HasPrefix(d.OldPath, path) {
			return true, nil
		}
	}
	return false, nil
}
