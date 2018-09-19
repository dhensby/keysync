package main

import (
	"bufio"
	"fmt"
	flag "github.com/ogier/pflag"
	"io/ioutil"
	"os"
	"strings"
	"os/user"
)

var (
	gist string
	username string
)

func main() {
	flag.Parse()
	if flag.NFlag() == 0 {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	userDir := getUserDir()

	// try to find authorized_keys file
	// create it if it doesn't exist with correct perms / owner
	// look in file for start marker
	// remove contents from start marker to end marker
	// insert new keys between markers
	// write authorized_keys file with correct perms / owner
	ioutil.ReadDir(userDir)

	fmt.Printf("Loading public keys from gist: %s\n", gist)
	result := getGist(gist)

	scanner := bufio.NewScanner(strings.NewReader(result))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		fmt.Println(line)
	}
}

func init() {
	flag.StringVarP(&gist, "gist", "g", "", "The gist to use as the source of your public keys")
	flag.StringVarP(&username, "user", "u", "", "The user to manage the authorised keys of")
}

func getUserDir() string {
	var osUser *user.User
	var err error
	if username == "" {
		osUser, err = user.Current()
		fmt.Printf("No user supplied, assuming %s\n", osUser.Username)
	} else {
		osUser, err = user.Lookup(username)
	}

	if err != nil {
		fmt.Println("Failed to get user")
	}

	return osUser.HomeDir
}