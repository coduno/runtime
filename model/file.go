package model

import (
	"io"
	"mime/multipart"
	"os"
	"strings"
)

type CodeFile struct {
	Filename,
	Path string
	Size int64

	Bytes []byte
}

func NewCodeFile(fh *multipart.FileHeader) (CodeFile, error) {
	slashIndex := strings.LastIndex(fh.Filename, "/")
	var filename, path string
	if slashIndex > 0 {
		filename = fh.Filename[slashIndex:]
		path = fh.Filename[:slashIndex]
	} else {
		filename = fh.Filename
		path = "/"
	}

	f, err := fh.Open()
	if err != nil {
		return CodeFile{}, err
	}
	size := size(f)
	b := make([]byte, size)
	f.Read(b)
	file := CodeFile{
		Filename: filename,
		Path:     path,
		Size:     size,
		Bytes:    b,
	}

	return file, nil
}

func size(s io.Seeker) int64 {
	var size int64
	var err error
	if size, err = s.Seek(0, os.SEEK_END); err != nil {
		return -1
	}
	if _, err = s.Seek(0, os.SEEK_SET); err != nil {
		return -1
	}
	return size
}
