package apic

import (
	"context"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
)

// DoFunc is do-http-request function type
type DoFunc func(req *http.Request) (*http.Response, error)

// InterceptDoFunc is function type for do-http-request intercepted calls
type InterceptDoFunc func(DoFunc) DoFunc

// Client for resin.io service
type Client struct {
	client     *http.Client   // HTTP api client to make requests
	newBackOff NewBackOffFunc // new backoff factory function
	notify     backoff.Notify // backoff error notify callback
}

// Do performs HTTP request to resin.io in a given context.
// If the client is created with backoff option, the context is bound to generated backoff policy.
// Otherwise the context gets bound to request object.
func (c *Client) Do(req *http.Request, interceptors ...InterceptDoFunc) (*http.Response, error) {
	if c.newBackOff != nil {
		// extract context from request and set it to background
		// original request context is going to be used in backoff
		ctx := req.Context()
		req = req.WithContext(context.Background())

		// If backoff factory function is provided, bind context to backoff instance.
		b := func() backoff.BackOff { return backoff.WithContext(c.newBackOff(), ctx) }
		n := func(err error, d time.Duration) {
			if c.notify != nil {
				c.notify(err, d)
			}
		}
		interceptors = append(interceptors, WithRetryNotify(b, n))
	} else {
		// Otherwise bind context to request object.
		// req = req.WithContext(ctx)
	}

	// Uncomment to debug request/response
	// interceptors = append([]InterceptDoFunc{apiutil.WithDumpRequest(os.Stdout, true)}, interceptors...)
	// interceptors = append(interceptors, apiutil.WithDumpResponse(os.Stdout, true))

	do := c.client.Do
	for _, intercept := range interceptors {
		do = intercept(do)
	}

	// perform request
	return do(req)
}

// CtxClientCfgFunc is functional type to configure client
type CtxClientCfgFunc func(c *Client)

// WithHTTPClient sets http cloent to make requests
func WithHTTPClient(client *http.Client) CtxClientCfgFunc {
	return func(c *Client) {
		c.client = client
	}
}

// WithBackOff configures backoff factory
func WithBackOff(b NewBackOffFunc) CtxClientCfgFunc {
	return func(c *Client) {
		c.newBackOff = b
	}
}

// WithExponentialBackOff configures backoff factory
func WithExponentialBackOff() CtxClientCfgFunc {
	return WithBackOff(func() backoff.BackOff {
		return backoff.NewExponentialBackOff()
	})
}

// WithConstantBackOff configures backoff factory
func WithConstantBackOff(d time.Duration) CtxClientCfgFunc {
	return WithBackOff(func() backoff.BackOff {
		return backoff.NewConstantBackOff(d)
	})
}

// WithMaxRetries configures how many retries to make
// make sure you setup bsome backoff factory function before.
func WithMaxRetries(n uint64) CtxClientCfgFunc {
	return func(c *Client) {
		b := c.newBackOff
		c.newBackOff = func() backoff.BackOff {
			return backoff.WithMaxRetries(b(), n)
		}
	}
}

// WithNotify allows to setup external notify callback
func WithNotify(n backoff.Notify) CtxClientCfgFunc {
	return func(c *Client) {
		c.notify = n
	}
}

// NewClient constructs a new resin.io client
// all HTTP requests are done via provided
func NewClient(cfgs ...CtxClientCfgFunc) *Client {
	client := &Client{client: http.DefaultClient}
	for _, cfg := range cfgs {
		cfg(client)
	}
	return client
}
