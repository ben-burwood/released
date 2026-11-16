package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

// Release-note images on private repos point at github.com/user-attachments/assets/<id>, which 404 for anyone without a GitHub session.
// Our token gets a redirect to a signed, publicly-fetchable S3 URL instead, valid for 5 minutes.
const assetsURL = "https://github.com/user-attachments/assets/"

// AssetURLRe matches a private attachment URL, capturing its asset ID
var AssetURLRe = regexp.MustCompile(`https://github\.com/user-attachments/assets/([0-9a-fA-F-]+)`)

var assetIDRe = regexp.MustCompile(`^[0-9a-fA-F-]+$`)

var ErrInvalidAssetID = errors.New("invalid asset id")

// AssetURL returns the original (session-only on private repos) URL of an attachment
func AssetURL(id string) string {
	return assetsURL + id
}

// ResolveAsset returns the signed, browser-fetchable URL of an attachment
func (c *Client) ResolveAsset(ctx context.Context, id string) (string, error) {
	if !assetIDRe.MatchString(id) {
		return "", ErrInvalidAssetID
	}
	req, err := c.newRequest(ctx, c.assetsURL+id)
	if err != nil {
		return "", err
	}

	noRedirect := *c.http
	noRedirect.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := noRedirect.Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	location := resp.Header.Get("Location")
	if !isRedirect(resp.StatusCode) || location == "" {
		return "", fmt.Errorf("GitHub %d resolving asset %s", resp.StatusCode, id)
	}
	return location, nil
}

func isRedirect(status int) bool {
	switch status {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	}
	return false
}
