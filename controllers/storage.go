package controllers

import (
	"fmt"
	"net/http"

	storage "github.com/coduno/runtime-dummy/storage/google"
	"google.golang.org/appengine"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"time"
	"bytes"
)

var cloudClient *http.Client

func init() {
	router.Handle("/gcs", Adapt(Wrap(upload), Method("GET")))
	var err error
	cloudClient, err = google.DefaultClient(context.Background())
	if err != nil {
		panic(err)
	}
}
const projID = "coduno"

func CloudContext(parent context.Context) context.Context {
	if parent == nil {
		return cloud.NewContext(projID, cloudClient)
	}
	return cloud.WithContext(parent, projID, cloudClient)
}

func upload(rd requestData, w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r);

	p := storage.NewProvider();
	o, err := p.Create(ctx, "coduno/testfile.txt", time.Hour, "text/plain");
	if (err != nil){
		fmt.Printf("Error", err);
	}
	b := new([]byte);
	if nr, err := o.Read(*b); err == nil {
		fmt.Println("Was an error ", err);
	}else {
		fmt.Println("**************************************************************")
		fmt.Println("All good, Nr read = ", nr);
		buf := bytes.NewBuffer(*b);
		buf.String()
		fmt.Println("Smth", buf.String());
	}
}
