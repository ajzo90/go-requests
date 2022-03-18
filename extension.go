package requests

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type ExtendedRequest struct {
	*Request
}

func (req *Request) Extended() *ExtendedRequest {
	return &ExtendedRequest{req}
}

func (req *ExtendedRequest) Write(w io.Writer) error {
	if r, err := req.NewRequest(); err != nil {
		return err
	} else {
		return r.Write(w)
	}
}

func (req *ExtendedRequest) NewRequest() (*http.Request, error) {
	var bodyR io.Reader

	if req.body != nil {
		bodyR = bytes.NewBufferString(req.body.String())
	}

	var request, err = http.NewRequest(req.method.String(), req.url.String(), bodyR)
	if err != nil {
		return nil, err
	}
	for k, v := range req.header {
		request.Header.Add(k, v.String())
	}
	var q = request.URL.Query()
	for k, v := range req.query {
		q.Add(k, v.String())
	}
	if qv := q.Encode(); qv != "" {
		if request.URL.RawQuery != "" {
			return nil, fmt.Errorf("raw query and query param not allowed")
		}
		request.URL.RawQuery = qv
	}

	if err := req.err; err != nil {
		return nil, err
	}
	return request, err
}

func (req *ExtendedRequest) Do() (*http.Response, error) {
	request, err := req.NewRequest()
	if err != nil {
		return nil, err
	}
	return req.doer.Do(request)
}

func (req *ExtendedRequest) Clone() *Request {

	var newClient = New(req.url).Method(req.method)

	newClient.body = req.body
	for k, v := range req.header {
		newClient.header[k] = v
	}
	for k, v := range req.query {
		newClient.query[k] = v
	}

	return newClient

}
