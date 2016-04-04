package runner

import (
	"io"

	"github.com/coduno/runtime/model"
)

func SimpleRun(ball io.Reader, image string) (*model.SimpleTestResult, error) {
	runner := &BestDockerRunner{
		config: dockerConfig{
			image: image,
		}}
	if err := runner.createContainer(); err != nil {
		return nil, err
	}
	if err := runner.upload(ball); err != nil {
		return nil, err
	}
	if err := runner.start(); err != nil {
		return nil, err
	}
	if err := runner.wait(); err != nil {
		return nil, err
	}
	return runner.logs()
}
