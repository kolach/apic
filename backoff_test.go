package apic_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kolach/apic"
	. "github.com/kolach/gomega-matchers"
)

var errSeek = fmt.Errorf("failed to seek")

type failOnSeek struct {
	b *bytes.Buffer
}

func (r *failOnSeek) Read(p []byte) (int, error) { return r.b.Read(p) }
func (r *failOnSeek) Close() error               { return nil }
func (r *failOnSeek) Seek(offset int64, whence int) (int64, error) {
	return 0, errSeek
}

func failWith(err error, count *int) DoFunc {
	return func(req *http.Request) (*http.Response, error) {
		*count++
		// emulate request body read and close
		defer req.Body.Close()
		if _, err := ioutil.ReadAll(req.Body); err != nil {
			return nil, err
		}
		return nil, err
	}
}

var _ = Describe("Retry", func() {
	const maxRetries = 7
	var (
		count int
		b     backoff.BackOff
	)

	BeforeEach(func() {
		count = 0
		b = backoff.WithMaxRetries(backoff.NewConstantBackOff(10*time.Millisecond), maxRetries)
	})

	It("should retry request according backoff policy", func() {
		req, _ := http.NewRequest("POST", "https://example.com", bytes.NewBufferString("Buy iPhoneX"))
		res, err := WithRetry(b)(failWith(fmt.Errorf("Error"), &count))(req)

		Ω(count).Should(Equal(maxRetries + 1))
		Ω(res).Should(BeNil())
		Ω(err).Should(MatchError("Error"))
	})

	Context("On request body read error", func() {
		It("should fail immediately", func() {
			req, _ := http.NewRequest("POST", "https://example.com", new(failOnRead))
			res, err := WithRetry(b)(failWith(fmt.Errorf("Error"), &count))(req)

			Ω(count).Should(Equal(0))
			Ω(res).Should(BeNil())
			Ω(err).Should(BeCausedBy(errRead))
		})
	})

	Context("On request body seek error", func() {
		It("should fail with seek error", func() {
			req, _ := http.NewRequest("POST", "https://example.com", &failOnSeek{bytes.NewBufferString("foo")})
			res, err := WithRetry(b)(failWith(fmt.Errorf("Error"), &count))(req)

			Ω(count).Should(Equal(1))
			Ω(res).Should(BeNil())
			Ω(err).Should(BeCausedBy(errSeek))
		})
	})
})
