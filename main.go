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
	"runtime/debug"
	"strconv"
	"strings"
)

var (
	ghuser      string
	username    string
	authfile    string
	showversion bool
	versionnum  string
)

const (
	STARTMARKER = "### AUTOMATICALLY MANAGED KEYS ###"
	ENDMARKER   = "### END OF AUTOMATICALLY MANAGED KEYS ###"
)

func printversion() {
	var versionstring string
	var commit string
	var modified bool
	if versionnum != "" {
		versionstring = fmt.Sprintf("%s", versionnum)
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				commit = setting.Value[0:7]
			} else if setting.Key == "vcs.modified" {
				modified = setting.Value == "true"
			}
		}
	}
	if modified {
		fmt.Println("🚨 WARNING! Binary was built from modified git working copy 🚨")
	} else if commit == "" {
		fmt.Println("🚨 WARNING! Binary was not built in git repository 🚨")
	}
	if commit != "" {
		versionstring = fmt.Sprintf("%s (%s)", versionstring, commit)
	}
	fmt.Printf("%s: %s\n", filepath.Base(os.Args[0]), versionstring)
}

func main() {
	// Parse the CLI flags
	flag.Parse()
	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if showversion {
		printversion()
		return
	}

	// get the file handle for the key file and pull out the data
	fh := getKeyFileHandle(authfile)
	defer fh.Close()
	keyData := getKeyLines(fh)

	// load the keys from the github user
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

	totalKeys := strings.Count(result, "\n")

	if len(result) > 0 && !strings.HasSuffix(result, "\n") {
		totalKeys++
	}

	notify(totalKeys)
}

func init() {
	osUser, err := user.Current()

	if err != nil {
		log.Fatal("Unable to determine current user")
	}

	flag.StringVar(&ghuser, "gh-user", "", "The GitHub user to fetch keys for")
	flag.StringVar(&username, "user", osUser.Username, "The user to manage the authorised keys of")
	flag.StringVar(&authfile, "authfile", "authorized_keys", "The authorized_keys file")
	flag.BoolVar(&showversion, "version", false, "Show the version info")
}

func getUserInfo() (string, int, int) {
	osUser, err := user.Lookup(username)

	if err != nil {
		log.Fatalf("Failed to get user: %s", username)
	}

	uid, err := strconv.Atoi(osUser.Uid)
	gid, err := strconv.Atoi(osUser.Gid)

	return osUser.HomeDir, uid, gid
}

func getKeyFileHandle(filename string) *os.File {
	userDir, uid, gid := getUserInfo()

	baseFile := filepath.ToSlash(userDir + "/.ssh/")

	// try to find authorized_keys file
	keysFile := filepath.Clean(filepath.ToSlash(userDir + "/.ssh/" + filename))

	relPath, err := filepath.Rel(baseFile, keysFile)

	if err != nil || strings.HasPrefix(relPath, "..") {
		log.Fatal("Invalid authfile name. Attempt to traverse out of .ssh")
	}

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
}

func notify(added int) {
	if config.PushoverAppKey == "" || config.PushoverUserKey == "" {
		return
	}

	var message string

	if added == 1 {
		message = fmt.Sprintf("is %d key", added)
	} else {
		message = fmt.Sprintf("are %d keys", added)
	}

	hostname, err := os.Hostname()

	formData := url.Values{
		"token":   {config.PushoverAppKey},
		"user":    {config.PushoverUserKey},
		"title":   {fmt.Sprintf("SSH keys update on %s", hostname)},
		"message": {fmt.Sprintf("There %s from GitHub now authorized for user %s", message, username)},
	}

	body := bytes.NewBufferString(formData.Encode())

	resp, err := http.Post("https://api.pushover.net/1/messages.json", "application/x-www-form-urlencoded", body)

	if err != nil {
		return
	}

	defer resp.Body.Close()
}
