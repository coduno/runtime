package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/coduno/runtime-dummy/runner"
)

func init() {
	router.Handle("/cccdronetest", Adapt(Wrap(droneCccTest), Test(), Files(), Language(supportedLanguages), Method("POST")))
	router.Handle("/cccdronerun", Adapt(Wrap(droneCccRun), Files(), Language(supportedLanguages), Method("POST")))
}

func droneCccTest(rd requestData, w http.ResponseWriter, r *http.Request) {
	image := "coduno/fingerprint-" + rd.language
	tr, err := runner.CCCRunWithOutput(rd.ball, rd.test, image)
	if err != nil {
		http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tr)
}

func droneCccRun(rd requestData, w http.ResponseWriter, r *http.Request) {
	image := "coduno/fingerprint-" + rd.language
	tr, err := runner.CCCRun(rd.ball, image)
	if err != nil {
		http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tr)
}
