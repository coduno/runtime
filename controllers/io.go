package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/coduno/runtime/runner"
)

func init() {
	router.Handle("/io", Adapt(Wrap(ioRun), Stdin(), Test(), Files(), Language(supportedLanguages), Method("POST")))
}

func ioRun(rd requestData, w http.ResponseWriter, r *http.Request) {
	image := "coduno/fingerprint-" + rd.language
	tr, err := runner.IORun(rd.ball, rd.test, rd.stdin, image)
	if err != nil {
		http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tr)
}
