package controllers

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"os"

	"github.com/gorilla/mux"
)

var router = mux.NewRouter()
var fileNames = map[string]string{"py": "app.py", "c": "app.c", "cpp": "app.cpp"}

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
