package repos

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	for entry, want := range map[string]string{
		"yuin/goldmark":   "yuin/goldmark",
		" yuin/goldmark ": "yuin/goldmark",
		"https://github.com/yuin/goldmark/releases":  "yuin/goldmark",
		"https://github.com/yuin/goldmark":           "yuin/goldmark",
		"github.com/yuin/goldmark.git":               "yuin/goldmark",
		"https://www.github.com/org/repo.name/tags/": "org/repo.name",
	} {
		if got, err := Parse(entry); err != nil || got != want {
			t.Errorf("Parse(%q) = %q, %v, want %q", entry, got, err, want)
		}
	}

	for _, entry := range []string{"", "goldmark", "https://gitlab.com/yuin/goldmark", "a/b c"} {
		if got, err := Parse(entry); err == nil {
			t.Errorf("Parse(%q) = %q, want error", entry, got)
		}
	}
}

func TestLoadFromEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	file := "repos:\n  - https://github.com/yuin/goldmark/releases\n  - foo/bar # duplicate\n"
	if err := os.WriteFile(path, []byte(file), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(ListEnvVar, "foo/bar, Foo/Baz")
	t.Setenv(FileEnvVar, path)

	got, err := LoadFromEnv()
	want := []string{"foo/bar", "Foo/Baz", "yuin/goldmark"}
	if err != nil || !slices.Equal(got, want) {
		t.Errorf("LoadFromEnv() = %v, %v, want %v", got, err, want)
	}
}

func TestLoadFromEnvEmpty(t *testing.T) {
	t.Setenv(ListEnvVar, "")
	t.Setenv(FileEnvVar, "")

	if _, err := LoadFromEnv(); err == nil || !strings.Contains(err.Error(), "no repos") {
		t.Errorf("LoadFromEnv() error = %v, want no repos", err)
	}
}
