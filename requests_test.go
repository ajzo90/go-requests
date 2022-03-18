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

func withTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request), fn func(t *testing.T, url string)) {
	srv := httptest.NewServer(http.HandlerFunc(handler))
	defer srv.Close()
	fn(t, srv.URL)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	io.Copy(w, r.Body)
}

func TestX(t *testing.T) {
	is := is.New(t)
	var w = &bytes.Buffer{}
	var req = requests.NewPost("example.com/test").
		Query("key", "val").
		Header("token", "secret")

	is.NoErr(req.Extended().Write(w))

	const expected = `POST example.com/test?key=val HTTP/1.1
Host: 
User-Agent: Go-http-client/1.1
Content-Length: 0
Token: secret

`
	var res = strings.ReplaceAll(w.String(), "\r\n", "\n")
	is.Equal(res, expected)

}
