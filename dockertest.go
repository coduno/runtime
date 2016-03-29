package main

import (
	"net/http"
	"os"
	"time"

	"github.com/coduno/runtime/controllers"
	"github.com/coduno/runtime/runner"
	"google.golang.org/appengine"
)

func main() {
	go runner.Scrape(10 * time.Minute)

	http.Handle("/", controllers.Router())
	os.Setenv("PORT", "8081")
	appengine.Main()
}
