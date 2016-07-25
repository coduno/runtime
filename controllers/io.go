package controllers

import (
	"log"
	"encoding/json"
	"net/http"

	"github.com/coduno/runtime/runner"
)

func init() {
	router.Handle("/io", Adapt(Wrap(ioRun), Stdin(), Test(), Files(true), Language(supportedLanguages), Method("POST")))
}

func ioRun(rd requestData, w http.ResponseWriter, r *http.Request) {
	log.Println("[controllers] [io.go] ioRun")
	image := "coduno/fingerprint-" + rd.language
	log.Println("[controllers] [io.go] spinning up docker container", image)
	tr, err := runner.IORun(rd.ball, rd.test, rd.stdin, image)
	if err != nil {
		http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tr)
}
