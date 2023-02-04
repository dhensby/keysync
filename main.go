package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	ghuser   string
	username string
)

const (
	STARTMARKER = "### AUTOMATICALLY MANAGED KEYS ###"
	ENDMARKER   = "### END OF AUTOMATICALLY MANAGED KEYS ###"
)

func main() {
	// Parse the CLI flags
	flag.Parse()
	if flag.NFlag() == 0 {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// get the file handle for the key file and pull out the data
	fh := getKeyFileHandle("")
	defer fh.Close()
	keyData := getKeyLines(fh)

	// load the keys from the github user
	fmt.Printf("Loading public keys from GitHub user: %s\n", ghuser)
	result := getGitHubUserKeys(ghuser)

	// build a new file content for the key file with the current file data
	scanner := bufio.NewScanner(strings.NewReader(result))

	keyData += "\n" + STARTMARKER + "\n\n"

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		keyData += line + "\n"
	}

	keyData += "\n" + ENDMARKER + "\n"

	writeFileContent(fh, keyData)
}

func init() {
	flag.StringVar(&ghuser, "gh-user", "", "The GitHub user to fetch keys for")
	flag.StringVar(&username, "user", "", "The user to manage the authorised keys of")
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

func getKeyFileHandle(filename string) *os.File {
	if filename == "" {
		filename = "authorized_keys"
	}

	userDir, uid, gid := getUserInfo()

	// try to find authorized_keys file
	keysFile := filepath.ToSlash(userDir + "/.ssh/authorized_keys")

	fmt.Printf("Attempting to find key file %s\n", keysFile)

	fh, err := os.OpenFile(keysFile, os.O_RDWR|os.O_CREATE, os.FileMode(0600))

	if err != nil {
		log.Fatal(err)
	}

	err = fh.Chown(uid, gid)

	if err != nil {
		log.Print(err)
	}

	return fh
}

func getKeyLines(fh *os.File) string {

	keyFileScanner := bufio.NewScanner(fh)

	newKeyFileData := ""
	inMarker := false

	for keyFileScanner.Scan() {
		line := keyFileScanner.Text()
		if inMarker {
			if line == ENDMARKER {
				inMarker = false
			}
		} else {
			if line == STARTMARKER {
				inMarker = true
			} else if line != "" {
				newKeyFileData += line + "\n"
			}
		}
	}

	return newKeyFileData
}

func writeFileContent(fh *os.File, content string) {

	_, err := fh.Seek(0, 0)

	if err != nil {
		log.Fatal(err)
	}

	originalMd5 := md5.New()

	_, err = io.Copy(originalMd5, fh)

	if err != nil {
		log.Fatal(err)
	}

	newMd5 := md5.New()

	_, err = io.WriteString(newMd5, content)

	if err != nil {
		log.Fatal(err)
	}

	if bytes.Equal(originalMd5.Sum(nil), newMd5.Sum(nil)) {
		return
	}

	_, err = fh.Seek(0, 0)

	if err != nil {
		log.Fatal(err)
	}

	bytesWritten, err := fh.WriteString(content)

	if err != nil {
		log.Fatal(err)
	}

	err = fh.Truncate(int64(bytesWritten))

	if err != nil {
		log.Fatal(err)
	}

	err = fh.Sync()

	if err != nil {
		log.Fatal(err)
	}
	notify()
}

func notify() {
	if config.PushoverAppKey == "" || config.PushoverUserKey == "" {
		return
	}

	hostname, err := os.Hostname()

	formData := url.Values{
		"token":   {config.PushoverAppKey},
		"user":    {config.PushoverUserKey},
		"title":   {"SSH keys update on " + hostname},
		"message": {"The SSH keys on the device " + hostname + " were updated"},
	}

	body := bytes.NewBufferString(formData.Encode())

	resp, err := http.Post("https://api.pushover.net/1/messages.json", "application/x-www-form-urlencoded", body)

	if err != nil {
		return
	}

	defer resp.Body.Close()
}
