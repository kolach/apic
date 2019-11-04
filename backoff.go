package apic

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
)

// NewBackOffFunc is factory function type to pass to WithRetryNotify interceptor
type NewBackOffFunc func() backoff.BackOff

type seekNopCloser struct {
	io.ReadSeeker
}

// Close does nothing, we just conform to Closer interface
func (r *seekNopCloser) Close() error { return nil }

// newSeekNopCloser constructs seekNopCloser from given body
func newSeekNopCloser(body io.ReadCloser) (io.ReadCloser, error) {
	var b []byte
	var err error
	if b, err = ioutil.ReadAll(body); err != nil {
		return nil, errors.Wrap(err, "failed to read body")
	}
	return &seekNopCloser{bytes.NewReader(b)}, nil
}

// WithRetry wraps http request executor function with provided backoff policy.
func WithRetry(b backoff.BackOff) InterceptDoFunc {
	return WithRetryNotify(func() backoff.BackOff { return b }, nil)
}

// WithRetryNotify wraps http request executor function with provided backoff policy
// and report error on  each unsuccessful attempt.
func WithRetryNotify(b NewBackOffFunc, n backoff.Notify) InterceptDoFunc {
	return func(do DoFunc) DoFunc {
		return func(req *http.Request) (res *http.Response, err error) {
			// recover from panic in case we get it
			defer func() {
				if r := recover(); r != nil {
					err = r.(error)
				}
			}()

			// make request body seek-able for later re-use in retry attempts
			if req.Body != nil {
				if _, ok := req.Body.(io.Seeker); !ok {
					if req.Body, err = newSeekNopCloser(req.Body); err != nil {
						return nil, errors.Wrap(err, "failed to convert body to io.Seeker")
					}
				}
			}

			// perform request,
			// in case of failure, rewind request body to start to make it ready for a new request
			op := func() (err error) {
				if res, err = do(req); err != nil && req.Body != nil {
					// restore request body in case of failure to re-use req object
					seeker := req.Body.(io.Seeker)
					if _, err := seeker.Seek(0, io.SeekStart); err != nil {
						// never happened before, but just in case:
						// we panic here to get out of backoff loop and recover in deffered function.
						panic(errors.Wrap(err, "failed to seek to start"))
					}
				}
				return
			}

			err = backoff.RetryNotify(op, b(), n)
			return
		}
	}
}
