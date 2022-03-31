package requests_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ajzo90/go-requests"
	"github.com/matryer/is"
)

type constStr string

func (s constStr) String() string {
	return string(s)
}

func TestX(t *testing.T) {
	token := "secret"
	req := requests.NewGet("example.com/test").
		BasicAuth("user", "secret").
		Query("key", constStr("val")).
		Header("token", &token)

	testReq(t, req, `GET example.com/test?key=val HTTP/1.1
Host: 
User-Agent: Go-http-client/1.1
Authorization: **********************
Token: secret

`)

	token = "super-secret"

	testReq(t, req.Extended().Clone(), `GET example.com/test?key=val HTTP/1.1
Host: 
User-Agent: Go-http-client/1.1
Authorization: **********************
Token: super-secret

`)

	err := requests.NewGet("example.com?a=1").Query("b", "2").Extended().Write(nil)
	is := is.New(t)
	is.True(err != nil)
	is.Equal(err.Error(), `raw query and query param not allowed`)
}

func TestIncorrectUsage(t *testing.T) {
	withTestServer(t, echoHandler, func(t *testing.T, url string) {
		tests := []struct {
			r   *requests.Request
			err string
		}{
			{r: requests.New(123), err: "can not convert 123 to stringer"},
			{r: requests.New("localhost"), err: `Get "localhost": unsupported protocol scheme ""`},
			{r: requests.New("https://example.com?test=1").Query("foo", "bar"), err: `raw query and query param not allowed`},
			{r: requests.New("https://example.com").Method("?"), err: `net/http: invalid method "?"`},
		}

		for _, test := range tests {
			t.Run(test.err, func(t *testing.T) {
				is := is.New(t)
				_, err := test.r.Extended().Clone().ExecJSON(context.Background())
				is.True(err != nil)
				is.Equal(err.Error(), test.err)

				_, err = test.r.ExecJSON(context.Background())
				is.True(err != nil)
				is.Equal(err.Error(), test.err)
			})
		}
	})
}

func testReq(t *testing.T, req *requests.Request, expected string) {
	is := is.New(t)
	w := &bytes.Buffer{}
	is.NoErr(req.Extended().Write(w))
	res := strings.ReplaceAll(w.String(), "\r\n", "\n")
	is.Equal(res, expected)
}

func TestRequest_ExecJSON(t *testing.T) {
	withTestServer(t, echoHandler, func(t *testing.T, url string) {
		is := is.New(t)
		q := requests.NewPost(url).JSONBody(map[string]interface{}{"foo": "bar", "baz": 1, "arr": []int{1, 2}})
		for _, q := range []*requests.Request{q, q.Extended().Clone()} {
			resp, err := q.ExecJSON(context.Background())
			is.NoErr(err)
			is.Equal(resp.String("foo"), "bar")
			is.Equal(resp.Int("baz"), 1)
			is.Equal(resp.Body().GetInt("baz"), 1)
			is.Equal(len(resp.GetArray("arr")), 2)
		}
	})
}

func withTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request), fn func(t *testing.T, url string)) {
	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()
	fn(t, srv.URL)
}

func TestInvalidBody(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1")
	}, func(t *testing.T, url string) {
		_, err := requests.NewGet(url).ExecJSON(context.Background())
		is := is.New(t)
		is.True(err != nil)
		is.Equal(err.Error(), `unexpected EOF`)
	})
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)

	// by some reason io.Copy fails at 512 bytes payload
	if err != nil {
		panic(err)
	} else if _, err := w.Write(b); err != nil {
		panic(err)
	}
}

func Test404Req(t *testing.T) {
	withTestServer(t, errHandler, func(t *testing.T, url string) {
		_, err := requests.New(url).ExecJSON(context.Background())
		is := is.New(t)
		is.True(err != nil)
		is.Equal(err.Error(), "invalid status 404 Not Found")
	})
}

func errHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
}
