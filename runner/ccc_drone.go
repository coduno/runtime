package runner

import (
	"io"
	"strconv"

	"github.com/coduno/runtime/model"
)

type CCCParams struct {
	Image, SimulatorImage string
	Level, Test           int
	Validate              bool
}

func CCCValidate(ball io.Reader, p *CCCParams) (*model.TestStats, error) {
	runner := &BestDockerRunner{
		config: dockerConfig{
			image:       p.SimulatorImage,
			openStdin:   true,
			stdinOnce:   true,
			networkMode: "none",
			cmd:         []string{strconv.Itoa(p.Level), strconv.Itoa(p.Test), "0"},
		},
	}

	if err := runner.createContainer(); err != nil {
		return nil, err
	}
	if err := runner.start(); err != nil {
		return nil, err
	}
	if err := runner.attach(ball); err != nil {
		return nil, err
	}
	if err := runner.wait(); err != nil {
		return nil, err
	}

	str, err := runner.logs()
	if err != nil {
		return nil, err
	}
	if err := runner.inspect(); err != nil {
		return nil, err
	}

	// NOTE(flowlo): Errors preventing removal are ignored.
	runner.remove()

	return &model.TestStats{
		Failed: runner.c.State.ExitCode != 0,
		Stdout: str.Stdout,
		Stderr: str.Stderr,
	}, nil
}

func CCCTest(ball io.Reader, p *CCCParams) (*model.TestStats, error) {
	ccc := &BestDockerRunner{
		config: dockerConfig{
			image:           p.SimulatorImage,
			cmd:             []string{strconv.Itoa(p.Level), strconv.Itoa(p.Test), "7000"},
			publishAllPorts: true,
		}}

	str, err := normalCCCRun(ccc, ball, p.Image)
	if err != nil {
		return nil, err
	}
	if err := ccc.wait(); err != nil {
		return nil, err
	}
	if err := ccc.inspect(); err != nil {
		return nil, err
	}

	// NOTE(flowlo): Errors preventing removal are ignored.
	ccc.remove()

	return &model.TestStats{
		Failed: ccc.c.State.ExitCode != 0,
		Stdout: str.Stdout,
		Stderr: str.Stderr,
	}, nil
}

func CCCRun(ball io.Reader, p *CCCParams) (*model.SimpleTestResult, error) {
	return normalCCCRun(&BestDockerRunner{
		config: dockerConfig{
			image:           p.SimulatorImage,
			cmd:             []string{strconv.Itoa(p.Level), "1", "7000"},
			publishAllPorts: true,
		},
	}, ball, p.Image)
}

func normalCCCRun(ccc *BestDockerRunner, ball io.Reader, image string) (*model.SimpleTestResult, error) {
	if err := ccc.createContainer(); err != nil {
		return nil, err
	}
	if err := ccc.start(); err != nil {
		return nil, err
	}
	if err := ccc.inspect(); err != nil {
		return nil, err
	}
	runner := &BestDockerRunner{
		config: dockerConfig{
			image:     image,
			openStdin: true,
			stdinOnce: true,
			links:     []string{ccc.c.ID + ":simulator"},
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

	// NOTE(flowlo): Errors preventing removal are ignored.
	defer runner.remove()

	return runner.logs()
}
