package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/coduno/runtime-dummy/runner"
)

func init() {
	router.HandleFunc("/diff", diff)
	router.HandleFunc("/io", iorun)
}

func diff(w http.ResponseWriter, r *http.Request) {
	rp := requestParams{
		method:   "POST",
		files:    true,
		language: true,
		test:     true,
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
		http.Error(w, "maketar error:"+err.Error(), http.StatusInternalServerError)
		return
	}

	tr, err := runner.OutMatchDiffRun(ball, rd.test, image)
	if err != nil {
		http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tr)
}

func iorun(w http.ResponseWriter, r *http.Request) {
	rp := requestParams{
		method:   "POST",
		files:    true,
		language: true,
		test:     true,
		stdin:    true,
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
		http.Error(w, "maketar error:"+err.Error(), http.StatusInternalServerError)
		return
	}

	tr, err := runner.IODiffRun(ball, rd.test, rd.stdin, image)
	if err != nil {
		http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tr)
}
