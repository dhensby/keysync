package main

import (
	"bufio"
	"fmt"
	flag "github.com/ogier/pflag"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"
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

	userDir, uid, gid := getUserInfo()

	// try to find authorized_keys file
	keysFile := userDir + "/.ssh/authorized_keys"

	fmt.Printf("Attempting to find key file %s\n", keysFile)
	fh, err := os.OpenFile(keysFile, os.O_RDWR | os.O_CREATE, os.FileMode(0600))

	if err != nil {
		log.Fatal(err)
	}

	fh.Chown(uid, gid)

	keyFileScanner := bufio.NewScanner(fh)

	newKeyFileData := ""
	inMarker := false

	for keyFileScanner.Scan() {
		line := keyFileScanner.Text()
		if inMarker {
			if line == "### END OF AUTOMATICALLY MANAGED KEYS ###" {
				inMarker = false
			}
		} else {
			if line == "### AUTOMATICALLY MANAGED KEYS ###" {
				inMarker = true
			} else if line != "" {
				newKeyFileData += line + "\n"
			}
		}
	}

	fileStat, err := fh.Stat()
	err = fh.Truncate(fileStat.Size())

	// insert new keys between markers
	// write authorized_keys file with correct perms / owner

	fmt.Printf("Loading public keys from gist: %s\n", gist)
	result := getGist(gist)

	scanner := bufio.NewScanner(strings.NewReader(result))

	newKeyFileData += "\n### AUTOMATICALLY MANAGED KEYS ###\n\n"

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		newKeyFileData += line + "\n"
	}

	newKeyFileData += "\n### END OF AUTOMATICALLY MANAGED KEYS ###\n\n"

	// this seems to append not overwrite
	_, err = fh.WriteString(newKeyFileData)

	if err != nil {
		log.Fatal(err)
	}

	err = fh.Close()

	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	flag.StringVarP(&gist, "gist", "g", "", "The gist to use as the source of your public keys")
	flag.StringVarP(&username, "user", "u", "", "The user to manage the authorised keys of")
}

func getUserInfo() (string, int, int) {
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

	uid, err := strconv.Atoi(osUser.Uid)
	gid, err := strconv.Atoi(osUser.Gid)

	return osUser.HomeDir, uid, gid
}