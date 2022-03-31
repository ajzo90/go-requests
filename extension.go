package requests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// WithExtended
func (req *Request) WithExtended(f func(request *ExtendedRequest)) *Request {
	f(req.Extended())
	return req
}

// Write writes the request into w
func (req *ExtendedRequest) Write(w io.Writer) error {
	r, err := req.NewRequestContext(context.Background())
	if err != nil {
		return err
	}
	return r.Write(w)
}

func (req *ExtendedRequest) fullUrl() string {
	base := req.baseUrl.String()
	if p := req.path.String(); p != "" {
		return strings.ReplaceAll(base+"/"+p, "//", "/")
	}
	return base
}

// NewRequestContext builds a *http.Request
func (req *ExtendedRequest) NewRequestContext(ctx context.Context) (*http.Request, error) {
	if req.method == nil {
		req.method = toStringer(http.MethodGet)
	}
	if err := req.err; err != nil {
		return nil, err
	}

	var body io.Reader
	if req.body != nil {
		body = bytes.NewReader([]byte(req.body.String()))
	}

	request, err := http.NewRequestWithContext(ctx, req.method.String(), req.fullUrl(), body)
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
func (req *ExtendedRequest) Do(ctxs ...context.Context) (*http.Response, error) {
	var ctx context.Context
	if len(ctxs) == 1 {
		ctx = ctxs[0]
	} else {
		ctx = context.Background()
	}

	if req.timeout != 0 {
		cancel := func() {}
		ctx, cancel = context.WithTimeout(ctx, req.timeout)
		defer cancel()
	}

	request, err := req.NewRequestContext(ctx)
	if err != nil {
		return nil, err
	}
	return req.doer.Do(request)
}

func (req *ExtendedRequest) Doer(client Doer) *ExtendedRequest {
	req.doer = client
	return req
}

// Reset the request
func (req *ExtendedRequest) Reset() {
	req.method = nil
	req.baseUrl = nil
	req.path = nil
	req.body = nil
	req.err = nil
	req.header.Reset()
	req.query.Reset()
}

// Clone clones the *Request to allow concurrent usage of the same base configuration
func (req *ExtendedRequest) Clone() *Request {
	newClient := New("")
	newClient.method = req.method
	newClient.baseUrl = req.baseUrl
	newClient.path = req.path
	newClient.body = req.body
	newClient.err = req.err
	newClient.doer = req.doer
	newClient.timeout = req.timeout
	req.header.CopyTo(newClient.header)
	req.query.CopyTo(newClient.query)
	return newClient
}
