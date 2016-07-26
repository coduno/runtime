package runner

import (
	"io"

	"github.com/coduno/runtime/model"
	"github.com/fsouza/go-dockerclient"
)

func SimpleRun(ball io.Reader, image string) (*model.SimpleTestResult, error) {
	runner := &Runner{
		config: &docker.Config{
			Image: image,
		},
	}

	return runner.
		createContainer().
		upload(ball).
		start().
		wait().
		logs()
}
