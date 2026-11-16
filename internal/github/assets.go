package github

import (
	"context"
	"net/http"
	"regexp"
)

// Release-note images on private repos point at github.com/user-attachments/assets/<id>, which 404 for anyone without a GitHub session.
// Resolve them server-side (our token gets a 302 to a signed, publicly-fetchable S3 URL) and rewrite the body so browsers can load them.
var assetRe = regexp.MustCompile(`https://github\.com/user-attachments/assets/[0-9a-fA-F-]+`)

// AssetResolver rewrites private GitHub attachment URLs to their signed, browser-fetchable equivalents.
// Signed URLs expire, so use a fresh resolver per fetch rather than sharing one.
type AssetResolver struct {
	ctx    context.Context
	client *Client
	http   *http.Client
	cache  map[string]string
}

// NewAssetResolver creates an AssetResolver that resolves attachments with the client's token, without following redirects
func (c *Client) NewAssetResolver(ctx context.Context) *AssetResolver {
	noRedirect := &http.Client{
		Timeout:       c.http.Timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	return &AssetResolver{ctx: ctx, client: c, http: noRedirect, cache: map[string]string{}}
}

// Rewrite replaces every attachment URL in body with its signed URL, leaving any that fail to resolve unchanged
func (a *AssetResolver) Rewrite(body string) string {
	return assetRe.ReplaceAllStringFunc(body, func(original string) string {
		if resolved, ok := a.cache[original]; ok {
			return resolved
		}
		resolved := a.resolve(original)
		a.cache[original] = resolved
		return resolved
	})
}

func (a *AssetResolver) resolve(url string) string {
	req, err := a.client.newRequest(a.ctx, url)
	if err != nil {
		return url
	}
	resp, err := a.http.Do(req)
	if err != nil {
		return url
	}
	resp.Body.Close()
	if location := resp.Header.Get("Location"); isRedirect(resp.StatusCode) && location != "" {
		return location
	}
	return url
}

func isRedirect(status int) bool {
	switch status {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	}
	return false
}
