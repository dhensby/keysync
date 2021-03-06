package main

import (
	"context"
	"github.com/google/go-github/github"
	"log"
)

func getGist(gistid string) string {
	client := github.NewClient(nil)
	// @todo look into caching mechanism
	gist, _, err := client.Gists.Get(context.Background(), gistid)
	if err != nil {
		log.Fatalf("Failed to fetch gist: %s\n", err)
	}
	var keys = "";
	for _, v := range gist.Files {
		keys += *v.Content
	}
	return keys
}
