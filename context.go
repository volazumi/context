package context

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const defaultTimeout = 300

var (
	hostname   string
	enableCors bool
)

// Context is used to handle Req
type Context struct {
	multiParam
	Body       []byte
	doneChan   chan struct{}
	Request    *http.Request
	w          http.ResponseWriter
	start      time.Time
	Status     int
	Response   interface{}
	CustomData map[string]interface{}
}

func init() {
	hostname, _ = os.Hostname()
}

//SetCors mark true if needed to send CORS
func SetCors(enabled bool) {
	enableCors = enabled
}

func newContext(w http.ResponseWriter, r *http.Request, needAuth bool) (ctx *Context, e error) {
	var body []byte
	ctx = &Context{}

	ctx.doneChan = make(chan struct{}, 1)
	defer close(ctx.doneChan)
	go func() {
		select {
		case <-ctx.doneChan:
			return
		case <-time.After(defaultTimeout * time.Second):
			ctx.Status = 537
			ctx.Response = "Imposible to answer. Server is hanged up or too busy"
			ctx.serveError()
		}
	}()
	ctx.Request = r

	ctx.w = w
	body, _ = ioutil.ReadAll(r.Body)
	ctx.Body = body

	e = r.Body.Close()
	if checkErr(r.Body.Close()) {
		return
	}

	r.ParseForm()
	ctx.multiParam = multiParam{ctx: ctx}
	ctx.CustomData = make(map[string]interface{})

	return
}

func (ctx *Context) serveJSON() {
	ctx.w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	switch {
	case ctx.Status == 0:
		ctx.Status = 200
	case ctx.Status > 399:
		ctx.serveError()
		return
	}

	var jsonResponse []byte
	if ctx.Response == nil {
		ctx.Response = map[string]string{"status": "ok"}
	}
	jsonResponse, _ = json.Marshal(ctx.Response)
	var myJS = string(jsonResponse)
	if myJS == "null" {
		myJS = "[]"
	}

	ctx.w.WriteHeader(ctx.Status)
	io.WriteString(ctx.w, myJS)
}

func (ctx *Context) serveError() {
	ctx.w.Header().Set("Content-Type", "application/json")
	ctx.w.WriteHeader(ctx.Status)
	jout, _ := json.Marshal(map[string]interface{}{"error": ctx.Response})
	io.WriteString(ctx.w, string(jout))
}

func writeCrossDomainHeaders(w http.ResponseWriter, req *http.Request) {
	// Cross domain headers
	if !enableCors {
		return
	}
	if acrh, ok := req.Header["Access-Control-Request-Headers"]; ok {
		w.Header().Set("Access-Control-Allow-Headers", acrh[0])
	} else {
		w.Header().Set("Access-Control-Allow-Headers", "*")
	}
	w.Header().Set("Access-Control-Allow-Credentials", "True")
	if acao, ok := req.Header["Access-Control-Allow-Origin"]; ok {
		w.Header().Set("Access-Control-Allow-Origin", acao[0])
	} else {
		if _, oko := req.Header["Origin"]; oko {
			w.Header().Set("Access-Control-Allow-Origin", req.Header["Origin"][0])
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH")
	w.Header().Set("Connection", "Close")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("Access-Control-Expose-Headers", "Auth-Token")

}
