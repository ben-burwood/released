package api

import (
	"encoding/json"
	"net/http"

	"releases/internal/assets"
	"releases/internal/releases"
)

// ReleasesHandler handles requests for the repo's releases, drafts first then newest first
func ReleasesHandler(svc *releases.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
