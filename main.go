package main

import (
	"fmt"
	flag "github.com/ogier/pflag"
	"os"
)

var (
	gist string
	user string
)

func main() {
	flag.Parse()
	if flag.NFlag() == 0 {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("Loading public keys from gist: %s\n", gist)
	result := getGist(gist)
	fmt.Println(result)
}

func init() {
	flag.StringVarP(&gist, "gist", "g", "", "The gist to use as the source of your public keys")
	flag.StringVarP(&user, "user", "u", "", "The user to manage the authorised keys of")
}