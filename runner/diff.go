package runner

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"

	"github.com/coduno/runtime/model"
	"github.com/fsouza/go-dockerclient"
)

var logger = log.New(os.Stderr, "", log.LstdFlags | log.LUTC | log.Lshortfile)

func DiffRun(ball, test io.Reader, image string) (*model.DiffTestResult, error) {
	runner := &Runner{
		Config: &docker.Config{
			Image: image,
		},
	}

	str, err := runner.
		CreateContainer().
		Upload(ball).
		Start().
		Wait().
		Logs()

	if err != nil {
		return nil, err
	}

	return processDiffResults(&model.DiffTestResult{SimpleTestResult: *str}, test)
}

func processDiffResults(tr *model.DiffTestResult, want io.Reader) (*model.DiffTestResult, error) {
	have := strings.NewReader(tr.Stdout)
	mismatch, err := compare(want, have)
	if err != nil {
		return nil, err
	}

	tr.Successful = mismatch == nil
	if !tr.Successful {
		tr.Mismatch = *mismatch
	}

	return tr, nil
}

func compare(want, have io.Reader) (*model.Mismatch, error) {
	sch := bufio.NewScanner(have)
	sch.Split(bufio.ScanLines)
	scw := bufio.NewScanner(want)
	scw.Split(bufio.ScanLines)

	for line := 1; ; line++ {
		sh := sch.Scan()
		sw := scw.Scan()
		if !sh && !sw {
			if sch.Err() == nil && scw.Err() == nil {
				return nil, nil
			}
		}
		if sch.Err() != nil {
			return nil, sch.Err()
		}
		if scw.Err() != nil {
			return nil, scw.Err()
		}

		h := sch.Text()
		w := scw.Text()

		l := len(w)
		if len(h) < l {
			l = len(h)
		}

		for offset := 1; offset <= l; offset++ {
			if w[offset - 1] != h[offset - 1] {
				logger.Printf("Mismatch %q (have) against %q (want) at %d:%d\n", h, w, line, offset)
				return &model.Mismatch{line, offset}, nil
			}
		}

		if len(h) != len(w) {
			logger.Printf("Length mismatch %d (have) against %d (want) at line %d\n", len(h), len(w), line)
			return &model.Mismatch{line, l}, nil
		}
	}

	return nil, nil
}
