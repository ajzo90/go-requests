package requests

import (
	"fmt"
	"net/http"

	"github.com/valyala/fastjson"
)

// Request hold the request builder data
type Request struct {
	method stringer
	url    stringer
	header map[string]stringer
	query  map[string]stringer
	body   stringer

	respBodyBuf    []byte
	respBodyParser fastjson.Parser

	err  error
	doer Doer
}

// New creates a new *Request
func New(url interface{}) *Request {
	c := &Request{header: map[string]stringer{}, query: map[string]stringer{}, doer: http.DefaultClient}
	return c.Url(url)
}

// Doer is the type that send the request, defaults to a http.DefaultClient
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

const (
	applicationJSON = "application/json"
)

func (req *Request) toStringer(v interface{}) stringer {
	s := toStringer(v)
	if s == nil {
		req.setErr(fmt.Errorf("can not convert %v to stringer", v))
	}
	return s
}

func (req *Request) setErr(err error) {
	if req.err == nil {
		req.err = err
	}
}

// Header sets a http header
func (req *Request) Header(key string, value interface{}) *Request {
	req.header[key] = req.toStringer(value)
	return req
}

// Body set the http body
func (req *Request) Body(contentType string, value interface{}) *Request {
	req.body = req.toStringer(value)
	return req.ContentType(contentType)
}

// Query sets a http query
func (req *Request) Query(key string, value interface{}) *Request {
	req.query[key] = req.toStringer(value)
	return req
}

// Method sets the http method
func (req *Request) Method(method interface{}) *Request {
	req.method = req.toStringer(method)
	return req
}

// Url sets the http url
func (req *Request) Url(url interface{}) *Request {
	req.url = req.toStringer(url)
	return req
}
