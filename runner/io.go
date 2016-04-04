package runner

import (
	"io"

	"github.com/coduno/runtime/model"
)

func IORun(ball, test, stdin io.Reader, image string) (*model.DiffTestResult, error) {
	runner := &BestDockerRunner{
		config: dockerConfig{
			image:     image,
			openStdin: true,
			stdinOnce: true,
		},
	}
	if err := runner.createContainer(); err != nil {
		return nil, err
	}
	if err := runner.upload(ball); err != nil {
		return nil, err
	}
	if err := runner.start(); err != nil {
		return nil, err
	}
	if err := runner.attach(stdin); err != nil {
		return nil, err
	}
	if err := runner.wait(); err != nil {
		return nil, err
	}
	str, err := runner.logs()
	if err != nil {
		return nil, err
	}
	dtr, err := processDiffResults(&model.DiffTestResult{SimpleTestResult: *str}, test)
	if err != nil {
		return nil, err
	}
	return dtr, nil
}
