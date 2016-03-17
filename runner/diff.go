package runner

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/coduno/runtime-dummy/model"
)

func DiffRun(ball, test io.Reader, image string) (ts model.TestStats, err error) {
	var str model.SimpleTestResult
	str, err = SimpleRun(ball, image)
	if err != nil {
		return
	}
	tr := model.DiffTestResult{
		SimpleTestResult: str,
	}

	ts, err = processDiffResults(&tr, test)
	return
}

func processDiffResults(tr *model.DiffTestResult, want io.Reader) (ts model.TestStats, err error) {
	have := strings.NewReader(tr.Stdout)
	diffLines, ok, err := compare(want, have)
	if err != nil {
		return
	}
	tr.DiffLines = diffLines
	tr.Failed = !ok

	ts = model.TestStats{
		Stdout: tr.Stdout,
		Stderr: tr.Stderr,
		Failed: !ok,
	}

	// _, err = tr.PutWithParent(ctx, sub.Key)
	return
}

func compare(want, have io.Reader) ([]int, bool, error) {
	w, err := ioutil.ReadAll(want)
	if err != nil {
		return nil, false, err
	}
	h, err := ioutil.ReadAll(have)
	if err != nil {
		return nil, false, err
	}
	w = bytes.Replace(w, []byte("\r\n"), []byte("\n"), -1)
	h = bytes.Replace(h, []byte("\r\n"), []byte("\n"), -1)
	wb := bytes.Split(w, []byte("\n"))
	hb := bytes.Split(h, []byte("\n"))

	if len(wb) != len(hb) {
		return nil, false, nil
	}

	var diff []int
	ok := true
	for i := 0; i < len(wb); i++ {
		// fmt.Println(string(wb[i]), string(hb[i]))
		if bytes.Compare(wb[i], hb[i]) != 0 {
			diff = append(diff, i)
			ok = false
		}
	}
	fmt.Println(diff)

	return diff, ok, nil
}
