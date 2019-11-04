package apic

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// RequestCfgFunc visitor to configure request
// Usage example:
//
// func WithAuthToken(token string) RequestCfgFunc {
// 	return func(req *http.Request) (*http.Request, error) {
// 		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
// 		return req, nil
//   }
// }
type RequestCfgFunc func(*http.Request) (*http.Request, error)

// WithContext binds context to request.
// The function is included mostly for example purposes.
// Usage example:
//
// req, err := api.NewRequest("GET", "/orders", nil, api.WithHostURL("https://example.com"))
//
func WithContext(ctx context.Context) RequestCfgFunc {
	return func(req *http.Request) (*http.Request, error) {
		return req.WithContext(ctx), nil
	}
}

// WithBaseURL assigns a host and scheme to request URL.
// Usefull with NewRequestFactory.
// Usage example:
//
// NewRequest := api.NewRequestFactory(api.WithBaseURL("https://example.com"))
// ...
// req, err = NewRequest("GET", "/orders", nil)
func WithBaseURL(hostURL string) RequestCfgFunc {
	return func(req *http.Request) (*http.Request, error) {
		u, err := url.Parse(hostURL)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse URL")
		}
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
		req.URL.Path = u.Path + req.URL.Path
		return req, nil
	}
}

// NewRequest creates HTTP preconfigured with resinio params
func NewRequest(method string, url string, body io.Reader, cfgs ...RequestCfgFunc) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	// apply all provided request configuration function
	for _, cfg := range cfgs {
		if req, err = cfg(req); err != nil {
			// exit with error in case of failure
			return nil, errors.Wrap(err, "failed to apply request configuration")
		}
	}
	return req, nil
}

// NewRequestFunc new request signature
type NewRequestFunc func(method string, url string, body io.Reader, cfgs ...RequestCfgFunc) (*http.Request, error)

// NewRequestFactory is factory function to capture some high level base params
// Usage example:
//
//	NewRequest = api.NewRequestFactory(
//		api.WithHostURL("https://example.com"),
//		WithAuthToken(token),
//  )
// req, err := NewRequest("GET", "/orders", nil)
func NewRequestFactory(baseCfgs ...RequestCfgFunc) NewRequestFunc {
	return func(method string, url string, body io.Reader, cfgs ...RequestCfgFunc) (*http.Request, error) {
		return NewRequest(method, url, body, append(baseCfgs, cfgs...)...)
	}
}
