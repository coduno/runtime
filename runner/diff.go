package runner

import (
	"io"
	"strings"
	"log"
	"bufio"

	"github.com/coduno/runtime/model"
)

func DiffRun(ball, test io.Reader, image string) (*model.DiffTestResult, error) {
	str, err := SimpleRun(ball, image)
	if err != nil {
		return nil, err
	}
	return processDiffResults(&model.DiffTestResult{SimpleTestResult: *str}, test)
}

func processDiffResults(tr *model.DiffTestResult, want io.Reader) (*model.DiffTestResult, error) {
	have := strings.NewReader(tr.Stdout)
	log.Println("[runner] [diff.go] processDiffResults compare")
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

		log.Println("[runner] [diff.go] processDiffResults W", w)
		log.Println("[runner] [diff.go] processDiffResults H", h)

		l := len(w)
		if len(h) < l {
			l = len(h)
		}

		for offset := 1; offset <= l; offset++ {
			if w[offset - 1] != h[offset - 1] {
				return &model.Mismatch{line, offset}, nil
			}
		}

		if len(h) != len(w) {
			log.Println("[runner] [diff.go] processDiffResults length of the line differs (want:", len(w), "have:", len(h), ")")
			return &model.Mismatch{line, l}, nil
		}
	}

	return nil, nil
}