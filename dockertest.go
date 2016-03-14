package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/coduno/runtime-dummy/controllers"
	"github.com/coduno/runtime-dummy/env"
	"google.golang.org/appengine"
)

func main() {
	http.Handle("/", controllers.Router())
	if env.IsDevAppServer() {
		fmt.Println("DEV SERVER")
		os.Setenv("PORT", "8081")
	}
	appengine.Main()
}
