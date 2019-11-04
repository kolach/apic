package apic_test

import (
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/kolach/apic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Client", func() {
	var server *ghttp.Server
	var NewRequest NewRequestFunc

	BeforeEach(func() {
		server = ghttp.NewServer()
		NewRequest = NewRequestFactory(WithBaseURL(server.URL()))
	})

	AfterEach(func() {
		server.Reset()
		server.Close()
	})

	Describe("Do", func() {
		Context("Without backoff", func() {
			var client *Client

			BeforeEach(func() {
				client = NewClient()
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/orders/1"),
						ghttp.RespondWith(http.StatusOK, "test"),
					),
				)
			})

			It("should make HTTP request", func() {
				req, _ := NewRequest("GET", "/api/orders/1", nil)
				res, err := client.Do(req, WithExpectStatus(http.StatusOK))
				Ω(server.ReceivedRequests()).Should(HaveLen(1))
				Ω(res.StatusCode).Should(Equal(http.StatusOK))
				Ω(err).ShouldNot(HaveOccurred())

				// check response body
				b, errRead := ioutil.ReadAll(res.Body)
				Ω(errRead).ShouldNot(HaveOccurred())
				Ω(b).Should(Equal([]byte("test")))
			})
		})

		Context("With backoff", func() {
			var (
				client  *Client
				retries int // to count retries
			)

			const (
				maxRetries        = 10
				successOnAtemptNo = 5
			)

			BeforeEach(func() {
				retries = 0 // reset retries counter

				notify := func(err error, dur time.Duration) {
					retries++ // count retries
				}

				client = NewClient(
					WithConstantBackOff(10*time.Millisecond),
					WithMaxRetries(maxRetries),
					WithNotify(notify),
				)

				// responnd with http.StatusForbidden to all /api/orders/101 requests
				for i := 0; i < maxRetries+1; i++ { // maxRetries + 1 (+1 because there is an original request too)
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/orders/101"),
							ghttp.RespondWith(http.StatusForbidden, "test"),
						),
					)
				}

				// but set 6th request to be OK
				server.SetHandler(successOnAtemptNo,
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/orders/101"),
						ghttp.RespondWith(http.StatusOK, "test"),
					),
				)
			})

			It("should make HTTP request", func() {
				req, _ := NewRequest("GET", "/api/orders/101", nil)
				res, err := client.Do(req, WithExpectStatus(http.StatusOK))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.StatusCode).Should(Equal(http.StatusOK))
				Ω(server.ReceivedRequests()).Should(HaveLen(successOnAtemptNo + 1))
				Ω(retries).Should(Equal(successOnAtemptNo)) // 5 because 6th one is success

				// check response body
				b, errRead := ioutil.ReadAll(res.Body)
				Ω(errRead).ShouldNot(HaveOccurred())
				Ω(b).Should(Equal([]byte("test")))
			})
		})
	})
})
