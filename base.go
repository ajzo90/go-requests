package requests

import (
	"fmt"
	"net/http"

	"github.com/valyala/fastjson"
)

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

func New(url interface{}) *Request {
	var c = &Request{header: map[string]stringer{}, query: map[string]stringer{}, doer: http.DefaultClient}
	return c.Url(url)
}

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

func (req *Request) Header(key string, value interface{}) *Request {
	req.header[key] = req.toStringer(value)
	return req
}

func (req *Request) Body(contentType string, value interface{}) *Request {
	req.body = req.toStringer(value)
	return req.ContentType(contentType)
}

func (req *Request) Query(key string, value interface{}) *Request {
	req.query[key] = req.toStringer(value)
	return req
}

func (req *Request) Method(method interface{}) *Request {
	req.method = req.toStringer(method)
	return req
}

func (req *Request) Url(url interface{}) *Request {
	req.url = req.toStringer(url)
	return req
}
