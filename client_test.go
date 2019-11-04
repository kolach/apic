package apic_test

import (
	"io/ioutil"
	"net/http"

	. "github.com/kolach/apic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Client", func() {
	var server *ghttp.Server
	var client *Client
	var NewRequest NewRequestFunc

	BeforeEach(func() {
		server = ghttp.NewServer()
		NewRequest = NewRequestFactory(WithBaseURL(server.URL()))
		client = NewClient()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Do", func() {
		BeforeEach(func() {
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
})
