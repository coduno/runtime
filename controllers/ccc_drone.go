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
	router.Handle("/drones/test", Adapt(Wrap(dispatch("drones", droneCccTest)), Files(false), Language(supportedLanguages), Method("POST")))
	router.Handle("/drones/run", Adapt(Wrap(dispatch("drones", droneCccRun)), Files(true), Language(supportedLanguages), Method("POST")))

	router.Handle("/drones-2d/test", Adapt(Wrap(dispatch("drones-2d", droneCccTest)), Files(false), Language(supportedLanguages), Method("POST")))
	router.Handle("/drones-2d/run", Adapt(Wrap(dispatch("drones-2d", droneCccRun)), Files(true), Language(supportedLanguages), Method("POST")))
}

func dispatch(simulatorImage string, handler func(rd requestData, w http.ResponseWriter, params *runner.CCCParams)) func(requestData, http.ResponseWriter, *http.Request) {
	return func(rd requestData, w http.ResponseWriter, r *http.Request) {
		params, err := cccParams(r, rd, simulatorImage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		handler(rd, w, params)
	}
}

func droneCccTest(rd requestData, w http.ResponseWriter, params *runner.CCCParams) {
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

func droneCccRun(rd requestData, w http.ResponseWriter, params *runner.CCCParams) {
	tr, err := runner.CCCRun(rd.ball, params)
	if err != nil {
		http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tr)
}

func cccParams(r *http.Request, rd requestData, simulatorImage string) (*runner.CCCParams, error) {
	var err error

	// TODO(flowlo): Maybe rename parameter "output_test" to "validate"?
	p := &runner.CCCParams{
		Image:          "coduno/fingerprint-" + rd.language,
		SimulatorImage: simulatorImage,
		Validate:       r.FormValue("output_test") == "true",
	}

	p.Test, err = strconv.Atoi(r.FormValue("test"))
	if err != nil {
		return nil, err
	}
	if p.Test < 1 || p.Test > 4 {
		return p, errors.New("invalid test value: expected integer greater than zero and smaller than five")
	}

	p.Level, err = strconv.Atoi(r.FormValue("level"))
	if err != nil {
		return nil, err
	}
	if p.Level < 1 || p.Level > 7 {
		return nil, errors.New("invalid level value: expected integer greater than zero and smaller than eight")
	}

	return p, nil
}
