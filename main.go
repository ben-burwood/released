package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"releases/internal/api"
	"releases/internal/assets"
	"releases/internal/github"
	"releases/internal/releases"
)

const (
	// RepoEnvVar is the repo (owner/name) to list releases from
	RepoEnvVar = "GITHUB_REPO"

	cacheTTL = 60 * time.Second

	// assetTTL caches signed attachment URLs, leaving a minute of GitHub's 5 minute signature for the browser to use them
	assetTTL = 4 * time.Minute
)

func main() {
	gh, err := github.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	repo := os.Getenv(RepoEnvVar)
	if repo == "" {
		log.Fatal(RepoEnvVar + " required")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/releases", api.ReleasesHandler(releases.NewService(gh, repo, cacheTTL)))
	mux.HandleFunc("GET "+assets.PathPrefix+"{id}", api.AssetHandler(assets.NewService(gh, assetTTL)))

	// Serve Static Frontend
	mux.Handle("/", http.FileServer(http.Dir("./frontend/dist")))

	log.Fatal(http.ListenAndServe("[::]:8080", api.CORSMiddleware(mux)))
}
