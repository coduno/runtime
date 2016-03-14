package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/coduno/runtime-dummy/runner"
)

func init() {
	router.Handle("/diff", Adapt(Wrap(diffRun), Test(), Files(), Language(supportedLanguages), Method("POST")))
}

func diffRun(rd requestData, w http.ResponseWriter, r *http.Request) {
	image := "coduno/fingerprint-" + rd.language
	tr, err := runner.DiffRun(rd.ball, rd.test, image)
	if err != nil {
		http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tr)
}
