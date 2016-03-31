package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/coduno/runtime/runner"
)

func init() {
	router.Handle("/cccdronetest", Adapt(Wrap(droneCccTest), Files(false), Language(supportedLanguages), Method("POST")))
	router.Handle("/cccdronerun", Adapt(Wrap(droneCccRun), Files(true), Language(supportedLanguages), Method("POST")))
}

func droneCccTest(rd requestData, w http.ResponseWriter, r *http.Request) {
	params, err := cccParams(r, rd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if params.Validate {
		tr, err := runner.CCCValidate(rd.ball, params)
		if err != nil {
			http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("validate response: %#v", tr)
		json.NewEncoder(w).Encode(tr)
		return
	}
	ball, err := readermaketar(rd.ball, fileNames[rd.language])
	if err != nil {
		http.Error(w, "reader maketar error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tr, err := runner.CCCTest(ball, params)
	if err != nil {
		http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("validate response: %#v", tr)
	json.NewEncoder(w).Encode(tr)
}

func droneCccRun(rd requestData, w http.ResponseWriter, r *http.Request) {
	params, err := cccParams(r, rd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tr, err := runner.CCCRun(rd.ball, params)
	if err != nil {
		http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tr)
}

func cccParams(r *http.Request, rd requestData) (p runner.CCCParams, err error) {
	p.Image = "coduno/fingerprint-" + rd.language
	if r.FormValue("test") == "" {
		return p, errors.New("invalid test value")
	}
	test, err := strconv.Atoi(r.FormValue("test"))
	if err != nil {
		return p, err
	}
	if test < 1 || test > 4 {
		return p, errors.New("invalid test value")
	}
	p.Test = strconv.Itoa(test)
	if r.FormValue("level") == "" {
		return p, errors.New("invalid level value")
	}
	level, err := strconv.Atoi(r.FormValue("level"))
	if err != nil {
		return p, err
	}
	if level < 1 || level > 7 {
		return p, errors.New("invalid level value")
	}
	p.Level = strconv.Itoa(level)
	if r.FormValue("output_test") == "true" {
		p.Validate = true
	}

	return
}
