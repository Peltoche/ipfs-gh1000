package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	res, err := http.Get("https://gitstar-ranking.com/repositories")
	if err != nil {
		log.Fatalf("failed to get the gitstar-ranking.com page: %s", err)
	}
	defer func() { _ = res.Body.Close() }()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("failed to read the body of gitstar-ranking.com: %s", err)
	}

	fmt.Printf("body: %q", body)
}
