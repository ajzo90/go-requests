package requests

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type Retryer struct {
	doer          Doer
	nextTry       time.Time
	mtx           sync.RWMutex
	backoff       func() Backoffer
	sharedBackoff Backoffer
	retryPolicy   func(*http.Response, error) (bool, error)
	drainer       func(io.ReadCloser) error
	logger        RequestLogger
}

var backoff = func() Backoffer {
	min := time.Second / 10
	var attempts int
	return Backoff(func(resp *http.Response) time.Duration {
		attempts++
		return time.Duration(math.Pow(2, float64(attempts))) * min
	})
}

type RequestLogger interface {
	NextID() int
	Log(id int, err error, msg string)
}

func NewRetryer(doer Doer, logger RequestLogger) Doer {
	return &Retryer{doer: doer, backoff: backoff, sharedBackoff: DefaultSharedBackoff, retryPolicy: DefaultRetryPolicy, drainer: drain, logger: logger}
}

type Backoffer interface {
	Next(resp *http.Response) time.Duration
}

func (r *Retryer) retryAfter() time.Time {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	return r.nextTry
}

func (r *Retryer) updateRetryAfter(retryAfter time.Duration) {
	if retryAfter > 0 {
		newTs := time.Now().Add(retryAfter)
		r.mtx.Lock()
		if newTs.After(r.nextTry) {
			r.nextTry = newTs
		}
		r.mtx.Unlock()
	}
}

var (
	errCanNotResetBody    = fmt.Errorf("can not reset body")
	errDeadlineBeforeNext = fmt.Errorf("deadline is before next try")
)

func Logger(f func(id int, err error, msg string)) RequestLogger {
	return &defaultLogger{logger: f}
}

type defaultLogger struct {
	id     int
	mtx    sync.Mutex
	logger func(id int, err error, msg string)
}

func (d *defaultLogger) NextID() int {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	d.id++
	return d.id
}

func (d *defaultLogger) Log(id int, err error, msg string) {
	d.logger(id, err, msg)
}

func (r *Retryer) Do(request *http.Request) (_ *http.Response, err error) {
	var nextTry time.Time

	id := r.logger.NextID()
	defer func() {
		r.logger.Log(id, err, "done")
	}()

	w := bytes.NewBuffer(nil)
	if err := request.Write(w); err != nil {
		return nil, err
	}
	r.logger.Log(id, nil, w.String())

	backoff := r.backoff()
	for attempts := 0; ; attempts++ {
		if retryAfter := r.retryAfter(); retryAfter.After(nextTry) {
			nextTry = retryAfter
		}

		if deadline, ok := request.Context().Deadline(); ok {
			if deadline.Before(nextTry) {
				return nil, errDeadlineBeforeNext
			}
		}

		if nextTry.Before(time.Now()) {
			// ok
		} else if err := sleepUntil(request.Context(), nextTry); err != nil {
			return nil, err
		} else {
			continue // we need to inspect shared nextTry in case we slept for a long time
		}

		if err := request.Context().Err(); err != nil {
			return nil, err
		}

		request.Body, err = request.GetBody()
		if err != nil {
			return nil, err
		} else if request.Body == nil {
			return nil, errCanNotResetBody
		}

		resp, err := r.doer.Do(request)
		if err == nil {
			resp.Body = &logReaderCloser{rc: resp.Body, logger: func(n int) {
				r.logger.Log(id, nil, fmt.Sprintf("close %d", n))
			}}
			r.updateRetryAfter(r.sharedBackoff.Next(resp))
		}
		retry, retryErr := r.retryPolicy(resp, err)
		if !retry {
			if err == nil && retryErr != nil {
				_ = r.drainer(resp.Body)
			}
			return resp, retryErr
		}
		if resp != nil {
			r.logger.Log(id, retryErr, fmt.Sprintf("retry: %s", resp.Status))
			_ = r.drainer(resp.Body)
		}

		nextTry = time.Now().Add(backoff.Next(resp))
	}
}

type logReaderCloser struct {
	n      int
	rc     io.ReadCloser
	logger func(n int)
}

func (l *logReaderCloser) Read(p []byte) (n int, err error) {
	n, err = l.rc.Read(p)
	l.n += n
	return
}

func (l *logReaderCloser) Close() error {
	l.logger(l.n)
	return l.rc.Close()
}

func drain(rc io.ReadCloser) error {
	defer rc.Close()
	_, err := io.Copy(ioutil.Discard, io.LimitReader(rc, 4096))
	return err
}

type Backoff func(resp *http.Response) time.Duration

func (r Backoff) Next(resp *http.Response) time.Duration {
	return r(resp)
}

var DefaultSharedBackoff = Backoff(func(resp *http.Response) time.Duration {
	if resp.StatusCode == http.StatusTooManyRequests {
		if s := resp.Header.Get("Retry-After"); s != "" {
			if sleep, err := strconv.ParseInt(s, 10, 64); err == nil {
				return time.Second * time.Duration(sleep)
			} else if after, err := time.Parse(time.RFC1123, s); err == nil {
				return time.Until(after)
			}
		}
	}
	return 0
})

func DefaultRetryPolicy(resp *http.Response, err error) (bool, error) {
	if err != nil {
		return retryErr(resp, err)
	}
	return retryStatus(resp)
}

func retryStatus(resp *http.Response) (bool, error) {
	if resp.StatusCode == http.StatusTooManyRequests {
		return true, fmt.Errorf("too many requests")
	}

	if resp.StatusCode == 0 || (resp.StatusCode >= 500 && resp.StatusCode != http.StatusNotImplemented) {
		return true, fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}

	return false, nil
}

var (
	redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)
	schemeErrorRe    = regexp.MustCompile(`unsupported protocol scheme`)
	noHostInURLRe    = regexp.MustCompile(`no Host in request URL`)
)

func retryErr(resp *http.Response, err error) (bool, error) {
	if v, ok := err.(*url.Error); ok {
		// Don't retry if the error was due to too many redirects.
		if redirectsErrorRe.MatchString(v.Error()) {
			return false, v
		}

		// Don't retry if the error was due to an invalid protocol scheme.
		if schemeErrorRe.MatchString(v.Error()) {
			return false, v
		}

		if noHostInURLRe.MatchString(v.Error()) {
			return false, v
		}

		// Don't retry if the error was due to TLS cert verification failure.
		if _, ok := v.Err.(x509.UnknownAuthorityError); ok {
			return false, v
		}

	}

	// The error is likely recoverable so retry.
	return true, err
}
