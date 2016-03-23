package controllers

import (
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	storage "github.com/coduno/runtime/storage/google"
	"google.golang.org/appengine"
)

func init() {
	router.Handle("/gcs", Wrap(read))
}

func read(rd requestData, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	p := storage.NewProvider()
	o, err := p.Create(ctx, "coduno/testfile2.txt", time.Hour, "text/plain")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case "GET":
		b, err := ioutil.ReadAll(o)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(b)
	case "POST":
		n, err := io.WriteString(o, "test string")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		o.Close()
		w.Write([]byte("we have written " + strconv.Itoa(n) + " bytes"))

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
