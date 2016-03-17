package runner

import (
	"io"

	"github.com/coduno/runtime-dummy/model"
)

func IORun(ball, test, stdin io.Reader, image string) (tr model.DiffTestResult, err error) {
	runner := &BestDockerRunner{
		config: dockerConfig{
			image:     image,
			openStdin: true,
			stdinOnce: true,
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
	if err = runner.attach(stdin); err != nil {
		return
	}
	if err = runner.wait(); err != nil {
		return
	}
	str, err := runner.logs()
	if err != nil {
		return tr, err
	}

	tr = model.DiffTestResult{
		SimpleTestResult: str,
	}
	processDiffResults(&tr, test)
	return
}
