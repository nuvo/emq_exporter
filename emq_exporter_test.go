package main

import (
	"io/ioutil"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/prometheus/client_golang/prometheus"
)

var _ = Describe("Utility Functions", func() {

	Describe("Loading credentials", func() {

		Context("loading from env vars", func() {

			AfterEach(func() {
				os.Unsetenv(usernameEnv)
				os.Unsetenv(passwordEnv)
			})

			It("should fail when no env var is set", func() {
				u, p, err := loadFromEnv()

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Equal("Can't find EMQ_USERNAME"))

				Expect(u).Should(BeEmpty())
				Expect(p).Should(BeEmpty())
			})

			It("Should fail when EMQ_USERNAME isn't set", func() {
				os.Setenv(passwordEnv, "secret")

				u, p, err := loadFromEnv()

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Equal("Can't find EMQ_USERNAME"))

				Expect(u).Should(BeEmpty())
				Expect(p).Should(BeEmpty())
			})

			It("should fail when EMQ_PASSWORD isn't set", func() {
				os.Setenv(usernameEnv, "admin")

				u, p, err := loadFromEnv()

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Equal("Can't find EMQ_PASSWORD"))

				Expect(u).Should(Equal("admin"))
				Expect(p).Should(BeEmpty())
			})

			It("should succeed when both env vars are set", func() {
				os.Setenv(usernameEnv, "admin")
				os.Setenv(passwordEnv, "secret")

				u, p, err := loadFromEnv()

				Expect(err).ShouldNot(HaveOccurred())

				Expect(u).Should(Equal("admin"))
				Expect(p).Should(Equal("secret"))
			})
		})

		Context("loading from file", func() {

			It("should fail on missing file", func() {
				path := "testdata/nothere.json"

				_, _, err := loadFromFile(path)
				Expect(err).Should(HaveOccurred())
			})

			It("should fail to unmarhsal invalid json data", func() {
				path := "testdata/malformed.txt"

				_, _, err := loadFromFile(path)
				Expect(err).Should(HaveOccurred())
			})

			It("should fail when username is missing", func() {
				path := "testdata/missinguser.json"

				u, p, err := loadFromFile(path)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Equal("missing username in testdata/missinguser.json"))

				Expect(u).Should(BeEmpty())
				Expect(p).Should(Equal("secret"))
			})

			It("should fail when password is missing", func() {
				path := "testdata/missingpass.json"

				u, p, err := loadFromFile(path)

				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(Equal("missing password in testdata/missingpass.json"))

				Expect(u).Should(Equal("Jonny"))
				Expect(p).Should(BeEmpty())
			})

			It("should succeed when both values are set", func() {
				path := "testdata/authfull.json"

				u, p, err := loadFromFile(path)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(u).Should(Equal("admin"))
				Expect(p).Should(Equal("public"))
			})
		})

		Context("finding credentails", func() {

			AfterEach(func() {
				os.Unsetenv(usernameEnv)
				os.Unsetenv(passwordEnv)
			})

			It("should return env vars", func() {
				os.Setenv(usernameEnv, "admin")
				os.Setenv(passwordEnv, "secret")
				path := "testdata/authfull.json"

				u, p, err := findCreds(path)

				Expect(err).ShouldNot(HaveOccurred())

				Expect(u).Should(Equal("admin"))
				Expect(p).Should(Equal("secret"))
			})

			It("should return values from file", func() {
				os.Setenv(usernameEnv, "admin")
				path := "testdata/authfull.json"

				u, p, err := findCreds(path)

				Expect(err).ShouldNot(HaveOccurred())

				Expect(u).Should(Equal("admin"))
				Expect(p).Should(Equal("public"))
			})

		})
	})

	Context("parsing strings", func() {

		It("should parse a simple float", func() {
			s := "0.5"

			v, err := parseString(s)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(v).Should(Equal(0.5))
		})

		It("should parse byte represented as string", func() {
			s := "123.19M"

			v, err := parseString(s)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(v).Should(Equal(1.29174077e+08))
		})

		It("should fail on invalid string", func() {
			s := "invalid string"

			v, err := parseString(s)

			Expect(err).Should(HaveOccurred())
			Expect(v).Should(Equal(float64(0)))
		})
	})

	Context("creating a new metric", func() {
		It("should return a valid metric", func() {
			m := metric{
				kind:  prometheus.GaugeValue,
				name:  "emq_node_memory_current",
				help:  "Current memory usage",
				value: 1.5533,
			}

			pm, err := newMetric(m)

			Expect(err).ToNot(HaveOccurred())
			Expect(pm.Desc().String()).To(Equal("Desc{fqName: \"emq_node_memory_current\", help: \"Current memory usage\", constLabels: {}, variableLabels: []}"))
		})

		It("should fail when the fqName isn't valid", func() {
			m := metric{
				kind:  prometheus.GaugeValue,
				name:  "/*3433##",
				help:  "Can't touch this",
				value: 0.003,
			}

			pm, err := newMetric(m)

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("\"/*3433##\" is not a valid metric name"))
			Expect(pm).To(BeNil())
		})
	})
})

//helper function to load json data from the testdata folder
func loadData(path string) []byte {
	b, err := ioutil.ReadFile("testdata/" + path)
	if err != nil {
		panic(err)
	}
	return b
}

var _ = Describe("Exporter", func() {
	const timeout = 5 * time.Second

	var (
		s *ghttp.Server
		e *Exporter

		//body []byte
	)

	BeforeEach(func() {
		s = ghttp.NewServer()

		c := &config{
			host:       s.URL(),
			username:   "admin",
			password:   "public",
			node:       "emq@" + s.URL(),
			apiVersion: "v3",
		}

		e = NewExporter(c, timeout)

	})

	AfterEach(func() {
		s.Close()
	})

	It("should send desc to the channel", func(done Done) {
		ch := make(chan *prometheus.Desc)

		go e.Describe(ch)
		Expect(<-ch).To(ContainSubstring("emq_up"))
		Expect(<-ch).To(ContainSubstring("emq_exporter_total_scrapes"))

		close(done)
	})
})
