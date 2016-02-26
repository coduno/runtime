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
	if r.Method != "POST" {
		http.Error(w, r.Method+" not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(16 << 20); err != nil {
		http.Error(w, "could not parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	files, ok := r.MultipartForm.File["files"]
	if !ok {
		http.Error(w, "missing files", http.StatusBadRequest)
		return
	}
	if len(files) != 1 {
		http.Error(w, "we currently support only single file uploads", http.StatusBadRequest)
		return
	}
	language := r.FormValue("language")
	fileName, found := fileNames[language]
	if !found {
		http.Error(w, "we currently don`t support "+language, http.StatusBadRequest)
		return
	}

	image := "coduno/fingerprint-" + language
	ball, err := maketar(files[0], fileName)
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
