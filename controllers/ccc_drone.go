package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/coduno/runtime-dummy/runner"
)

func init() {
	router.HandleFunc("/cccdrone", droneCcc)
}

func droneCcc(w http.ResponseWriter, r *http.Request) {
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
	testPath := r.FormValue("test")
	if testPath == "" {
		tr, err := runner.CCCRun(ball, image)
		if err != nil {
			http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(tr)
		return
	}

	test, err := getFile(testPath)
	if err != nil {
		http.Error(w, "get test file error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tr, err := runner.CCCRunWithOutput(ball, test, image)
	if err != nil {
		http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tr)
}
