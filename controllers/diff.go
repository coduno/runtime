package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/coduno/runtime-dummy/runner"
)

func init() {
	router.HandleFunc("/diff", diff)
	router.HandleFunc("/io", iorun)
}

func diff(w http.ResponseWriter, r *http.Request) {
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
	test := r.FormValue("path")
	if test == "" {
		http.Error(w, "you must input a test for diff runner", http.StatusBadRequest)
		return
	}

	image := "coduno/fingerprint-" + language
	ball, err := maketar(files[0], fileName)
	if err != nil {
		http.Error(w, "maketar error:"+err.Error(), http.StatusInternalServerError)
		return
	}
	testBall, err := getFile(test)
	if err != nil {
		http.Error(w, "get test file error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tr, err := runner.OutMatchDiffRun(ball, testBall, image)
	if err != nil {
		http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tr)
}

func iorun(w http.ResponseWriter, r *http.Request) {
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
	test := r.FormValue("path")
	if test == "" {
		http.Error(w, "you must input a test for io runner", http.StatusBadRequest)
		return
	}
	stdin := r.FormValue("stdin")
	if stdin == "" {
		http.Error(w, "you must input stdin for io runner", http.StatusBadRequest)
		return
	}

	image := "coduno/fingerprint-" + language
	ball, err := maketar(files[0], fileName)
	if err != nil {
		http.Error(w, "maketar error:"+err.Error(), http.StatusInternalServerError)
		return
	}
	testBall, err := getFile(test)
	if err != nil {
		http.Error(w, "get test file error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	stdinReader, err := getFile(stdin)
	if err != nil {
		http.Error(w, "get stdin file error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tr, err := runner.IODiffRun(ball, testBall, stdinReader, image)
	if err != nil {
		http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tr)
}

func getFile(filename string) (io.Reader, error) {
	b, err := ioutil.ReadFile("testfiles/" + filename)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}
