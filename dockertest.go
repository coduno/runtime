package main

import (
	"log"
	"net/http"
	"time"

	"github.com/coduno/runtime/controllers"
	"github.com/coduno/runtime/runner"
)

const sock = ":8081"

func main() {
	go runner.Scrape(10 * time.Minute)

	http.Handle("/", controllers.Router())
	log.Println("Listening at", sock)
	log.Fatal(http.ListenAndServe(sock, nil))
}
