package runner

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/coduno/runtime/env"
	"github.com/coduno/runtime/model"
	"github.com/fsouza/go-dockerclient"
)

const codunoFlag = "CODUNO=true"
const memoryLimit = 64 << 20  // That's 64MiB.
const bufferLimit = 512 << 10 // That's 512KiB, or 0.5MiB.

var dc *docker.Client

type waitResult struct {
	ExitCode int
	Err      error
}

func init() {
	var err error
	if !env.IsDevAppServer() {
		dc, err = docker.NewClientFromEnv()
	} else {
		// TODO(flowlo): Is this still used?
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

func Scrape(d time.Duration) {
	for {
		time.Sleep(d / 2)
		log.Println("Scraper woke up.")

		cs, err := dc.ListContainers(docker.ListContainersOptions{
			All:     true,
			Size:    false,
			Filters: map[string][]string{"status": {"running"}},
		})

		if err != nil {
			continue
		}

		now := time.Now()
		for _, c := range cs {
			if now.Sub(time.Unix(c.Created, 0)) < d {
				continue
			}

			err := dc.RemoveContainer(docker.RemoveContainerOptions{
				ID:            c.ID,
				RemoveVolumes: true,
				Force:         true,
			})

			if err == nil {
				log.Printf("Scraper removed container %q\n", c.ID)
			} else {
				log.Printf("Scraper failed to remove container %q: %s\n", c.ID, err)
			}
		}
	}
}

type BestDockerRunner struct {
	c                *docker.Container
	config           *docker.Config
	hostConfig       *docker.HostConfig
	started, stopped time.Time
	err              error
}

func (r *BestDockerRunner) prepare() *BestDockerRunner {
	if r.err != nil {
		return r
	}

	_, err := dc.InspectImage(r.config.Image)
	if err == nil {
		return r
	}

	if err != docker.ErrNoSuchImage {
		log.Printf("Error inspecting image %q: %s\n", r.config.Image, err)
		r.err = err
		return r
	}

	r.err = dc.PullImage(docker.PullImageOptions{
		Repository:   r.config.Image,
		OutputStream: os.Stderr,
	}, docker.AuthConfiguration{})

	if r.err != nil {
		log.Printf("Error pulling image %q: %s\n", r.config.Image, r.err)
	}
	return r
}

func (r *BestDockerRunner) createContainer() *BestDockerRunner {
	if r.err != nil {
		return r
	}

	if r.config.Env == nil {
		r.config.Env = []string{codunoFlag}
	} else {
		r.config.Env = append(r.config.Env, codunoFlag)
	}

	r.hostConfig.Privileged = false

	if r.hostConfig.Memory > memoryLimit {
		r.hostConfig.Memory = memoryLimit
	}

	r.c, r.err = dc.CreateContainer(docker.CreateContainerOptions{
		Config:     r.config,
		HostConfig: r.hostConfig,
	})

	if r.err != nil {
		log.Printf("Failed to create container: %s\n", r.err)
	}
	return r
}

func (r *BestDockerRunner) upload(ball io.Reader) *BestDockerRunner {
	if r.err != nil {
		return r
	}

	r.err = dc.UploadToContainer(r.c.ID, docker.UploadToContainerOptions{
		Path:        "/run",
		InputStream: ball,
	})

	if r.err != nil {
		log.Printf("Failed to upload to container: %s\n", r.err)
	}
	return r
}

func (r *BestDockerRunner) start() *BestDockerRunner {
	if r.err != nil {
		return r
	}

	r.err = dc.StartContainer(r.c.ID, r.c.HostConfig)
	if r.err != nil {
		log.Printf("Failed to start container %q: %s\n", r.c.ID, r.err)
		return r
	}
	r.started = time.Now()
	return r
}

func (r *BestDockerRunner) attach(stream io.Reader) *BestDockerRunner {
	if r.err != nil {
		return r
	}

	r.err = dc.AttachToContainer(docker.AttachToContainerOptions{
		Container:   r.c.ID,
		InputStream: stream,
		Stdin:       true,
		Stream:      true,
	})
	if r.err != nil {
		log.Printf("Failed to attach to container %q: %s\n", r.c.ID, r.err)
	}
	return r
}

func (r *BestDockerRunner) wait() *BestDockerRunner {
	if r.err != nil {
		return r
	}

	waitc := make(chan waitResult)
	go func() {
		exitCode, err := dc.WaitContainer(r.c.ID)
		waitc <- waitResult{exitCode, err}
	}()

	var res waitResult
	select {
	case res = <-waitc:
	case <-time.After(time.Minute):
		r.err = errors.New("execution timed out")
		return r
	}

	if res.Err != nil {
		r.err = res.Err
		return r
	}
	r.stopped = time.Now()
	return r
}

func (r *BestDockerRunner) logs() (*model.SimpleTestResult, error) {
	if r.err != nil {
		return nil, r.err
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	r.err = dc.Logs(docker.LogsOptions{
		Container:    r.c.ID,
		OutputStream: stdout,
		ErrorStream:  stderr,
		Stdout:       true,
		Stderr:       true,
	})
	if r.err != nil {
		log.Printf("Failed to obtain logs from %q: %s\n", r.c.ID, r.err)
		return nil, r.err
	}

	if stdout.Len() > bufferLimit {
		stdout.Truncate(bufferLimit)
		log.Println("Truncated standard output.")
	}

	if stderr.Len() > bufferLimit {
		stderr.Truncate(bufferLimit)
		log.Println("Truncated standard error.")
	}

	return &model.SimpleTestResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Start:  r.started,
		End:    r.stopped,
	}, nil
}

func (r *BestDockerRunner) inspect() *BestDockerRunner {
	if r.err != nil {
		return r
	}

	r.c, r.err = dc.InspectContainer(r.c.ID)
	if r.err != nil {
		log.Printf("Failed to inspect container %q: %s\n", r.c.ID, r.err)
	}
	return r
}

func (r *BestDockerRunner) remove() error {
	if r.err != nil {
		return r.err
	}

	r.err = dc.RemoveContainer(docker.RemoveContainerOptions{
		ID:            r.c.ID,
		RemoveVolumes: true,
		Force:         true,
	})

	return r.err
}
