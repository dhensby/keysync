package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"log"
	"strings"
)

func getGitHubUserKeys(ghuser string) string {
	client := github.NewClient(nil)
	opt := github.ListOptions{PerPage: 10}

	var allKeys = []*github.Key{}
	for {
		keys, resp, err := client.Users.ListKeys(context.Background(), ghuser, &opt)
		if err != nil {
			log.Fatalf("Failed to fetch user keys: %s\n", err)
		}
		allKeys = append(allKeys, keys...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	var keys = ""
	for _, key := range allKeys {
		keys += strings.TrimSpace(fmt.Sprintf("%s %s", key.GetKey(), key.GetTitle())) + "\n"
	}
	return keys
}
