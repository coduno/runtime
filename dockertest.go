package main

import (
	"net/http"

	"github.com/coduno/runtime-dummy/controllers"
	"google.golang.org/appengine"
)

func main() {
	http.Handle("/", controllers.Router())
//	log.Fatal(http.ListenAndServe(":8081", nil))
	appengine.Main()
}
