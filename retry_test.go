package requests_test

import (
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ajzo90/go-requests"
	"github.com/matryer/is"
)

var logger = requests.Logger(func(id int, err error, msg string) {
	log.Println(id, err, msg)
})

var doer = requests.NewRetryer(http.DefaultClient, logger)

var retryerOption = requests.WithRetryPolicy(func(resp *http.Response, err error) (bool, error) {
	if resp.StatusCode == http.StatusInternalServerError {
		return true, err
	}
	return false, nil
})
var doerWithRetryPolicy = requests.NewRetryer(http.DefaultClient, logger, retryerOption)

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
		req := requests.New(url).
			Method(http.MethodGet).
			Path("/foo/bar").
			Query("k", "${key}").
			Header("user-agent", "x").
			Header("Auth", "${key}").
			Header("Miss", "${miss}").
			SecretHeader("my-header", "secret2").
			BasicAuth("christian", "secret3").
			JSONBody("hello").
			WithExtended(func(req *requests.ExtendedRequest) {
				req.Doer(doer)
				req.Secret("key", "secret")
				req.Timeout(time.Second * 6)
				_ = req.Write(os.Stdout)
			})

		resp, err := req.ExecJSON()
		is := is.New(t)
		is.NoErr(err)
		is.Equal(resp.String(), "hello")
		is.Equal(resp.Header("foo"), "bar")
	})
}

func TestNewRetryerWithRetryPolicy_Do(t *testing.T) {
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
		req := requests.New(url).
			Method(http.MethodGet).
			Path("/foo/bar").
			JSONBody("hello").
			WithExtended(func(req *requests.ExtendedRequest) {
				req.Doer(doerWithRetryPolicy)
				_ = req.Write(os.Stdout)
			})

		resp, err := req.ExecJSON()
		is := is.New(t)
		is.NoErr(err)
		is.Equal(resp.String(), "hello")
		is.Equal(resp.Header("foo"), "bar")
	})
}
