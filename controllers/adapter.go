package controllers

import (
	"archive/tar"
	"bytes"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/coduno/runtime/model"
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

func MakeTar() Adapter {
	return func(hw *wrapper) {
		oldWrapper := hw.h
		h := func(rd requestData, w http.ResponseWriter, r *http.Request) {
			buf := new(bytes.Buffer)
			tarw := tar.NewWriter(buf)

			for _, fh := range rd.files {
				if err := tarw.WriteHeader(&tar.Header{
					Name: fh.Filename,
					Mode: 0600,
					Size: fh.Size,
				}); err != nil {
					http.Error(w, "maketar error"+err.Error(), http.StatusInternalServerError)
				}
				tarw.Write(fh.Bytes)
			}
			rd.ball = bytes.NewReader(buf.Bytes())
			oldWrapper(rd, w, r)
		}
		hw.h = h
	}
}

func CodeFiles() Adapter {
	return func(hw *wrapper) {
		oldWrapper := hw.h
		h := func(rd requestData, w http.ResponseWriter, r *http.Request) {
			submissionPath := r.FormValue("files_gcs")
			if submissionPath != "" {
				// Load gcs files
			} else {
				files, found := r.MultipartForm.File["files"]
				if !found {
					http.Error(w, "missing files", http.StatusBadRequest)
					return
				}
				rd.files = make([]model.CodeFile, len(files))
				for i, file := range files {
					codeFile, err := model.NewCodeFile(file)
					if err != nil {
						http.Error(w, "file extraction error: "+err.Error(), http.StatusInternalServerError)
						return
					}
					rd.files[i] = codeFile
				}
			}
			oldWrapper(rd, w, r)
		}
		hw.h = h
	}
}

// TODO(victorbalan): refactor this, split into multiple functions, adapt for
// multiple files
func Files(tar bool) Adapter {
	return func(hw *wrapper) {
		oldWrapper := hw.h
		h := func(rd requestData, w http.ResponseWriter, r *http.Request) {
			submissionPath := r.FormValue("files")
			if submissionPath != "" {
				p := s.NewProvider()
				o, err := p.Create(context.TODO(), submissionPath, time.Hour, "text/plain")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				rd.ball = o
				if tar {
					ball, err := gcsmaketar(o, fileNames[r.FormValue("language")])
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					rd.ball = ball
				}
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
				if tar {
					ball, err := maketar(files[0], fileNames[rd.language])
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					rd.ball = ball
				} else {
					file, err := getReader(files[0])
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					rd.ball = file
				}
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
