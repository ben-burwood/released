package api

import (
	"encoding/json"
	"net/http"

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
