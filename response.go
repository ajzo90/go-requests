package requests

import (
	"fmt"
	"io"
	"net/http"

	"github.com/valyala/fastjson"
)

type JSONResponse struct {
	response
	v *fastjson.Value
}

func (r *JSONResponse) Get(keys ...string) string {
	return string(r.v.GetStringBytes(keys...))
}
func (r *JSONResponse) GetInt(keys ...string) int {
	return r.v.GetInt(keys...)
}
func (r *JSONResponse) GetArray(keys ...string) []*fastjson.Value {
	return r.v.GetArray(keys...)
}

func (req *Request) ExecJSON() (*JSONResponse, error) {
	req.Header("accept", applicationJSON)
	req.respBodyBuf = make([]byte, 1000)

	resp, err := req.Extended().Do()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if req.respBodyBuf, err = ReadToBuf(req.respBodyBuf, resp.Body); err != nil {
		return nil, err
	}

	if err := resp.Body.Close(); err != nil {
		return nil, err
	}

	var jsonResp = &JSONResponse{response: response{raw: resp}}

	if jsonResp.v, err = req.respBodyParser.ParseBytes(req.respBodyBuf); err != nil {
		return nil, err
	}

	return jsonResp, nil
}

type response struct {
	raw *http.Response
}

func ReadToBuf(buf []byte, r io.Reader) ([]byte, error) {
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
