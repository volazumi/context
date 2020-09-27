package context

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"
)

var timeoutSync = sync.Mutex{}

type testResponse struct {
	Body    string
	Headers http.Header
	Status  int
	writen  chan struct{}
}

func (tr *testResponse) Header() http.Header {
	return tr.Headers
}

func (tr *testResponse) Write(msg []byte) (int, error) {
	tr.Body = string(msg)
	close(tr.writen)
	return len(msg), nil
}

func (tr *testResponse) WriteHeader(sc int) {
	tr.Status = sc
}

func (tr *testResponse) Reset() {
	tr.Body = ""
	tr.Status = 0
	tr.Headers = make(map[string][]string)
	tr.writen = make(chan struct{})
}

func newRW() *testResponse {
	return &testResponse{
		Headers: make(map[string][]string),
		writen:  make(chan struct{}),
	}
}

func TestNewContext(t *testing.T) {
	defaultTimeout = 1
	rw := newRW()

	// Test for GET request context (no body)
	req, _ := http.NewRequest("GET", "http://test.go/", nil)
	_, e := newContext(http.ResponseWriter(rw), req)
	if e != nil {
		t.Errorf("Error on create context: %s", e.Error())
	}

	select {
	case <-time.After((time.Duration(defaultTimeout) + 2) * time.Second):
		t.Errorf("Did not get answer after timeout")
	case <-rw.writen:
		if rw.Status != 537 {
			t.Errorf("Timeout answer expected")
		}
	}

	// Test for body request context
	body := "{\"test\": true}"
	req, _ = http.NewRequest("POST", "http://test.go/", bytes.NewBuffer([]byte(body)))
	_, e = newContext(http.ResponseWriter(rw), req)
	if e != nil {
		t.Errorf("Error on create context: %s", e.Error())
	}

	select {
	case <-time.After((time.Duration(defaultTimeout) + 2) * time.Second):
		t.Errorf("Did not get answer after timeout")
	case <-rw.writen:
		if rw.Status != 537 {
			t.Errorf("Timeout answer expected")
		}
	}
}

func TestServJSON(t *testing.T) {
	defaultTimeout = 300
	rw := newRW()
	req, _ := http.NewRequest("GET", "http://test.go/", nil)
	ctx, _ := newContext(http.ResponseWriter(rw), req)

	type answer struct {
		code int
		body interface{}
	}
	var inputs = []answer{
		answer{code: 200, body: map[string]bool{"success": true}},
	}
	var outputs = []answer{
		answer{code: 200, body: "{\"success\":true}"},
	}

	for i := 0; i < len(inputs); i++ {
		ctx.Status = inputs[i].code
		ctx.Response = inputs[i].code
		ctx.serveJSON()
		if rw.Status != outputs[i].code {
			t.Errorf(
				"Invalid status code provided. Expected %d, received %d",
				outputs[i].code,
				rw.Status,
			)
		}
		jsn, _ := json.Marshal(inputs[i].body)
		if string(jsn) != outputs[i].body {
			t.Errorf(
				"Invalid response provided. Expected %s, received %s",
				outputs[i].body,
				string(jsn),
			)
		}
		// rw.Status = outputs[i].code
	}
}
