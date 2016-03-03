package runner

import (
	"bytes"
	"io"
	"time"

	"github.com/coduno/runtime-dummy/model"
	"github.com/fsouza/go-dockerclient"
)

func CCCRunWithOutput(ball, test io.Reader, image string) (ts model.TestStats, err error) {
	var str model.SimpleTestResult
	str, err = CCCRun(ball, image)
	if err != nil {
		return
	}
	tr := model.DiffTestResult{
		SimpleTestResult: str,
		Endpoint:         "diff-result",
	}

	ts, err = processDiffResults(&tr, test)
	return
}

func CCCRun(ball io.Reader, image string) (testResult model.SimpleTestResult, err error) {
	sc, err := cccDroneSimulator("level1")
	if err != nil {
		panic(err)
	}
	if err = prepareImage(image); err != nil {
		panic(err)
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
			NetworkMode: "bridge",
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

	network := sc.NetworkSettings.Networks["bridge"]

	start := time.Now()
	if err = dc.StartContainer(c.ID, c.HostConfig); err != nil {
		return
	}
	err = dc.AttachToContainer(docker.AttachToContainerOptions{
		Container:   c.ID,
		InputStream: bytes.NewReader([]byte(network.IPAddress)),
		Stdin:       true,
		Stream:      true,
	})

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

func cccDroneSimulator(level string) (c *docker.Container, err error) {
	var image = "coduno/ccc_drone_simulator"
	if err = prepareImage(image); err != nil {
		panic(err)
	}
	c, err = dc.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:      image,
			Entrypoint: []string{"/bin/bash", "-c", "java -jar simulator.jar " + level + " tcp 7000"},
		},
		HostConfig: &docker.HostConfig{
			Privileged:  false,
			NetworkMode: "bridge",
			Memory:      0, // TODO(flowlo): Limit memory
		},
	})
	if err != nil {
		return
	}
	if err = dc.StartContainer(c.ID, c.HostConfig); err != nil {
		return
	}
	sc, err := dc.InspectContainer(c.ID)
	if err != nil {
		panic(err)
	}
	return sc, nil
}
