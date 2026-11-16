package assets

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"releases/internal/github"
)

// PathPrefix is where the API serves attachments; release notes link here rather than to GitHub
const PathPrefix = "/api/assets/"

// Rewrite points the private GitHub attachment URLs in body at PathPrefix, so browsers fetch them through Service.
// Signed URLs expire after 5 minutes, so they're resolved when the browser asks for the image, not when the notes are fetched.
func Rewrite(body string) string {
	return github.AssetURLRe.ReplaceAllString(body, PathPrefix+"$1")
}

// Service resolves and caches the signed URLs of attachments
type Service struct {
	gh  *github.Client
	ttl time.Duration // must be shorter than GitHub's 5 minute signature, so a cached URL is still valid when the browser follows it

	mu      sync.Mutex
	entries map[string]entry
}

type entry struct {
	url       string
	expiresAt time.Time
}

func NewService(gh *github.Client, ttl time.Duration) *Service {
	return &Service{gh: gh, ttl: ttl, entries: map[string]entry{}}
}

// URL returns the attachment's signed URL, served from cache unless expired. If it can't be resolved, the original
// GitHub URL is returned (and not cached), which still loads for public repos and signed-in users.
func (s *Service) URL(ctx context.Context, id string) (string, error) {
	s.mu.Lock()
	e, ok := s.entries[id]
	s.mu.Unlock()
	if ok && time.Now().Before(e.expiresAt) {
		return e.url, nil
	}

	url, err := s.gh.ResolveAsset(ctx, id)
	if errors.Is(err, github.ErrInvalidAssetID) {
		return "", err
	}
	if err != nil {
		log.Printf("[assets.URL] %v", err)
		return github.AssetURL(id), nil
	}

	s.mu.Lock()
	s.entries[id] = entry{url: url, expiresAt: time.Now().Add(s.ttl)}
	s.mu.Unlock()
	return url, nil
}
