package runner

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"time"

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
	dc, err = docker.NewClientFromEnv()
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

type Runner struct {
	c                *docker.Container
	Config           *docker.Config
	hostConfig       *docker.HostConfig
	started, stopped time.Time
	err              error
}

func (r *Runner) prepare() *Runner {
	if r.err != nil {
		return r
	}

	_, r.err = dc.InspectImage(r.Config.Image)
	if r.err == nil {
		return r
	}

	if r.err != docker.ErrNoSuchImage {
		log.Printf("Error inspecting image %q: %s\n", r.Config.Image, r.err)
		return r
	}

	r.err = dc.PullImage(docker.PullImageOptions{
		Repository:   r.Config.Image,
		OutputStream: os.Stderr,
	}, docker.AuthConfiguration{})

	if r.err != nil {
		log.Printf("Error pulling image %q: %s\n", r.Config.Image, r.err)
	}
	return r
}

func (r *Runner) CreateContainer() *Runner {
	if r.err != nil {
		return r
	}

	if r.Config.Env == nil {
		r.Config.Env = []string{codunoFlag}
	} else {
		r.Config.Env = append(r.Config.Env, codunoFlag)
	}

	if r.hostConfig == nil {
		r.hostConfig = &docker.HostConfig{
			Privileged:false,
		}
	} else {
		r.hostConfig.Privileged = false
	}

	if r.hostConfig.Memory > memoryLimit {
		r.hostConfig.Memory = memoryLimit
	}

	r.c, r.err = dc.CreateContainer(docker.CreateContainerOptions{
		Config:     r.Config,
		HostConfig: r.hostConfig,
	})

	if r.err != nil {
		log.Printf("Failed to create container from image %q: %s\n", r.Config.Image, r.err)
	}
	return r
}

func (r *Runner) Upload(ball io.Reader) *Runner {
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

func (r *Runner) Start() *Runner {
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

func (r *Runner) Attach(stream io.Reader) *Runner {
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

func (r *Runner) Wait() *Runner {
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

func (r *Runner) Logs() (*model.SimpleTestResult, error) {
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

func (r *Runner) inspect() *Runner {
	if r.err != nil {
		return r
	}

	r.c, r.err = dc.InspectContainer(r.c.ID)
	if r.err != nil {
		log.Printf("Failed to inspect container %q: %s\n", r.c.ID, r.err)
	}
	return r
}

func (r *Runner) remove() error {
	if r.err != nil {
		return r.err
	}

	// NOTE(flowlo): If removal of the container fails,
	// this does not currupt the instance.
	return dc.RemoveContainer(docker.RemoveContainerOptions{
		ID:            r.c.ID,
		RemoveVolumes: true,
		Force:         true,
	})
}

func (r *Runner) download(path string, w io.Writer) error {
	if r.err != nil {
		return r.err
	}

	return dc.DownloadFromContainer(r.c.ID, docker.DownloadFromContainerOptions{
		Path:         path,
		OutputStream: w,
	})
}
