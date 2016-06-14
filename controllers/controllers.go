package controllers

import (
	"io"
	"net/http"

	"github.com/coduno/runtime/model"
	"github.com/gorilla/mux"
)

type requestParams struct {
	method                       string
	files, language, test, stdin bool
}

type requestData struct {
	ball        io.Reader
	files       []model.CodeFile
	language    string
	test, stdin io.Reader
}

var router = mux.NewRouter()

func Router() *mux.Router {
	return router
}

type wrapper struct {
	h  handlerFuncWrapper
	rd requestData
}

func (hw wrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hw.h(hw.rd, w, r)
}

func Wrap(h handlerFuncWrapper) *wrapper {
	return &wrapper{h: h}
}

type handlerFuncWrapper func(requestData, http.ResponseWriter, *http.Request)
