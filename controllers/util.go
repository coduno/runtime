package controllers

import (
	"archive/tar"
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
)

// TODO send correct filename from app
var fileNames = map[string]string{
	"py":     "app.py",
	"c":      "app.c",
	"cpp":    "app.cpp",
	"java":   "Application.java",
	"csharp": "application.cs",
	"js":     "app.js",
	"php":    "app.php",
	"go":     "app.go",
	"groovy": "app.groovy",
	"scala":  "Application.scala",
	"pascal": "app.pas",
}

func contains(s string, strings []string) bool {
	for _, str := range strings {
		if s == str {
			return true
		}
	}
	return false
}

func tarobjects(fs []submittedFile) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tarw := tar.NewWriter(buf)

	for _, f := range fs {
		b, err := ioutil.ReadAll(f)
		if err != nil {
			f.Close()
			return nil, err
		}
		if err = tarw.WriteHeader(&tar.Header{
			Name: f.name,
			Mode: 0660,
			Size: int64(len(b)),
		}); err != nil {
			f.Close()
		}

		tarw.Write(b)
		f.Close()
	}

	return bytes.NewReader(buf.Bytes()), nil
}

func getFile(filename string) (io.Reader, error) {
	b, err := ioutil.ReadFile("testfiles/" + filename)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func getReader(f *multipart.FileHeader) (io.Reader, error) {
	file, err := f.Open()
	if err != nil {
		return nil, err
	}
	return file, nil
}
