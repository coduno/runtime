package runner

import (
	"io"
	"strconv"

	"github.com/coduno/runtime/model"
	"github.com/fsouza/go-dockerclient"
)

type CCCParams struct {
	Image, SimulatorImage string
	Level, Test           int
	Validate              bool
}

func CCCValidate(ball io.Reader, p *CCCParams) (*model.TestStats, error) {
	runner := &BestDockerRunner{
		config: &docker.Config{
			Image:     p.SimulatorImage,
			OpenStdin: true,
			StdinOnce: true,
			Cmd:       []string{strconv.Itoa(p.Level), strconv.Itoa(p.Test), "0"},
		},
		hostConfig: &docker.HostConfig{
			NetworkMode: "none",
		},
	}

	str, err := runner.
		createContainer().
		start().
		attach(ball).
		wait().
		logs()

	if err != nil {
		return nil, err
	}

	// NOTE(flowlo): Errors preventing removal are ignored.
	runner.remove()

	return &model.TestStats{
		Successful: runner.c.State.ExitCode == 0,
		Stdout:     str.Stdout,
		Stderr:     str.Stderr,
	}, nil
}

func CCCTest(ball io.Reader, p *CCCParams) (*model.TestStats, error) {
	ccc := &BestDockerRunner{
		config: &docker.Config{
			Image: p.SimulatorImage,
			Cmd:   []string{strconv.Itoa(p.Level), strconv.Itoa(p.Test), "7000"},
		},
		hostConfig: &docker.HostConfig{
			PublishAllPorts: true,
		},
	}

	str, err := normalCCCRun(ccc, ball, p.Image)
	if err != nil {
		return nil, err
	}
	ccc.wait()

	if ccc.err != nil {
		return nil, ccc.err
	}

	// NOTE(flowlo): Errors preventing removal are ignored.
	ccc.remove()

	return &model.TestStats{
		Successful: ccc.c.State.ExitCode == 0,
		Stdout:     str.Stdout,
		Stderr:     str.Stderr,
	}, nil
}

func CCCRun(ball io.Reader, p *CCCParams) (*model.SimpleTestResult, error) {
	return normalCCCRun(&BestDockerRunner{
		config: &docker.Config{
			Image: p.SimulatorImage,
			Cmd:   []string{strconv.Itoa(p.Level), "1", "7000"},
		},
		hostConfig: &docker.HostConfig{
			PublishAllPorts: true,
		},
	}, ball, p.Image)
}

func normalCCCRun(ccc *BestDockerRunner, ball io.Reader, image string) (*model.SimpleTestResult, error) {
	ccc.createContainer().start()
	if ccc.err != nil {
		return nil, ccc.err
	}

	runner := &BestDockerRunner{
		config: &docker.Config{
			Image:     image,
			OpenStdin: true,
			StdinOnce: true,
		},
		hostConfig: &docker.HostConfig{
			Links: []string{ccc.c.ID + ":simulator"},
		},
	}

	tr, err := runner.
		createContainer().
		upload(ball).
		start().
		wait().
		logs()

	// NOTE(flowlo): Errors preventing removal are ignored.
	defer runner.remove()

	return tr, err
}
