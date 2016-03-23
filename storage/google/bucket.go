package google

import "github.com/coduno/runtime/env"

func SubmissionsBucket() string {
	if env.IsDevAppServer() {
		return "coduno-dev"
	}
	return "coduno-submissions"
}

func TestsBucket() string {
	return "coduno-tests"
}
