package runner

import (
	"bytes"
	"io"
	"time"

	"github.com/coduno/runtime-dummy/model"
	"github.com/fsouza/go-dockerclient"
)

func SimpleRun(ball io.Reader, image string) (testResult model.SimpleTestResult, err error) {
	if err = prepareImage(image); err != nil {
		panic(err)
	}
	var c *docker.Container
	if c, err = itoc(image); err != nil {
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

	if err = waitForContainer(c.ID); err != nil {
		return
	}
	end := time.Now()

	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	if stdout, stderr, err = getLogs(c.ID); err != nil {
		return
	}

	testResult = model.SimpleTestResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Start:  start,
		End:    end,
	}

	return
}
