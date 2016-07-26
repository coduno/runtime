package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/coduno/runtime/runner"
	"github.com/fsouza/go-dockerclient"
)

var supportedLanguages = []string{"py", "c", "cpp", "java", "csharp", "js", "php", "go", "groovy", "scala", "pascal"}

func init() {
	router.Handle("/simple", Adapt(Wrap(simpleRun), Files(true), Language(supportedLanguages), Method("POST")))
}

func simpleRun(rd requestData, w http.ResponseWriter, r *http.Request) {

	runner := &runner.Runner{
		Config: &docker.Config{
			Image: "coduno/fingerprint-" + rd.language,
		},
	}

	tr, err := runner.
		CreateContainer().
		Upload(rd.ball).
		Start().
		Wait().
		Logs()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tr)
}
