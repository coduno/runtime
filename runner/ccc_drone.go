package runner

import (
	"io"

	"github.com/coduno/runtime/model"
)

type CCCParams struct {
	Image, Level, Test string
	Validate           bool
}

func CCCValidate(ball io.Reader, p CCCParams) (ts model.TestStats, err error) {
	runner := &BestDockerRunner{
		config: dockerConfig{
			image:       "flowlo/coduno:simulator",
			openStdin:   true,
			stdinOnce:   true,
			networkMode: "none",
			cmd:         []string{p.Level, p.Test, "0"},
		}}
	if err = runner.createContainer(); err != nil {
		return
	}
	if err = runner.start(); err != nil {
		return
	}
	if err = runner.attach(ball); err != nil {
		return
	}
	if err = runner.wait(); err != nil {
		return
	}
	str, err := runner.logs()
	if err != nil {
		return ts, err
	}
	if err = runner.inspect(); err != nil {
		return
	}

	// NOTE(flowlo): Errors preventing removal are ignored.
	runner.remove()

	return model.TestStats{
		Failed: runner.c.State.ExitCode != 0,
		Stdout: str.Stdout,
		Stderr: str.Stderr,
	}, nil
}

func CCCTest(ball io.Reader, p CCCParams) (ts model.TestStats, err error) {
	var str model.SimpleTestResult
	ccc := &BestDockerRunner{
		config: dockerConfig{
			image:           "flowlo/coduno:simulator",
			cmd:             []string{p.Level, p.Test, "7000"},
			publishAllPorts: true,
		}}

	str, err = normalCCCRun(ccc, ball, p.Image)
	if err != nil {
		return ts, err
	}
	if err = ccc.wait(); err != nil {
		return
	}
	if err = ccc.inspect(); err != nil {
		return
	}

	// NOTE(flowlo): Errors preventing removal are ignored.
	ccc.remove()

	return model.TestStats{
		Failed: ccc.c.State.ExitCode != 0,
		Stdout: str.Stdout,
		Stderr: str.Stderr,
	}, nil
}

func CCCRun(ball io.Reader, p CCCParams) (testResult model.SimpleTestResult, err error) {
	ccc := &BestDockerRunner{
		config: dockerConfig{
			image:           "flowlo/coduno:simulator",
			cmd:             []string{p.Level, "1", "7000"},
			publishAllPorts: true,
		}}
	return normalCCCRun(ccc, ball, p.Image)
}

func normalCCCRun(ccc *BestDockerRunner, ball io.Reader, image string) (testResult model.SimpleTestResult, err error) {
	if err = ccc.createContainer(); err != nil {
		return
	}
	if err = ccc.start(); err != nil {
		return
	}
	if err = ccc.inspect(); err != nil {
		return
	}
	runner := &BestDockerRunner{
		config: dockerConfig{
			image:     image,
			openStdin: true,
			stdinOnce: true,
			links:     []string{ccc.c.ID + ":simulator"},
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

	// NOTE(flowlo): Errors preventing removal are ignored.
	runner.remove()

	return runner.logs()
}
