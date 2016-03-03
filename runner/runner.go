package runner

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fsouza/go-dockerclient"
)

var dc *docker.Client

type waitResult struct {
	ExitCode int
	Err      error
}

func init() {
	var err error
	// path := os.Getenv("DOCKER_CERT_PATH")
	path := "C:\\Users\\vbalan\\.docker\\machine\\machines\\default"
	ca := fmt.Sprintf("%s/ca.pem", path)
	cert := fmt.Sprintf("%s/cert.pem", path)
	key := fmt.Sprintf("%s/key.pem", path)
	dc, err = docker.NewTLSClient("http://192.168.99.100:2376", cert, key, ca)
	if err != nil {
		panic(err)
	}
}

func prepareImage(name string) (err error) {
	if _, err = dc.InspectImage(name); err == nil {
		return nil
	}

	if err != docker.ErrNoSuchImage {
		return
	}

	err = dc.PullImage(docker.PullImageOptions{
		Repository:   name,
		OutputStream: os.Stderr,
	}, docker.AuthConfiguration{})
	return
}

func itoc(image string) (*docker.Container, error) {
	// TODO(victorbalan): Pass the memory limit from test
	return dc.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: image,
		},
		HostConfig: &docker.HostConfig{
			Privileged:  false,
			NetworkMode: "bridge",
			Memory:      0, // TODO(flowlo): Limit memory
		},
	})
}

func waitForContainer(cID string) (err error) {
	waitc := make(chan waitResult)
	go func() {
		exitCode, err := dc.WaitContainer(cID)
		waitc <- waitResult{exitCode, err}
	}()

	var res waitResult
	select {
	case res = <-waitc:
	case <-time.After(time.Minute):
		err = errors.New("execution timed out")
		return
	}

	return res.Err
}

func getLogs(cID string) (stdout, stderr *bytes.Buffer, err error) {
	stdout = new(bytes.Buffer)
	stderr = new(bytes.Buffer)
	err = dc.Logs(docker.LogsOptions{
		Container:    cID,
		OutputStream: stdout,
		ErrorStream:  stderr,
		Stdout:       true,
		Stderr:       true,
	})
	return
}
