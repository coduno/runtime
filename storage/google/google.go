package google

import (
	"errors"
	"io"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/coduno/runtime/storage"
	s "google.golang.org/cloud/storage"
)

func split(fullName string) (bucket string, name string, err error) {
	i := strings.Index(fullName, "/")

	if i == -1 {
		return "", "", errors.New("name invalid, must contain '/'")
	}

	return fullName[:i], fullName[i+1:], nil
}

type provider struct{}

func (p provider) Create(ctx context.Context, name string, maxAge time.Duration, contentType string) (storage.Object, error) {
	b, n, err := split(name)
	if err != nil {
		return nil, err
	}

	a := s.ObjectAttrs{
		ContentType:        contentType,
		ContentLanguage:    "", // TODO(flowlo): Does it make sense to set this?
		ContentEncoding:    "utf-8",
		CacheControl:       "max-age=" + strconv.FormatFloat(maxAge.Seconds(), 'f', 0, 64),
		ContentDisposition: "attachment; filename=\"" + path.Base(name) + "\"",
		Bucket:             b,
		Name:               n,
	}

	return &object{
		b:   b,
		n:   n,
		w:   nil,
		r:   nil,
		a:   a,
		ctx: ctx,
	}, nil
}

func (p provider) Open(ctx context.Context, name string) (storage.Object, error) {
	b, n, err := split(name)
	if err != nil {
		return nil, err
	}

	return &object{
		b: b,
		n: n,
		w: nil,
		r: nil,
		a: s.ObjectAttrs{
			Bucket: b,
			Name:   n,
		},
		ctx: ctx,
	}, nil
}

type object struct {
	b   string
	n   string
	w   *s.Writer
	r   io.ReadCloser
	a   s.ObjectAttrs
	ctx context.Context
}

func (o *object) Write(p []byte) (n int, err error) {
	c, err := s.NewClient(o.ctx)
	if err != nil {
		return 0, errors.New("cannot create client:" + err.Error())
	}
	defer c.Close()
	if o.r != nil {
		return 0, errors.New("object is already opened for reading")
	}
	if o.w == nil {
		o.w = c.Bucket(o.b).Object(o.n).NewWriter(o.ctx)
		o.w.ObjectAttrs = o.a
	}
	if o.w == nil {
		return 0, errors.New("failed to connect to gcs")
	}
	return o.w.Write(p)
}

func (o *object) Close() error {
	if o.w != nil && o.r != nil {
		return errors.New("object is opened for reading and writing")
	}
	if o.w != nil {
		return o.w.Close()
	}
	if o.r != nil {
		return (o.r).Close()
	}
	return errors.New("nothing to close")
}

func (o *object) Read(p []byte) (n int, err error) {
	c, err := s.NewClient(o.ctx)
	if err != nil {
		return 0, errors.New("cannot create client")
	}
	if o.w != nil {
		return 0, errors.New("object is already opened for writing")
	}
	if o.r == nil {
		rc, err := c.Bucket(o.b).Object(o.n).NewReader(o.ctx)
		if err != nil {
			return 0, err
		}
		o.r = ioutil.NopCloser(rc)
	}

	return (o.r).Read(p)
}

func (o *object) Name() string {
	return o.b + o.n
}

func NewProvider() storage.Provider {
	return provider{}
}
