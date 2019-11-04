package apicutil

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/kolach/apic"
)

// WithDumpRequest writes request object
func WithDumpRequest(w io.Writer, body bool) apic.InterceptDoFunc {
	return func(do apic.DoFunc) apic.DoFunc {
		return func(req *http.Request) (*http.Response, error) {
			// Save a copy of this request for debugging.
			reqDump, err := httputil.DumpRequest(req, body)
			if err != nil {
				return nil, err
			}
			fmt.Fprintln(w, string(reqDump))
			return do(req)
		}
	}
}

// WithDumpResponse writes response object
func WithDumpResponse(w io.Writer, body bool) apic.InterceptDoFunc {
	return func(do apic.DoFunc) apic.DoFunc {
		return func(req *http.Request) (*http.Response, error) {
			res, err := do(req)
			resDump, err := httputil.DumpResponse(res, body)
			if err != nil {
				return nil, err
			}
			fmt.Fprintln(w, string(resDump))
			return res, err
		}
	}
}
