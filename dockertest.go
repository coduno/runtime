package main

import (
	"log"
	"net/http"

	"github.com/coduno/runtime-dummy/controllers"
)

func main() {
	http.Handle("/", controllers.Router())
	log.Fatal(http.ListenAndServe(":8081", nil))
}
