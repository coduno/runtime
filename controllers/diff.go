package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/coduno/runtime/runner"
	"github.com/coduno/runtime/model"
)

func init() {
	router.Handle("/diff", Adapt(Wrap(diffRun), Test(), Files(true), Language(supportedLanguages), Method("POST")))
}

func diffRun(rd requestData, w http.ResponseWriter, r *http.Request) {
	var tr *model.DiffTestResult
	var err error
	if  validate, _ := strconv.ParseBool(r.FormValue("validate")); validate {
		tr, err = runner.DiffValidate(rd.ball, rd.test)
	} else {
		image := "coduno/fingerprint-" + rd.language
		tr, err = runner.DiffRun(rd.ball, rd.test, image)
	}
	if err != nil {
		http.Error(w, "run error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tr)
}
