package main

import (
	go_vanity_urls "go-vanity-urls"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(cwd)

	file, err := os.Open(path.Join(cwd, "config.yaml"))
	if err != nil {
		log.Fatal(err)
	}

	config, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	err = file.Close()
	if err != nil {
		log.Fatal(err)
	}

	h, err := go_vanity_urls.NewHandler(config)
	if err != nil {
		log.Fatal(err)
	}

	err = http.ListenAndServe(":8080", h)
	if err != nil {
		log.Fatal(err)
	}
}
