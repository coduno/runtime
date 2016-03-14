package controllers

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"

	"github.com/coduno/runtime-dummy/storage"
)

// TODO send correct filename from app
var fileNames = map[string]string{"py": "app.py", "c": "app.c", "cpp": "app.cpp",
	"java": "Application.java", "csharp": "application.cs", "js": "app.js", "php": "app.php", "go": "app.go",
	"groovy": "app.groovy", "scala": "Application.scala"}

func contains(s string, strings []string) bool {
	for _, str := range strings {
		if s == str {
			return true
		}
	}
	return false
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

func gcsmaketar(o storage.Object, fileName string) (ball io.Reader, err error) {
	buf := new(bytes.Buffer)
	tarw := tar.NewWriter(buf)

	b, err := ioutil.ReadAll(o)
	if err != nil {
		o.Close()
		return nil, err
	}
	if err = tarw.WriteHeader(&tar.Header{
		Name: fileName,
		Mode: 0600,
		Size: int64(len(b)),
	}); err != nil {
		o.Close()
	}
	io.WriteString(tarw, string(b))
	o.Close()

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
