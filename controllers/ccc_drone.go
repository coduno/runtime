package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/coduno/runtime/model"
	"github.com/coduno/runtime/runner"
)

func init() {
	router.Handle("/drones/test", Adapt(Wrap(dispatch("flowlo/coduno:drones", droneCcc)), Files(false), Method("POST")))
	router.Handle("/drones/run", Adapt(Wrap(dispatch("flowlo/coduno:drones", droneCcc)), Files(true), Language(supportedLanguages), Method("POST")))

	router.Handle("/drones-2d/test", Adapt(Wrap(dispatch("flowlo/coduno:drones-2d", droneCcc)), Files(false), Method("POST")))
	router.Handle("/drones-2d/run", Adapt(Wrap(dispatch("flowlo/coduno:drones-2d", droneCcc)), Files(true), Language(supportedLanguages), Method("POST")))
}

func dispatch(simulatorImage string, handler func(rd requestData, w http.ResponseWriter, params *runner.CCCParams)) func(requestData, http.ResponseWriter, *http.Request) {
	return func(rd requestData, w http.ResponseWriter, r *http.Request) {
		rd.language = r.FormValue("language")
		params, err := cccParams(r, rd, simulatorImage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		handler(rd, w, params)
	}
}

func droneCcc(rd requestData, w http.ResponseWriter, params *runner.CCCParams) {
	var ts *model.TestStats
	var err error

	if params.Validate {
		ts, err = runner.CCCValidate(rd.ball, params)
	} else {
		ts, err = runner.CCCTest(rd.ball, params)
	}

	if err != nil {
		http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Responding: %#v\n", ts)
	json.NewEncoder(w).Encode(ts)
}

func cccParams(r *http.Request, rd requestData, simulatorImage string) (*runner.CCCParams, error) {
	validate, _ := strconv.ParseBool(r.FormValue("validate"))
	p := &runner.CCCParams{
		Image:          "coduno/fingerprint-" + rd.language,
		SimulatorImage: simulatorImage,
		Validate:       validate,
		Test:           1,
	}

	var err error
	p.Level, err = strconv.Atoi(r.FormValue("level"))
	if err != nil {
		return nil, err
	}
	if p.Level < 1 || p.Level > 7 {
		return nil, errors.New("invalid level value: expected integer greater than zero and smaller than eight")
	}

	testStr := r.FormValue("test")

	if testStr == "" {
		return p, nil
	}

	p.Test, err = strconv.Atoi(testStr)
	if err != nil {
		return nil, err
	}
	if p.Test < 1 || p.Test > 3 {
		return nil, errors.New("invalid level value: expected integer greater than one and smaller than four")
	}

	return p, nil
}
