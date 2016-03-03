package controllers

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type requestParams struct {
	method                       string
	files, language, test, stdin bool
}

type requestData struct {
	files       []*multipart.FileHeader
	language    string
	test, stdin io.Reader
}

func processRequest(w http.ResponseWriter, r *http.Request, rp requestParams) (rd requestData, ok bool) {
	ok = false
	if r.Method != rp.method {
		http.Error(w, r.Method+" not allowed", http.StatusMethodNotAllowed)
		return
	}
	if rp.files {
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
		rd.files = files
	}
	if rp.language {
		rd.language = r.FormValue("language")
	}

	if rp.test {
		testPath := r.FormValue("test")
		if testPath == "" {
			http.Error(w, "test path missing.", http.StatusBadRequest)
			return
		}
		test, err := getFile(testPath)
		if err != nil {
			http.Error(w, "get test file error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		rd.test = test
	}

	if rp.stdin {
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
	}

	return rd, true
}

var router = mux.NewRouter()

// TODO send correct filename from app
var fileNames = map[string]string{"py": "app.py", "c": "app.c", "cpp": "app.cpp", "java": "Application.java"}

func Router() *mux.Router {
	return router
}

func maketar(fh *multipart.FileHeader, fileName string) (ball io.Reader, err error) {
	sizeFunc := func(s io.Seeker) int64 {
		var size int64
		if size, err = s.Seek(0, os.SEEK_END); err != nil {
			return -1
		}
		if _, err = s.Seek(0, os.SEEK_SET); err != nil {
			return -1
		}
		return size
	}
	buf := new(bytes.Buffer)
	tarw := tar.NewWriter(buf)

	var f multipart.File

	if f, err = fh.Open(); err != nil {
		return
	}
	size := sizeFunc(f)
	if size < 0 {
		return nil, errors.New("seeker can't seek")
	}
	if err = tarw.WriteHeader(&tar.Header{
		Name: fileName,
		Mode: 0600,
		Size: size,
	}); err != nil {
		f.Close()
	}
	io.Copy(tarw, f)
	f.Close()

	ball = bytes.NewReader(buf.Bytes())
	return
}

func getFile(filename string) (io.Reader, error) {
	b, err := ioutil.ReadFile("testfiles/" + filename)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}
