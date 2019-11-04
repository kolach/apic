package apic

import (
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// StatusError incapsulate HTTP error with unexpected status code
type StatusError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (err *StatusError) Error() string {
	return err.Status
}

// WithExpectStatus watch response status and returns StatusError in case
// expections are not met.
// Usage example:
//
// c := NewClient()
// req := http.NewRequest("GET", "https://example.com/101", nil)
// res, err := c.Do(req, WithExpectStatus(http.StatusOK))
//
func WithExpectStatus(status ...int) InterceptDoFunc {
	m := make(map[int]bool)
	for _, s := range status {
		m[s] = true
	}

	return func(do DoFunc) DoFunc {
		return func(req *http.Request) (*http.Response, error) {
			res, err := do(req)
			if err != nil {
				return nil, err
			}

			if _, ok := m[res.StatusCode]; !ok {
				var body []byte
				if res.Body != nil {
					defer res.Body.Close()

					body, err = ioutil.ReadAll(res.Body)
					if err != nil {
						return nil, errors.Wrap(err, "failed to read response body")
					}
				}
				return nil, &StatusError{
					StatusCode: res.StatusCode,
					Status:     res.Status,
					Body:       body,
				}
			}

			return res, nil
		}
	}
}
