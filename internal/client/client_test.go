package client

import (
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

//helper function to load json data from the testdata folder
func loadData(path string) []byte {
	b, err := ioutil.ReadFile("testdata/" + path)
	if err != nil {
		panic(err)
	}
	return b
}

var _ = Describe("Client", func() {

	var (
		s *ghttp.Server
		c *Client
	)

	BeforeEach(func() {
		s = ghttp.NewServer()
		//TODO(hagaibarel) add v2 tests
		c = NewClient(
			s.URL(),
			"emqx",
			"v3",
			"admin",
			"public",
			false,
		)
	})

	AfterEach(func() {
		s.Close()
	})

	Context("Successful fetching", func() {

		BeforeEach(func() {
			//register route handlers
			s.RouteToHandler("GET", "/api/v3/nodes/emqx/metrics/", ghttp.CombineHandlers(
				ghttp.VerifyBasicAuth("admin", "public"),
				ghttp.VerifyHeader(http.Header{
					"Accept": []string{"application/json"},
				}),
				ghttp.RespondWith(200, loadData("metrics.json")),
			))

			s.RouteToHandler("GET", "/api/v3/nodes/emqx/stats/", ghttp.CombineHandlers(
				ghttp.VerifyBasicAuth("admin", "public"),
				ghttp.VerifyHeader(http.Header{
					"Accept": []string{"application/json"},
				}),
				ghttp.RespondWith(200, loadData("stats.json")),
			))

			s.RouteToHandler("GET", "/api/v3/nodes/emqx", ghttp.CombineHandlers(
				ghttp.VerifyBasicAuth("admin", "public"),
				ghttp.VerifyHeader(http.Header{
					"Accept": []string{"application/json"},
				}),
				ghttp.RespondWith(200, loadData("node.json")),
			))
		})

		It("should succeed fetching the metrics", func() {
			res, err := c.Fetch()

			Expect(err).ShouldNot(HaveOccurred())
			Expect(res).To(HaveKeyWithValue("nodes_version", "v3.0.1"))
		})
	})

	Context("Failed requests", func() {

		var (
			statusCode int
			body       []byte
			path       = "/api/v3/nodes/%s/stats"
		)

		BeforeEach(func() {
			s.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v3/nodes/emqx/stats"),
					ghttp.VerifyBasicAuth("admin", "public"),
					ghttp.VerifyHeader(http.Header{
						"Accept": []string{"application/json"},
					}),
					ghttp.RespondWithPtr(&statusCode, &body),
				),
			)
		})

		It("should fail when the response status code is not 200", func() {
			statusCode = http.StatusNotFound
			body = loadData("badresponse.json")

			data, err := c.get(path)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Received status code not ok"))
			Expect(data).To(BeNil())
		})

		It("should fail when the response body isn't valid json", func() {
			statusCode = http.StatusOK
			body = []byte("not valid json")

			data, err := c.get(path)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Error in json decoder"))
			Expect(data).To(BeNil())
		})

		It("should fail when the response body has Code != 0", func() {
			statusCode = http.StatusOK
			body = loadData("badresponse.json")

			data, err := c.get(path)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Recvied code != 0"))
			Expect(data).To(BeNil())
		})

		It("should fail to create a request for a bad path", func() {
			path := "/api/v3/nodes/emqx/stats"
			statusCode = http.StatusOK
			body = loadData("badresponse.json")

			data, err := c.get(path)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to create http request"))
			Expect(data).To(BeNil())

		})

		It("should fail for bad urls", func() {
			statusCode = http.StatusOK
			body = loadData("badresponse.json")

			c.setHost("localhost:1859")

			data, err := c.get(path)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to get metrics"))
			Expect(data).To(BeNil())
		})

	})
})
