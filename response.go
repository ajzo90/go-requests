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
	v *fastjson.Value
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

// ExecJSON executes the request and return a *JSONResponse
func (req *Request) ExecJSON(ctxs ...context.Context) (*JSONResponse, error) {
	resp, err := req.doJSON(ctxs...)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	req.respBodyBuf, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	jsonResp := &JSONResponse{response: response{raw: resp}}

	jsonResp.v, err = req.respBodyParser.ParseBytes(req.respBodyBuf)

	return jsonResp, err
}

func (r *response) Header(key string) string {
	return r.raw.Header.Get(key)
}

type response struct {
	raw *http.Response
}
