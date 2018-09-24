package main

import (
	"encoding/json"
	"log"
	"os"
)

var (
	config Config
)

type Config struct {
	PushoverAppKey string
	PushoverUserKey string
}

func init() {
	fh, err := os.Open(".keysync.config")
	if err != nil {
		return
	}

	defer fh.Close()

	decoder := json.NewDecoder(fh)

	err =  decoder.Decode(&config)

	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println(config.PushoverAppKey)

}
