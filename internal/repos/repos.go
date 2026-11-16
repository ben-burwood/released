package repos

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

// config is the YAML config file, e.g.
//
//	repos:
//	  - org/repo
//	  - https://github.com/yuin/goldmark/releases
type config struct {
	Repos []string `yaml:"repos"`
}

const (
	// ListEnvVar is a comma or whitespace separated list of repos to list releases from
	ListEnvVar = "GITHUB_REPOS"

	// FileEnvVar is the path to a YAML config file listing repos under `repos:`
	FileEnvVar = "GITHUB_REPOS_FILE"
)

// repoRe matches a repo as owner/name, optionally as a github.com URL with a trailing path (e.g. /releases)
var repoRe = regexp.MustCompile(`^(?:(?:https?://)?(?:www\.)?github\.com/)?([A-Za-z0-9-]+)/([A-Za-z0-9._-]+?)(?:\.git)?(?:/.*)?$`)

// Parse normalises a repo entry (owner/name or a github.com URL) to owner/name
func Parse(entry string) (string, error) {
	m := repoRe.FindStringSubmatch(strings.TrimSpace(entry))
	if m == nil {
		return "", fmt.Errorf("invalid repo %q, want owner/name or a github.com URL", entry)
	}
	return m[1] + "/" + m[2], nil
}

// LoadFromEnv returns the configured repos, from GITHUB_REPOS then GITHUB_REPOS_FILE, deduplicated
func LoadFromEnv() ([]string, error) {
	entries := strings.FieldsFunc(os.Getenv(ListEnvVar), func(r rune) bool { return r == ',' || r == ' ' || r == '\n' || r == '\t' })

	if path := os.Getenv(FileEnvVar); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var cfg config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		entries = append(entries, cfg.Repos...)
	}

	var repos []string
	for _, entry := range entries {
		repo, err := Parse(entry)
		if err != nil {
			return nil, err
		}
		if !slices.ContainsFunc(repos, func(r string) bool { return strings.EqualFold(r, repo) }) {
			repos = append(repos, repo)
		}
	}
	if len(repos) == 0 {
		return nil, fmt.Errorf("no repos configured, set %s or %s", ListEnvVar, FileEnvVar)
	}
	return repos, nil
}
