package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/coduno/runtime/runner"
)

var supportedLanguages = []string{"py", "c", "cpp", "java", "csharp", "js", "php", "go", "groovy", "scala"}

func init() {
	router.Handle("/simple", Adapt(Wrap(simpleRun), Files(true), Language(supportedLanguages), Method("POST")))
}

func simpleRun(rd requestData, w http.ResponseWriter, r *http.Request) {
	image := "coduno/fingerprint-" + rd.language
	tr, err := runner.SimpleRun(rd.ball, image)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tr)
}
