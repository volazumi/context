package context

import (
	"fmt"
	"strconv"

	"github.com/gorilla/mux"
)

type multiParam struct {
	ctx *Context
	e   error
}

func (ctx *Context) getParam(name string) (retval string, ok bool) {
	var value interface{}
	vars := mux.Vars(ctx.Request)
	value, ok = vars[name]

	if ok {
		retval, ok = value.(string)
		return
	}

	vars2 := ctx.Request.URL.Query()
	value, ok = vars2[name]
	if ok {
		retval = string(value.([]string)[0])
		return
	}

	return
}

func (t *multiParam) Int(name string) int {
	if t.e != nil {
		return 0
	}
	p, ok := t.ctx.getParam(name)
	if !ok {
		t.e = fmt.Errorf("Param %s is not ok", name)
		return 0
	}
	i, e := strconv.Atoi(p)
	t.e = e
	return i
}

func (t *multiParam) String(name string) string {
	if t.e != nil {
		return ""
	}
	p, ok := t.ctx.getParam(name)
	if !ok {
		t.e = fmt.Errorf("Param %s is not ok", name)
	}
	return p
}

func (t *multiParam) Error() error {
	return t.e
}
