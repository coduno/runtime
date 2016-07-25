package runner

import (
	"io"
	"log"

	"github.com/coduno/runtime/model"
	"github.com/fsouza/go-dockerclient"
)

func IORun(ball, test, stdin io.Reader, image string) (*model.DiffTestResult, error) {
	runner := &BestDockerRunner{
		config: &docker.Config{
			Image:     image,
			OpenStdin: true,
			StdinOnce: true,
		},
	}

	str, err := runner.
		createContainer().
		upload(ball).
		start().
		attach(stdin).
		wait().
		logs()

	if err != nil {
		return nil, err
	}

	log.Println("[runner] [io.go] IORun")
	
	return processDiffResults(&model.DiffTestResult{SimpleTestResult: *str}, test)
}
