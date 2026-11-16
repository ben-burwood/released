package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"releases/internal/api"
	"releases/internal/github"
	"releases/internal/releases"
)

const (
	// RepoEnvVar is the repo (owner/name) to list releases from
	RepoEnvVar = "GITHUB_REPO"

	cacheTTL = 60 * time.Second
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

	svc := releases.NewService(gh, repo, cacheTTL)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/releases", api.ReleasesHandler(svc))

	// Serve Static Frontend
	mux.Handle("/", http.FileServer(http.Dir("./frontend/dist")))

	log.Fatal(http.ListenAndServe("[::]:8080", api.CORSMiddleware(mux)))
}
