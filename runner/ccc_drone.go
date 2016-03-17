package runner

import (
	"bytes"
	"io"

	"github.com/coduno/runtime-dummy/model"
)

type CCCParams struct {
	Image, Level, Test string
	Validate           bool
}

func cccValidate(ball io.Reader, p CCCParams) (ts model.TestStats, err error) {
	runner := &BestDockerRunner{
		config: dockerConfig{
			image:       "flowlo/coduno:simulator",
			openStdin:   true,
			stdinOnce:   true,
			networkMode: "none",
			entrypoint:  []string{"/bin/bash", "-c", "java -jar simulator.jar " + "0 " + p.Level + " " + p.Test},
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
	return model.TestStats{
		Failed: runner.c.State.ExitCode == 0,
		Stdout: str.Stdout,
		Stderr: str.Stderr,
	}, nil
}

func CCCTest(ball io.Reader, p CCCParams) (ts model.TestStats, err error) {
	if p.Validate {
		return cccValidate(ball, p)
	} else {
		var str model.SimpleTestResult
		ccc := &BestDockerRunner{
			config: dockerConfig{
				image:       "flowlo/coduno:simulator",
				networkMode: "bridge",
				entrypoint:  []string{"/bin/bash", "-c", "java -jar simulator.jar " + "7000 " + p.Level + p.Test},
			}}
		str, err = normalCCCRun(ccc, ball, p.Image)
		if err != nil {
			return ts, err
		}
		if err = ccc.inspect(); err != nil {
			return
		}
		return model.TestStats{
			Failed: ccc.c.State.ExitCode == 0,
			Stdout: str.Stdout,
			Stderr: str.Stderr,
		}, nil
	}

}

func CCCRun(ball io.Reader, p CCCParams) (testResult model.SimpleTestResult, err error) {
	ccc := &BestDockerRunner{
		config: dockerConfig{
			image:       "flowlo/coduno:simulator",
			networkMode: "bridge",
			entrypoint:  []string{"/bin/bash", "-c", "java -jar simulator.jar " + "7000 " + p.Level + " 1"},
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
			image:       image,
			networkMode: "bridge",
			openStdin:   true,
			stdinOnce:   true,
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
	network := ccc.c.NetworkSettings.Networks["bridge"]
	if err = runner.attach(bytes.NewReader([]byte(network.IPAddress))); err != nil {
		return
	}
	if err = runner.wait(); err != nil {
		return
	}

	return runner.logs()
}
