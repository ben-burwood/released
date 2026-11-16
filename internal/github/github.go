package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"
)

const (
	apiBaseURL = "https://api.github.com"

	// TokenEnvVar is the environment variable holding a read-only PAT with access to the target repo
	TokenEnvVar = "GITHUB_TOKEN"
)

var ErrTokenNotSet = errors.New(TokenEnvVar + " required")

// Release is the subset of a GitHub release used by the dashboard
type Release struct {
	TagName     string     `json:"tag_name"`
	Name        string     `json:"name"`
	Body        string     `json:"body"`
	PublishedAt *time.Time `json:"published_at"`
	Author      *struct {
		Login string `json:"login"`
	} `json:"author"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
	HTMLURL    string `json:"html_url"`
}

type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

// NewClientFromEnv creates a Client authenticated with the GITHUB_TOKEN environment variable
func NewClientFromEnv() (*Client, error) {
	token := os.Getenv(TokenEnvVar)
	if token == "" {
		return nil, ErrTokenNotSet
	}
	return &Client{token: token, baseURL: apiBaseURL, http: &http.Client{Timeout: 30 * time.Second}}, nil
}

func (c *Client) newRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return req, nil
}

// ListReleases returns every release of the repo (owner/name), following pagination, in API order
func (c *Client) ListReleases(ctx context.Context, repo string) ([]Release, error) {
	var releases []Release
	next := fmt.Sprintf("%s/repos/%s/releases?per_page=100", c.baseURL, repo)
	for next != "" {
		req, err := c.newRequest(ctx, next)
		if err != nil {
			return nil, err
		}
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		page, err := decodePage(resp)
		if err != nil {
			return nil, err
		}
		releases = append(releases, page...)
		next = nextLink(resp.Header.Get("Link"))
	}
	return releases, nil
}

func decodePage(resp *http.Response) ([]Release, error) {
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("GitHub API %d on %s: %s", resp.StatusCode, resp.Request.URL.Path, body)
	}
	var page []Release
	return page, json.NewDecoder(resp.Body).Decode(&page)
}

var linkNextRe = regexp.MustCompile(`<([^>]+)>\s*;\s*rel="next"`)

// nextLink returns the rel="next" URL of a Link header, or "" on the last page
func nextLink(header string) string {
	if m := linkNextRe.FindStringSubmatch(header); m != nil {
		return m[1]
	}
	return ""
}
