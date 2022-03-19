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
	var req = requests.NewPost("example.com/test").
		BasicAuth("user", "secret").
		Query("key", "val").
		Header("token", &token)

	testReq(t, req, `POST example.com/test?key=val HTTP/1.1
Host: 
User-Agent: Go-http-client/1.1
Content-Length: 0
Authorization: Basic dXNlcjpzZWNyZXQ=
Token: secret

`)

	token = "super-secret"

	testReq(t, req, `POST example.com/test?key=val HTTP/1.1
Host: 
User-Agent: Go-http-client/1.1
Content-Length: 0
Authorization: Basic dXNlcjpzZWNyZXQ=
Token: super-secret

`)

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
		resp, err := requests.NewPost(url).JSONBody(map[string]interface{}{"foo": "bar"}).ExecJSON()
		is.NoErr(err)
		is.Equal(resp.Get("foo"), "bar")
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
