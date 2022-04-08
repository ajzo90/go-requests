package requests

import (
	"context"
	"io"
	"net/http"

	"github.com/valyala/fastjson"
)

// JSONResponse is holding the JSON specific response
type JSONResponse struct {
	response
	v   *fastjson.Value
	p   JSONParser
	buf []byte
}

func (r *JSONResponse) SetParser(p JSONParser) {
	r.p = p
}

// String get string from JSON body
func (r *JSONResponse) String(keys ...string) string {
	return string(r.v.GetStringBytes(keys...))
}

// Int gets int from JSON body
func (r *JSONResponse) Int(keys ...string) int {
	return r.v.GetInt(keys...)
}

// GetArray gets array from JSON body
func (r *JSONResponse) GetArray(keys ...string) []*fastjson.Value {
	return r.v.GetArray(keys...)
}

// Body gets the JSON body
func (r *JSONResponse) Body() *fastjson.Value {
	return r.v
}

func (req *Request) doJSON(ctxs ...context.Context) (*http.Response, error) {
	return req.Header("accept", applicationJSON).Extended().Do(ctxs...)
}

// ExecJSONPreAlloc executes the request and fill jsonResp
func (req *ExtendedRequest) ExecJSONPreAlloc(jsonResp *JSONResponse, ctxs ...context.Context) error {
	resp, err := req.doJSON(ctxs...)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.ContentLength == 0 {
		// do nothing
	} else if resp.ContentLength > 0 {
		if cap(jsonResp.buf) >= int(resp.ContentLength) {
			jsonResp.buf = jsonResp.buf[:resp.ContentLength]
		} else {
			jsonResp.buf = make([]byte, resp.ContentLength)
		}
		_, err = io.ReadFull(resp.Body, jsonResp.buf)
	} else {
		jsonResp.buf, err = io.ReadAll(resp.Body)
	}

	if err != nil {
		return err
	}

	jsonResp.response.raw = resp
	if jsonResp.p == nil {
		jsonResp.p = &fastjson.Parser{}
	}
	jsonResp.v, err = jsonResp.p.ParseBytes(jsonResp.buf)

	return err
}

type JSONParser interface {
	ParseBytes([]byte) (*fastjson.Value, error)
}

// ExecJSON executes the request and return a *JSONResponse
func (req *Request) ExecJSON(ctxs ...context.Context) (*JSONResponse, error) {
	var r JSONResponse
	err := req.Extended().ExecJSONPreAlloc(&r, ctxs...)
	return &r, err
}

func (r *response) Header(key string) string {
	return r.raw.Header.Get(key)
}

type response struct {
	raw *http.Response
}
