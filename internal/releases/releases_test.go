package releases

import (
	"slices"
	"strings"
	"testing"
	"time"
)

func at(s string) *time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return &t
}

func TestCompareReleases(t *testing.T) {
	releases := []Release{
		{Tag: "old", PublishedAt: at("2026-01-01T00:00:00Z")},
		{Tag: "unpublished"},
		{Tag: "draft", Draft: true},
		{Tag: "new", PublishedAt: at("2026-03-01T00:00:00Z")},
		{Tag: "draft-dated", Draft: true, PublishedAt: at("2026-02-01T00:00:00Z")},
	}
	slices.SortStableFunc(releases, compareReleases)

	var got []string
	for _, r := range releases {
		got = append(got, r.Tag)
	}
	want := []string{"draft-dated", "draft", "new", "old", "unpublished"}
	if !slices.Equal(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}

func TestRenderBody(t *testing.T) {
	for _, body := range []string{"", "  \n", "**Full Changelog**: https://github.com/o/r/compare/v1...v2", "full changelog: x"} {
		if got := renderBody(body); got != "" {
			t.Errorf("renderBody(%q) = %q, want empty", body, got)
		}
	}

	got := renderBody("## What's Changed\n* Fix bug\nline one\nline two\n\n<img src=\"https://x/y.png\">\n\nSee https://example.com\n\n**Full Changelog**: https://github.com/o/r/compare/v1...v2")
	for _, want := range []string{"<h2>What", "<li>Fix bug", "line one<br", `<img src="https://x/y.png">`, `<a href="https://example.com">`} {
		if !strings.Contains(got, want) {
			t.Errorf("renderBody missing %q in %q", want, got)
		}
	}
	if strings.Contains(got, "Full Changelog") {
		t.Errorf("renderBody kept changelog trailer: %q", got)
	}
}
