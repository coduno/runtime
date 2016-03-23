package runner

import (
	"io"

	"github.com/coduno/runtime/model"
)

func SimpleRun(ball io.Reader, image string) (testResult model.SimpleTestResult, err error) {
	runner := &BestDockerRunner{
		config: dockerConfig{
			image: image,
		}}
	if err = runner.createContainer(); err != nil {
		return
	}
	if err = runner.upload(ball); err != nil {
		return
	}
	if err = runner.start(); err != nil {
		return
	}
	if err = runner.wait(); err != nil {
		return
	}
	return runner.logs()
}
