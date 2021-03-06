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
	r, err := req.NewRequestContext(context.Background(), true)
	if err != nil {
		return err
	}
	return r.Write(w)
}

func (req *ExtendedRequest) fullUrl(render func(stringer) string) string {
	base := render(req.baseUrl)
	if p := render(req.path); p != "" {
		return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(p, "/")
	}
	return base
}

func (req *ExtendedRequest) renderFn(masked bool) func(st stringer) string {
	mapper := func(s string) string {
		return s
	}
	if masked {
		mapper = func(s string) string {
			return strings.Repeat("x", len(s))
		}
	}

	return func(st stringer) string {
		s := st.String()
		for k, v := range req.secrets {
			s = strings.ReplaceAll(s, k, mapper(v.String()))
		}
		return s
	}
}

// NewRequestContext builds a *http.Request
func (req *ExtendedRequest) NewRequestContext(ctx context.Context, masked bool) (*http.Request, error) {
	if req.method == nil {
		req.method = toStringer(http.MethodGet)
	}
	if err := req.err; err != nil {
		return nil, err
	}
	renderer := req.renderFn(masked)

	var body io.Reader
	if req.body != nil {
		body = bytes.NewReader([]byte(renderer(req.body)))
	}

	request, err := http.NewRequestWithContext(ctx, renderer(req.method), req.fullUrl(renderer), body)
	if err != nil {
		return nil, err
	}

	for k, v := range req.header {
		request.Header.Add(k, renderer(v))
	}

	if len(req.query) > 0 {
		if request.URL.RawQuery != "" {
			return nil, fmt.Errorf("raw query and query param not allowed")
		}
		q := request.URL.Query()
		for k, v := range req.query {
			q.Add(k, renderer(v))
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
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, req.timeout)
		defer cancel()
	}

	request, err := req.NewRequestContext(ctx, false)
	if err != nil {
		return nil, err
	}
	return req.doer.Do(request)
}

func (req *ExtendedRequest) Doer(client Doer) *ExtendedRequest {
	req.doer = client
	return req
}

//// Reset the request
//func (req *ExtendedRequest) Reset() {
//	req.method = nil
//	req.baseUrl = nil
//	req.path = nil
//	req.body = nil
//	req.err = nil
//	req.header.Reset()
//	req.query.Reset()
//}

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
	req.secrets.CopyTo(newClient.secrets)
	return newClient
}
