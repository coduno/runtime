package runner

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/coduno/runtime/env"
	"github.com/coduno/runtime/model"
	"github.com/fsouza/go-dockerclient"
)

var dc *docker.Client

type dockerConfig struct {
	image           string
	entrypoint      []string
	cmd             []string
	networkMode     string
	openStdin       bool
	stdinOnce       bool
	publishAllPorts bool
	links           []string
}

type waitResult struct {
	ExitCode int
	Err      error
}

func init() {
	var err error
	if !env.IsDevAppServer() {
		dc, err = docker.NewClientFromEnv()
	} else {
		path := os.Getenv("DOCKER_CERT_PATH")
		ca := fmt.Sprintf("%s/ca.pem", path)
		cert := fmt.Sprintf("%s/cert.pem", path)
		key := fmt.Sprintf("%s/key.pem", path)
		dc, err = docker.NewTLSClient("http://192.168.99.100:2376", cert, key, ca)
	}
	if err != nil {
		panic(err)
	}
}

type runInfo struct {
	start, end time.Time
}
type BestDockerRunner struct {
	c      *docker.Container
	config dockerConfig
	info   runInfo
}

func (r *BestDockerRunner) prepare() (err error) {
	if _, err = dc.InspectImage(r.config.image); err == nil {
		return nil
	}

	if err != docker.ErrNoSuchImage {
		return
	}

	err = dc.PullImage(docker.PullImageOptions{
		Repository:   r.config.image,
		OutputStream: os.Stderr,
	}, docker.AuthConfiguration{})
	if err != nil {
		err = errors.New("runner.prepare: " + err.Error())
	}
	return
}

func (r *BestDockerRunner) createContainer() (err error) {
	if err := r.prepare(); err != nil {
		return err
	}
	// TODO(victorbalan): Pass the memory limit from test
	r.c, err = dc.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:      r.config.image,
			OpenStdin:  r.config.openStdin,
			StdinOnce:  r.config.stdinOnce,
			Entrypoint: r.config.entrypoint,
			Cmd:        r.config.cmd,
			Env:        docker.Env([]string{"CODUNO=true"}),
		},
		HostConfig: &docker.HostConfig{
			Privileged:      false,
			NetworkMode:     r.config.networkMode,
			Links:           r.config.links,
			PublishAllPorts: r.config.publishAllPorts,
			Memory:          64 >> 20, // That's 64MiB.
		},
	})
	return err
}

func (r *BestDockerRunner) upload(ball io.Reader) (err error) {
	err = dc.UploadToContainer(r.c.ID, docker.UploadToContainerOptions{
		Path:        "/run",
		InputStream: ball,
	})
	if err != nil {
		err = errors.New("runner.upload: " + err.Error())
	}
	return
}

func (r *BestDockerRunner) start() (err error) {
	r.info.start = time.Now()
	err = dc.StartContainer(r.c.ID, r.c.HostConfig)
	if err != nil {
		err = errors.New("runner.start: " + err.Error())
	}
	return
}

func (r *BestDockerRunner) attach(stream io.Reader) (err error) {
	err = dc.AttachToContainer(docker.AttachToContainerOptions{
		Container:   r.c.ID,
		InputStream: stream,
		Stdin:       true,
		Stream:      true,
	})
	if err != nil {
		err = errors.New("runner.attach: " + err.Error())
	}
	return
}

func (r *BestDockerRunner) wait() (err error) {
	waitc := make(chan waitResult)
	go func() {
		exitCode, err := dc.WaitContainer(r.c.ID)
		waitc <- waitResult{exitCode, err}
	}()

	var res waitResult
	select {
	case res = <-waitc:
	case <-time.After(time.Minute):
		return errors.New("execution timed out")
	}

	if res.Err != nil {
		return errors.New("runner.wait: " + res.Err.Error())
	}
	r.info.end = time.Now()
	return nil
}

func (r *BestDockerRunner) logs() (str model.SimpleTestResult, err error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	err = dc.Logs(docker.LogsOptions{
		Container:    r.c.ID,
		OutputStream: stdout,
		ErrorStream:  stderr,
		Stdout:       true,
		Stderr:       true,
	})
	if err != nil {
		err = errors.New("runner.logs: " + err.Error())
		return
	}

	return model.SimpleTestResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Start:  r.info.start,
		End:    r.info.end,
	}, nil
}

func (r *BestDockerRunner) inspect() error {
	c, err := dc.InspectContainer(r.c.ID)
	if err != nil {
		return errors.New("runner.inspect: " + err.Error())
	}
	r.c = c
	return nil
}
