package requests_test

import (
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/ajzo90/go-requests"
	"github.com/matryer/is"
)

var doer = requests.NewRetryer(http.DefaultClient, requests.Logger(func(id int, err error, msg string) {
	log.Println(id, err, msg)
}))

func TestRetryer_Do(t *testing.T) {
	var attempt int
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Add("foo", "bar")
		io.Copy(w, r.Body)
	}, func(t *testing.T, url string) {
		builder := requests.New(url).
			Method(http.MethodGet).
			Path("xx").
			JSONBody("hello").
			Timeout(time.Second*6).
			Header("Auth", "${key}").
			Header("Miss", "${miss}").
			SecretHeader("my-header", "secret2").
			BasicAuth("christian", "secret3").
			Secret("key", "secret")

		resp, err := builder.Extended().Doer(doer).ExecJSON()

		is := is.New(t)
		is.NoErr(err)
		is.Equal(resp.String(), "hello")
		is.Equal(resp.Header("foo"), "bar")
	})
}
