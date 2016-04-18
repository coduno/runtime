package main

import (
	"log"
	"net/http"
	"time"

	"github.com/coduno/runtime/controllers"
	"github.com/coduno/runtime/runner"
)

const sock = ":8081"

func logged(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	go runner.Scrape(2 * time.Minute)

	http.Handle("/", controllers.Router())
	log.Println("Listening at", sock)
	log.Fatal(http.ListenAndServe(sock, logged(http.DefaultServeMux)))
}
