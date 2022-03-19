package requests

import (
	"fmt"
	"io"
	"net/http"

	"github.com/valyala/fastjson"
)

// JSONResponse is holding the JSON specific response
type JSONResponse struct {
	response
	v *fastjson.Value
}

// Get string from JSON body
func (r *JSONResponse) Get(keys ...string) string {
	return string(r.v.GetStringBytes(keys...))
}

// GetInt gets int from JSON body
func (r *JSONResponse) GetInt(keys ...string) int {
	return r.v.GetInt(keys...)
}

// GetArray gets array from JSON body
func (r *JSONResponse) GetArray(keys ...string) []*fastjson.Value {
	return r.v.GetArray(keys...)
}

// ExecJSON executes the request and return a *JSONResponse
func (req *Request) ExecJSON() (*JSONResponse, error) {
	req.Header("accept", applicationJSON)

	resp, err := req.Extended().Do()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if req.respBodyBuf, err = readToBuf(make([]byte, 4096), resp.Body); err != nil {
		return nil, err
	}

	var jsonResp = &JSONResponse{response: response{raw: resp}}

	jsonResp.v, err = req.respBodyParser.ParseBytes(req.respBodyBuf)

	return jsonResp, err
}

type response struct {
	raw *http.Response
}

func readToBuf(buf []byte, r io.Reader) ([]byte, error) {
	buf = buf[:cap(buf)]
	var n int
	for n < len(buf) {
		m, err := r.Read(buf[n:])
		n += m
		if err == io.EOF {
			return buf[:n], nil
		} else if err != nil {
			return nil, err
		}
	}
	return nil, fmt.Errorf("full buffer")
}
