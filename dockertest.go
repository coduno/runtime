package main

import (
	"net/http"
	"os"

	"github.com/coduno/runtime-dummy/controllers"
	"google.golang.org/appengine"
)

func main() {
	http.Handle("/", controllers.Router())
	os.Setenv("PORT", "8081")
	appengine.Main()
}
