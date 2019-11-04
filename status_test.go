package apic_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kolach/apic"
	. "github.com/kolach/gomega-matchers"
)

var errRead = fmt.Errorf("failed to read")

type failOnRead int

func (r failOnRead) Read([]byte) (int, error) {
	return 0, errRead
}

var _ = Describe("WithExpectStatus", func() {
	var req *http.Request

	BeforeEach(func() {
		req, _ = NewRequest("GET", "https://example.com/orders/1", nil)
	})

	Describe("StatusError", func() {
		Describe("Error()", func() {
			It("should return response status text", func() {
				err := &StatusError{
					StatusCode: 404,
					Status:     "Not Found",
					Body:       []byte("Order not found"),
				}
				Ω(err.Error()).Should(Equal("Not Found"))
			})
		})
	})

	Context("When response status is expected", func() {
		It("should return response", func() {
			res, err := WithExpectStatus(200)(func(req *http.Request) (*http.Response, error) {
				res := new(http.Response)
				res.StatusCode = 200
				return res, nil
			})(req)
			Ω(err).Should(BeNil())
			Ω(res).Should(BeIdenticalTo(res))
		})
	})

	Context("When response status is NOT expected", func() {
		var (
			notFoundStatusCode = 404
			notFoundStatus     = "Not Found"
			notFoundBody       = []byte("Order not found")
		)

		It("should fail with error", func() {
			res, err := WithExpectStatus(200)(func(req *http.Request) (*http.Response, error) {
				res := new(http.Response)
				res.StatusCode = notFoundStatusCode
				res.Status = notFoundStatus
				res.Body = ioutil.NopCloser(bytes.NewBuffer(notFoundBody))
				return res, nil
			})(req)
			Ω(err).Should(MatchError(&StatusError{
				StatusCode: notFoundStatusCode,
				Status:     notFoundStatus,
				Body:       notFoundBody,
			}))
			Ω(res).Should(BeNil())
		})
	})

	Context("When underlying do function returns error", func() {
		It("should forward the error", func() {
			reqErr := fmt.Errorf("Error")
			_, err := WithExpectStatus(200)(func(req *http.Request) (*http.Response, error) {
				return nil, reqErr
			})(req)
			Ω(err).Should(BeIdenticalTo(reqErr))
		})
	})

	Context("On failed read of response body", func() {
		It("should return body read error", func() {
			res, err := WithExpectStatus(200)(func(req *http.Request) (*http.Response, error) {
				res := new(http.Response)
				res.Body = ioutil.NopCloser(new(failOnRead))
				return res, nil
			})(req)
			Ω(res).Should(BeNil())
			Ω(err).Should(BeCausedBy(errRead))
		})
	})
})
