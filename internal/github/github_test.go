package github

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(srv *httptest.Server) *Client {
	return &Client{token: "test-token", baseURL: srv.URL, http: srv.Client()}
}

func TestListReleasesFollowsPagination(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		if r.URL.Query().Get("page") == "2" {
			fmt.Fprint(w, `[{"tag_name":"v1"}]`)
			return
		}
		w.Header().Set("Link", fmt.Sprintf(`<%s/repos/o/r/releases?per_page=100&page=2>; rel="next", <%[1]s/x>; rel="last"`, srv.URL))
		fmt.Fprint(w, `[{"tag_name":"v2","name":"Two","author":{"login":"me"}}]`)
	}))
	defer srv.Close()

	releases, err := newTestClient(srv).ListReleases(context.Background(), "o/r")
	if err != nil {
		t.Fatal(err)
	}
	if len(releases) != 2 || releases[0].TagName != "v2" || releases[1].TagName != "v1" {
		t.Fatalf("releases = %+v", releases)
	}
	if releases[0].Author.Login != "me" || releases[1].Author != nil {
		t.Errorf("authors = %+v, %+v", releases[0].Author, releases[1].Author)
	}
}

func TestListReleasesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer srv.Close()

	if _, err := newTestClient(srv).ListReleases(context.Background(), "o/r"); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("err = %v", err)
	}
}

func TestNextLink(t *testing.T) {
	for header, want := range map[string]string{
		"": "",
		`<https://a/x?page=3>; rel="next", <https://a/x?page=9>; rel="last"`:  "https://a/x?page=3",
		`<https://a/x?page=1>; rel="prev", <https://a/x?page=1>; rel="first"`: "",
	} {
		if got := nextLink(header); got != want {
			t.Errorf("nextLink(%q) = %q, want %q", header, got, want)
		}
	}
}

func TestAssetResolverRewrite(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if strings.HasSuffix(r.URL.Path, "/abc-123") {
			http.Redirect(w, r, "https://signed.example/abc", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	a := c.NewAssetResolver(context.Background())
	// Point the resolver at the test server, keeping the no-redirect policy
	a.http.Transport = &rewriteTransport{target: srv.URL}

	body := "![x](https://github.com/user-attachments/assets/abc-123) and https://github.com/user-attachments/assets/abc-123 and https://github.com/user-attachments/assets/def"
	got := a.Rewrite(body)
	want := "![x](https://signed.example/abc) and https://signed.example/abc and https://github.com/user-attachments/assets/def"
	if got != want {
		t.Errorf("Rewrite = %q, want %q", got, want)
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2 (repeated URLs are cached)", calls)
	}
}

// rewriteTransport sends every request to target, keeping the path
type rewriteTransport struct{ target string }

func (t *rewriteTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	req := r.Clone(r.Context())
	u, _ := req.URL.Parse(t.target + r.URL.Path)
	req.URL, req.Host = u, u.Host
	return http.DefaultTransport.RoundTrip(req)
}
