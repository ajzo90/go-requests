package requests_test

import (
	"bytes"
	"github.com/ajzo90/go-requests"
	"github.com/matryer/is"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestX(t *testing.T) {

	var token = "secret"
	var req = requests.NewGet("example.com/test").
		BasicAuth("user", "secret").
		Query("key", "val").
		Header("token", &token)

	testReq(t, req, `GET example.com/test?key=val HTTP/1.1
Host: 
User-Agent: Go-http-client/1.1
Authorization: Basic dXNlcjpzZWNyZXQ=
Token: secret

`)

	token = "super-secret"

	testReq(t, req.Extended().Clone(), `GET example.com/test?key=val HTTP/1.1
Host: 
User-Agent: Go-http-client/1.1
Authorization: Basic dXNlcjpzZWNyZXQ=
Token: super-secret

`)

}

func TestIncorrectUsage(t *testing.T) {
	var tests = []struct {
		r   *requests.Request
		err string
	}{
		{r: requests.New(123), err: "can not convert 123 to stringer"},
		{r: requests.New("localhost"), err: `Get "localhost": unsupported protocol scheme ""`},
		{r: requests.New("https://example.com?test=1").Query("foo", "bar"), err: `raw query and query param not allowed`},
	}

	for _, test := range tests {
		t.Run(test.err, func(t *testing.T) {
			is := is.New(t)
			_, err := test.r.ExecJSON()
			is.True(err != nil)
			is.Equal(err.Error(), test.err)
		})
	}

}

func testReq(t *testing.T, req *requests.Request, expected string) {
	is := is.New(t)
	var w = &bytes.Buffer{}
	is.NoErr(req.Extended().Write(w))
	var res = strings.ReplaceAll(w.String(), "\r\n", "\n")
	is.Equal(res, expected)
}

func TestRequest_ExecJSON(t *testing.T) {
	withTestServer(t, echoHandler, func(t *testing.T, url string) {
		is := is.New(t)
		resp, err := requests.NewPost(url).JSONBody(map[string]interface{}{"foo": "bar", "baz": 1, "arr": []int{1, 2}}).ExecJSON()
		is.NoErr(err)
		is.Equal(resp.Get("foo"), "bar")
		is.Equal(resp.GetInt("baz"), 1)
		is.Equal(len(resp.GetArray("arr")), 2)
	})
}

func withTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request), fn func(t *testing.T, url string)) {
	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()
	fn(t, srv.URL)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	io.Copy(w, r.Body)
}
