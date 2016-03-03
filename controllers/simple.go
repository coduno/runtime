package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/coduno/runtime-dummy/runner"
)

func init() {
	router.HandleFunc("/simple", run)
}

func run(w http.ResponseWriter, r *http.Request) {
	rp := requestParams{
		method:   "POST",
		files:    true,
		language: true,
	}
	rd, ok := processRequest(w, r, rp)
	if !ok {
		return
	}
	fileName, found := fileNames[rd.language]
	if !found {
		http.Error(w, "we currently don`t support "+rd.language, http.StatusBadRequest)
		return
	}

	image := "coduno/fingerprint-" + rd.language
	ball, err := maketar(rd.files[0], fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tr, err := runner.SimpleRun(ball, image)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tr)
}
