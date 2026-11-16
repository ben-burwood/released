package releases

import (
	"bytes"
	"cmp"
	"context"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"

	"releases/internal/assets"
	"releases/internal/github"
)

// fetchTimeout bounds a single upstream fetch, which runs detached from the request that triggered it
const fetchTimeout = 2 * time.Minute

// Release is a single release of the repo, with its notes rendered to HTML
type Release struct {
	Tag         string     `json:"tag"`
	Name        string     `json:"name"`
	BodyHTML    string     `json:"bodyHtml,omitempty"` // empty when the release has no notes
	PublishedAt *time.Time `json:"publishedAt"`
	Author      string     `json:"author,omitempty"`
	Prerelease  bool       `json:"prerelease"`
	Draft       bool       `json:"draft"`
	URL         string     `json:"url"`
}

// Releases is every release of Repo, drafts first then newest first
type Releases struct {
	Repo     string    `json:"repo"`
	Releases []Release `json:"releases"`
	Fetched  time.Time `json:"fetchedAt"`
}

// Service fetches and caches the releases of a single repo
type Service struct {
	gh   *github.Client
	repo string
	ttl  time.Duration

	mu        sync.Mutex // held for the duration of a fetch, so concurrent requests share it
	cached    *Releases
	expiresAt time.Time
}

func NewService(gh *github.Client, repo string, ttl time.Duration) *Service {
	return &Service{gh: gh, repo: repo, ttl: ttl}
}

// Releases returns the repo's releases, served from cache unless expired. Fetch errors are returned, not cached.
func (s *Service) Releases(ctx context.Context) (*Releases, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cached != nil && time.Now().Before(s.expiresAt) {
		return s.cached, nil
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), fetchTimeout)
	defer cancel()
	releases, err := s.fetch(ctx)
	if err != nil {
		return nil, err
	}
	s.cached, s.expiresAt = releases, time.Now().Add(s.ttl)
	return releases, nil
}

func (s *Service) fetch(ctx context.Context) (*Releases, error) {
	items, err := s.gh.ListReleases(ctx, s.repo)
	if err != nil {
		return nil, err
	}

	releases := make([]Release, 0, len(items))
	for _, item := range items {
		r := Release{
			Tag:         item.TagName,
			Name:        cmp.Or(item.Name, item.TagName),
			BodyHTML:    renderBody(assets.Rewrite(item.Body)),
			PublishedAt: item.PublishedAt,
			Prerelease:  item.Prerelease,
			Draft:       item.Draft,
			URL:         item.HTMLURL,
		}
		if item.Author != nil {
			r.Author = item.Author.Login
		}
		releases = append(releases, r)
	}
	slices.SortStableFunc(releases, compareReleases)

	return &Releases{Repo: s.repo, Releases: releases, Fetched: time.Now().UTC()}, nil
}

// compareReleases orders drafts first, then published releases newest first, then any unpublished releases
func compareReleases(a, b Release) int {
	if a.Draft != b.Draft {
		if a.Draft {
			return -1
		}
		return 1
	}
	switch {
	case a.PublishedAt == nil && b.PublishedAt == nil:
		return 0
	case a.PublishedAt == nil:
		return 1
	case b.PublishedAt == nil:
		return -1
	}
	return b.PublishedAt.Compare(*a.PublishedAt)
}

var (
	markdownParser   = parser.New(parser.WithExtensions(extension.NewTableParser(), extension.NewLinkifyParser()))
	markdownRenderer = html.New(
		html.WithExtensions(extension.NewTableHTMLRenderer()),
		html.WithHardWraps(), // newlines become <br>, as GitHub renders release notes
		html.WithUnsafe(),    // release notes commonly embed raw <img> tags
	)
)

// Drop GitHub's auto-generated "Full Changelog: …" trailer line.
var changelogRe = regexp.MustCompile(`(?im)^\s*\*{0,2}Full Changelog\*{0,2}:.*$`)

// renderBody renders release notes to HTML, or "" if there is nothing left to show
func renderBody(body string) string {
	body = changelogRe.ReplaceAllString(body, "")
	if strings.TrimSpace(body) == "" {
		return ""
	}
	source := []byte(body)
	var buf bytes.Buffer
	if err := markdownRenderer.Render(&buf, source, markdownParser.Parse(source)); err != nil {
		return ""
	}
	return buf.String()
}
