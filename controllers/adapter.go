package controllers

import (
	"net/http"
	"time"

	s "github.com/coduno/runtime-dummy/storage/google"
	"google.golang.org/appengine"
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

func Files() Adapter {
	return func(hw *wrapper) {
		oldWrapper := hw.h
		h := func(rd requestData, w http.ResponseWriter, r *http.Request) {
			submissionPath := r.FormValue("files_gcs")
			if submissionPath != "" {
				ctx := appengine.NewContext(r)

				p := s.NewProvider()
				o, err := p.Create(ctx, s.SubmissionsBucket()+"/"+submissionPath, time.Hour, "text/plain")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				ball, err := gcsmaketar(o, fileNames[rd.language])
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				rd.ball = ball
			} else {
				if err := r.ParseMultipartForm(16 << 20); err != nil {
					http.Error(w, "could not parse multipart form: "+err.Error(), http.StatusBadRequest)
					return
				}

				files, found := r.MultipartForm.File["files"]
				if !found {
					http.Error(w, "missing files", http.StatusBadRequest)
					return
				}
				if len(files) != 1 {
					http.Error(w, "we currently support only single file uploads", http.StatusBadRequest)
					return
				}
				ball, err := maketar(files[0], fileNames[rd.language])
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				rd.ball = ball
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
			ctx := appengine.NewContext(r)

			p := s.NewProvider()
			o, err := p.Create(ctx, s.TestsBucket()+"/"+testPath, time.Hour, "text/plain")
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
			stdin, err := getFile(stdinPath)
			if err != nil {
				http.Error(w, "get stdin file error: "+err.Error(), http.StatusInternalServerError)
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
