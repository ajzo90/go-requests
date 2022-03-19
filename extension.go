package requests

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// ExtendedRequest contain more advanced functionality.
// The purpose is to separate those from the visible methods in Request to make the package easier to use
type ExtendedRequest struct {
	*Request
}

// Extended returns the *ExtendedRequest
func (req *Request) Extended() *ExtendedRequest {
	return &ExtendedRequest{req}
}

// Write writes the request into w
func (req *ExtendedRequest) Write(w io.Writer) error {
	r, err := req.NewRequest()
	if err != nil {
		return err
	}
	return r.Write(w)
}

// NewRequest builds a *http.Request
func (req *ExtendedRequest) NewRequest() (*http.Request, error) {
	if req.method == nil {
		req.method = toStringer(http.MethodGet)
	}

	if err := req.err; err != nil {
		return nil, err
	}

	var bodyR io.Reader

	if req.body != nil {
		bodyR = bytes.NewBufferString(req.body.String())
	}

	request, err := http.NewRequest(req.method.String(), req.url.String(), bodyR)
	if err != nil {
		return nil, err
	}
	for k, v := range req.header {
		request.Header.Add(k, v.String())
	}

	if len(req.query) > 0 {
		if request.URL.RawQuery != "" {
			return nil, fmt.Errorf("raw query and query param not allowed")
		}
		q := request.URL.Query()
		for k, v := range req.query {
			q.Add(k, v.String())
		}
		request.URL.RawQuery = q.Encode()
	}

	return request, err
}

// Do execute do the request. Caller must close resp.Body in case of non-nil error
func (req *ExtendedRequest) Do() (*http.Response, error) {
	request, err := req.NewRequest()
	if err != nil {
		return nil, err
	}
	return req.doer.Do(request)
}

// Clone clones the *Request to allow concurrent usage of the same base configuration
func (req *ExtendedRequest) Clone() *Request {
	newClient := New("")
	newClient.body = req.body
	newClient.url = req.url
	newClient.method = req.method
	newClient.err = req.err
	for k, v := range req.header {
		newClient.header[k] = v
	}
	for k, v := range req.query {
		newClient.query[k] = v
	}
	return newClient
}
