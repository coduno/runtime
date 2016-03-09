package runner

import (
	"bytes"
	"io"
	"time"

	"github.com/coduno/runtime-dummy/model"
	"github.com/fsouza/go-dockerclient"
)

func IORun(ball, test, stdin io.Reader, image string) (tr model.DiffTestResult, err error) {
	if err = prepareImage(image); err != nil {
		return
	}

	var c *docker.Container
	c, err = dc.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:     image,
			OpenStdin: true,
			StdinOnce: true,
		},
		HostConfig: &docker.HostConfig{
			Privileged:  false,
			NetworkMode: "none",
			Memory:      0, // TODO(flowlo): Limit memory
		},
	})
	if err != nil {
		return
	}

	err = dc.UploadToContainer(c.ID, docker.UploadToContainerOptions{
		Path:        "/run",
		InputStream: ball,
	})
	if err != nil {
		return
	}

	start := time.Now()
	if err = dc.StartContainer(c.ID, c.HostConfig); err != nil {
		return
	}

	err = dc.AttachToContainer(docker.AttachToContainerOptions{
		Container:   c.ID,
		InputStream: stdin,
		Stdin:       true,
		Stream:      true,
	})
	if err != nil {
		return
	}

	if err = waitForContainer(c.ID); err != nil {
		return
	}
	end := time.Now()

	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	if stdout, stderr, err = getLogs(c.ID); err != nil {
		return
	}

	tr = model.DiffTestResult{
		SimpleTestResult: model.SimpleTestResult{
			Stdout: stdout.String(),
			Stderr: stderr.String(),
			Start:  start,
			End:    end,
		},
		Endpoint: "diff-result",
	}
	processDiffResults(&tr, test)
	return
}
