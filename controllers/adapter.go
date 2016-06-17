package controllers

import (
	"io"
	"net/http"
	"time"

	"golang.org/x/net/context"

	s "github.com/coduno/runtime/storage/google"
)

type Adapter func(hw *wrapper)

func Method(method string) Adapter {
	return func(hw *wrapper) {
		oldWrapper := hw.h
		h := func(rd requestData, w http.ResponseWriter, r *http.Request) {
			if r.Method != method {
				http.Error(w, r.Method+" not allowed", http.StatusMethodNotAllowed)
				return
			}
			oldWrapper(rd, w, r)
		}
		hw.h = h
	}
}

type submittedFile struct {
	name string
	io.ReadCloser
}

func parseForm(r *http.Request) ([]submittedFile, error) {
	r.ParseMultipartForm(16 << 20) // That's 16MiB.
	files := r.MultipartForm.File["files"]

	if files == nil {
		fileNames := r.MultipartForm.Value["files"]
		return resolveStorage(fileNames)
	}

	result := make([]submittedFile, len(files))
	for i, f := range files {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		result[i] = submittedFile{name: f.Filename, ReadCloser: rc}
	}
	return result, nil
}

func resolveStorage(files []string) ([]submittedFile, error) {
	p := s.NewProvider()

	result := make([]submittedFile, len(files))
	for i, f := range files {
		o, err := p.Open(context.TODO(), f)
		if err != nil {
			return nil, err
		}
		result[i] = submittedFile{name: f, ReadCloser: o}
	}
	return result, nil
}

// TODO(victorbalan): Refactor this, split into multiple functions.
func Files(tar bool) Adapter {
	return func(hw *wrapper) {
		oldWrapper := hw.h
		h := func(rd requestData, w http.ResponseWriter, r *http.Request) {
			fs, err := parseForm(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			if tar {
				ball, err := tarobjects(fs)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				rd.ball = ball
			} else {
				if len(fs) < 1 {
					http.Error(w, "expected a file", http.StatusBadRequest)
					return
				}
				rd.files = fs
			}

			oldWrapper(rd, w, r)
		}
		hw.h = h
	}
}

func Language(languages []string) Adapter {
	return func(hw *wrapper) {
		oldWrapper := hw.h
		h := func(rd requestData, w http.ResponseWriter, r *http.Request) {
			lang := r.FormValue("language")
			if lang == "" {
				http.Error(w, "language param required", http.StatusBadRequest)
				return
			}
			if !contains(lang, languages) {
				http.Error(w, "language not supported", http.StatusBadRequest)
				return
			}

			rd.language = lang
			oldWrapper(rd, w, r)
		}
		hw.h = h
	}
}

func Test() Adapter {
	return func(hw *wrapper) {
		oldWrapper := hw.h
		h := func(rd requestData, w http.ResponseWriter, r *http.Request) {
			testPath := r.FormValue("test")
			if testPath == "" {
				http.Error(w, "test path missing.", http.StatusBadRequest)
				return
			}

			p := s.NewProvider()
			o, err := p.Create(context.TODO(), testPath, time.Hour, "text/plain")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			rd.test = o
			oldWrapper(rd, w, r)
		}
		hw.h = h
	}
}

func Stdin() Adapter {
	return func(hw *wrapper) {
		oldWrapper := hw.h
		h := func(rd requestData, w http.ResponseWriter, r *http.Request) {
			stdinPath := r.FormValue("stdin")
			if stdinPath == "" {
				http.Error(w, "stdin path missing", http.StatusBadRequest)
				return
			}
			p := s.NewProvider()
			stdin, err := p.Create(context.TODO(), stdinPath, time.Hour, "text/plain")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			rd.stdin = stdin
			oldWrapper(rd, w, r)
		}
		hw.h = h
	}
}

func Adapt(h *wrapper, adapters ...Adapter) wrapper {
	for _, adapter := range adapters {
		adapter(h)
	}
	return *h
}
