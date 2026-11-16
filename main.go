package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"releases/internal/api"
	"releases/internal/assets"
	"releases/internal/github"
	"releases/internal/releases"
	"releases/internal/repos"
)

const (
	cacheTTL = 60 * time.Second

	// assetTTL caches signed attachment URLs, leaving a minute of GitHub's 5 minute signature for the browser to use them
	assetTTL = 4 * time.Minute
)

func main() {
	gh, err := github.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	repoList, err := repos.LoadFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	services := make(map[string]*releases.Service, len(repoList))
	for _, repo := range repoList {
		services[strings.ToLower(repo)] = releases.NewService(gh, repo, cacheTTL)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/repos", api.ReposHandler(repoList))
	mux.HandleFunc("GET /api/releases", api.ReleasesHandler(repoList, services))
	mux.HandleFunc("GET "+assets.PathPrefix+"{id}", api.AssetHandler(assets.NewService(gh, assetTTL)))

	// Serve Static Frontend
	mux.Handle("/", http.FileServer(http.Dir("./frontend/dist")))

	log.Fatal(http.ListenAndServe("[::]:8080", api.CORSMiddleware(mux)))
}
