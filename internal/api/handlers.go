package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"releases/internal/assets"
	"releases/internal/releases"
)

// ReposHandler handles requests for the configured repos (owner/name), the first being the default
func ReposHandler(repos []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repos)
	}
}

// ReleasesHandler handles requests for the releases of ?repo= (owner/name, defaulting to the first configured repo),
// drafts first then newest first. Only configured repos are served, so the token can't be used to read any other.
func ReleasesHandler(repos []string, services map[string]*releases.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := r.URL.Query().Get("repo")
		if repo == "" {
			repo = repos[0]
		}
		svc, ok := services[strings.ToLower(repo)]
		if !ok {
			http.Error(w, "unknown repo "+repo, http.StatusNotFound)
			return
		}

		data, err := svc.Releases(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

// AssetHandler handles requests for a release-note attachment ({id}), redirecting to its signed URL
func AssetHandler(svc *assets.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, err := svc.URL(r.Context(), r.PathValue("id"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Cache-Control", "no-store") // the signed URL expires, so always come back for a fresh one
		http.Redirect(w, r, url, http.StatusFound)
	}
}
