package runner

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
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
	runner := &Runner{
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

	runner.
		createContainer().
		start().
		attach(ball).
		wait().
		inspect().
		remove()

	if runner.err != nil {
		return nil, runner.err
	}

	return &model.TestStats{
		Successful: runner.c.State.ExitCode == 0,
	}, nil
}

func CCCTest(ball io.Reader, p *CCCParams) (*model.TestStats, error) {
	simulator := &Runner{
		config: &docker.Config{
			Image: p.SimulatorImage,
			Cmd:   []string{strconv.Itoa(p.Level), strconv.Itoa(p.Test), "7000"},
		},
		hostConfig: &docker.HostConfig{},
	}

	simulator.createContainer().start()
	if simulator.err != nil {
		return nil, simulator.err
	}

	runner := &Runner{
		config: &docker.Config{
			Image: p.Image,
		},
		hostConfig: &docker.HostConfig{
			Links: []string{simulator.c.ID + ":simulator"},
		},
	}

	tr, err := runner.
		createContainer().
		upload(ball).
		start().
		wait().
		logs()

	if err != nil {
		return nil, err
	}

	stats := new(bytes.Buffer)
	err = runner.download("/run/stats.log", stats)
	var statsData interface{}

	if err == nil {
		err := json.NewDecoder(stats).Decode(&statsData)
		if err != nil {
			log.Printf("Error decoding stats: %s\n", err)
		}
	} else {
		log.Printf("Error getting stats: %s", err)
	}

	runner.inspect().remove()
	if runner.err != nil {
		return nil, runner.err
	}

	exitCode := 1
	if tr.Stderr == "" {
		simulator.wait().inspect().remove()
		if simulator.err != nil {
			return nil, simulator.err
		}
		exitCode = simulator.c.State.ExitCode
	}

	return &model.TestStats{
		ExitCode:   exitCode,
		Successful: exitCode == 0,
		Stdout:     tr.Stdout,
		Stderr:     tr.Stderr,
		Stats:      statsData,
	}, nil
}
