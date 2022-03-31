package requests

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/valyala/fastjson"
)

// Request hold the request builder data
type Request struct {
	method  stringer
	baseUrl stringer
	path    stringer
	body    stringer
	header  stringerMap
	query   stringerMap
	secrets stringerMap

	respBodyBuf    []byte
	respBodyParser fastjson.Parser
	timeout        time.Duration

	err  error
	doer Doer // this doer should do all error handling, if it returns err=nil we are ready to use the payload
}

type Doer interface {
	// Do attempt to do one http request (and retries/redirects)
	Do(r *http.Request) (*http.Response, error)
}

func sleepUntil(ctx context.Context, until time.Time) error {
	d := -time.Since(until)
	if d < 0 {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

type stringerMap map[string]stringer

func (m stringerMap) CopyTo(newMap stringerMap) {
	for k, v := range m {
		newMap[k] = v
	}
}

type defaultDoer struct {
	doer Doer
}

func (d defaultDoer) Do(r *http.Request) (*http.Response, error) {
	resp, err := d.doer.Do(r)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		_ = drain(resp.Body)
		return nil, fmt.Errorf("invalid status %s", resp.Status)
	}
	return resp, nil
}

// New creates a new *Request
func New(url interface{}) *Request {
	c := &Request{header: map[string]stringer{}, query: map[string]stringer{}, doer: &defaultDoer{doer: http.DefaultClient}, secrets: map[string]stringer{}}
	return c.Url(url).Path("")
}

const (
	applicationJSON = "application/json"
)

func (req *Request) toStringer(v interface{}) stringer {
	return req._toStringer(v, false)
}

func (req *Request) _toStringer(v interface{}, secret bool) stringer {
	s := toStringer(v)
	if s == nil {
		req.setErr(fmt.Errorf("can not convert %v to stringer", v))
	}

	if secret {
		key := fmt.Sprintf("MASKED_%d", len(req.secrets)+1)
		req.Secret(key, s)
		return toStringer(SecretKey(key))
	} else {
		return s
	}
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

// SecretHeader sets a http header
func (req *Request) SecretHeader(key string, value interface{}) *Request {
	req.header[key] = req._toStringer(value, true)
	return req
}

func SecretKey(key string) string {
	return "${" + key + "}"
}

func (req *Request) Secret(key string, value interface{}) *Request {
	req.secrets[SecretKey(key)] = req.toStringer(value)
	return req
}

// Path sets a path
func (req *Request) Path(value interface{}) *Request {
	req.path = req.toStringer(value)
	return req
}

// Timeout sets the timeout
func (req *Request) Timeout(d time.Duration) *Request {
	req.timeout = d
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
	req.baseUrl = req.toStringer(url)
	return req
}
