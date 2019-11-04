package apic_test

import (
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kolach/apic"
)

// sample config function
func WithAuthToken(token string) RequestOptionFunc {
	return func(req *http.Request) (*http.Request, error) {
		req.Header.Set("Authorization", "Bearer "+token)
		return req, nil
	}
}

// request config to trigger failure
func WithFailure(err error) RequestOptionFunc {
	return func(*http.Request) (*http.Request, error) {
		return nil, err
	}
}

// to capture if config has been called
func WithCallCount(called *bool) RequestOptionFunc {
	*called = false
	return func(req *http.Request) (*http.Request, error) {
		*called = true
		return req, nil
	}
}

var _ = Describe("Request", func() {
	var req *http.Request

	BeforeEach(func() {
		req, _ = http.NewRequest("GET", "/orders/1", nil)
	})

	Describe("WithBaseURL", func() {
		It("should prefix request URL with baseURL", func() {
			req, err := WithBaseURL("https://api.example.com/v1")(req)
			Ω(err).Should(BeNil())
			Ω(req.URL.String()).Should(Equal("https://api.example.com/v1/orders/1"))
		})

		It("should fail if baseURL parsing fails", func() {
			_, err := WithBaseURL(":api.example.com/v1")(req)
			Ω(err).ShouldNot(Succeed())
		})
	})

	Describe("WithContext", func() {
		It("should return new request with context", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			req, _ = WithContext(ctx)(req)
			Ω(req.Context()).To(Equal(ctx))
		})
	})

	Describe("NewRequest", func() {
		It("should create request and configure it", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			req, err := NewRequest("GET", "/orders/1/items", nil,
				WithContext(ctx),
				WithBaseURL("https://api.example.com/v1"),
				WithAuthToken("abc123"),
			)

			Ω(err).Should(BeNil())
			Ω(req.Context()).Should(BeIdenticalTo(ctx))
			Ω(req.URL.String()).Should(Equal("https://api.example.com/v1/orders/1/items"))
			Ω(req.Header.Get("Authorization")).Should(Equal("Bearer abc123"))
		})

		Context("When initial request params are bad", func() {
			It("should fail even before config functions are applied", func() {
				var called bool
				req, err := NewRequest("GET", "://example.com", nil, WithCallCount(&called))
				Ω(req).Should(BeNil())
				Ω(err).ShouldNot(BeNil())
				Ω(called).Should(BeFalse())
			})
		})

		Context("When one of config functions fail with error", func() {
			It("should return the error and exit", func() {
				var called bool
				err := fmt.Errorf("Error")
				req, err := NewRequest("GET", "https://example.com", nil, WithFailure(err), WithCallCount(&called))
				Ω(req).Should(BeNil())
				Ω(err).Should(BeIdenticalTo(err))
				Ω(called).Should(BeFalse())
			})
		})
	})

	Describe("NewRequestFactory", func() {
		It("should capture high level configuration", func() {
			newReq := NewRequestFactory(
				WithBaseURL("https://api.example.com/v1"),
				WithAuthToken("abc123"),
			)
			req, err := newReq("GET", "/orders/2", nil)
			Ω(err).Should(BeNil())
			Ω(req.URL.String()).Should(Equal("https://api.example.com/v1/orders/2"))
			Ω(req.Header.Get("Authorization")).Should(Equal("Bearer abc123"))
		})
	})
})
