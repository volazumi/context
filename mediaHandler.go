package context

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"bitbucket.org/sgatenea/sitigo.backend/clog"
)

//MediaHandler will store midds and controller
type MediaHandler struct {
	ctx         *Context
	ContentType string
	Mids        []func(*Context) error
	Ctrl        func(*Context) *os.File
}

//HandleFile is for public routes
func HandleFile(contentType string, mids []string, ctrl func(*Context) *os.File) MediaHandler {
	var h = MediaHandler{
		ContentType: contentType,
		Mids:        make([]func(*Context) error, len(mids)),
	}
	for i := 0; i < len(mids); i++ {
		fn, ok := middlewares[mids[i]]
		if !ok {
			panic(fmt.Sprintf("%s is not into registered middlewares", mids[i]))
		}
		h.Mids[i] = fn
	}
	h.Ctrl = ctrl
	return h
}

func (h MediaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, e := newContext(w, r, false)
	if clog.CheckError(e) {
		ctx.Status = 500
		ctx.Response = e.Error()
		ctx.serveError()
		return
	}

	h.ctx = ctx
	for i := 0; i < len(h.Mids) && e == nil; i++ {
		e = h.Mids[i](ctx)
	}

	if e != nil {
		ctx.Response = e.Error()
		ctx.serveError()
		return
	}

	file := h.Ctrl(ctx)
	if file == nil {
		ctx.Status = 404
		ctx.Response = "File not found"
		return
	}

	defer func() {
		e := recover()
		if e != nil {
			clog.CheckError(fmt.Errorf("recovered from: %s", e))
		}
		file.Close()
	}()

	if ctx.Status == 0 {
		ctx.Status = 200
	}

	var lastModified string
	if file != nil {
		stat, _ := file.Stat()
		lastModified = stat.ModTime().String()
	}

	var size string
	s, e := file.Stat()
	if e == nil {
		size = fmt.Sprintf("%d", s.Size())
	} else {
		size = "0"
	}

	var buf = make([]byte, s.Size())
	file.Read(buf)
	file.Seek(0, 0)

	filetype := h.ContentType
	if filetype == "" {
		filetype = http.DetectContentType(buf)
	}

	writeCrossDomainHeaders(w, r)
	w.Header().Set("Content-Type", filetype)
	w.Header().Set("Content-Length", size)
	w.Header().Set("Last-Modified", lastModified)

	io.Copy(w, file)

	ctx.w.WriteHeader(ctx.Status)
	io.Copy(w, file)

}
