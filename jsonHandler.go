// +build !general

package context

import (
	"fmt"
	"net/http"
	"time"
)

//Handler will store midds and controller
type Handler struct {
	ctx         *Context
	ContentType func() string
	Funcs       []func(*Context) error
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		e := recover()
		if e != nil {
			checkErr(fmt.Errorf("jsonHandler recovered from: %+v", e))
		}
	}()

	ctx, e := newContext(w, r)
	if checkErr(e) {
		ctx.Status = 500
		ctx.Response = e.Error()
		ctx.serveError()
		return
	}

	h.ctx = ctx
	for i := 0; i < len(h.Funcs) && e == nil; i++ {
		e = h.Funcs[i](ctx)
	}

	if e != nil {
		ctx.Response = e.Error()
	}

	w.Header().Set("Last-Modified", time.Now().Format("Mon Jan 2 15:04:05 MST 2006"))
	w.Header().Set("Content-Type", h.ContentType())
	writeCrossDomainHeaders(w, r)
	ctx.serveJSON()
}

// HandleJSON is for public routes
func HandleJSON(mids []string, ctrl ...func(*Context) error) Handler {
	var h = Handler{
		ContentType: func() string { return "application/json;charset=UTF-8" },
		Funcs:       make([]func(*Context) error, len(mids)+len(ctrl)),
	}
	for i := 0; i < len(mids); i++ {
		fn, ok := middlewares[mids[i]]
		if !ok {
			panic(fmt.Sprintf("%s is not into registered middlewares", mids[i]))
		}
		h.Funcs[i] = fn
	}
	for i := 0; i < len(ctrl); i++ {
		h.Funcs[len(mids)+i] = ctrl[i]
	}

	return h
}
