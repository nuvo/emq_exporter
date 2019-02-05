package main

import (
	"math/rand"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

//helper function to create random floats
func randFloat() float64 {
	return rand.Float64() * 1000
}

//mock fetcher for testing
type mockFetcher struct{}

func (m *mockFetcher) Fetch() (data map[string]interface{}, err error) {
	data = map[string]interface{}{
		"nodes_metrics_messages_qos1_sent":      randFloat(),
		"nodes_metrics_packets_pubrel_missed":   randFloat(),
		"nodes_metrics_packets_puback_sent":     randFloat(),
		"nodes_metrics_messages_received":       randFloat(),
		"nodes_metrics_packets_unsuback":        randFloat(),
		"nodes_metrics_packets_pubrel_sent":     randFloat(),
		"nodes_metrics_packets_subscribe":       randFloat(),
		"nodes_metrics_packets_connack":         randFloat(),
		"nodes_metrics_packets_disconnect_sent": randFloat(),
		"nodes_metrics_packets_pubcomp_sent":    randFloat(),
		"nodes_metrics_packets_unsubscribe":     randFloat(),
		"nodes_metrics_packets_auth":            randFloat(),
		"nodes_metrics_packets_suback":          randFloat(),
		"nodes_metrics_packets_pubrec_received": randFloat(),
		"nodes_metrics_messages_expired":        randFloat(),
		"nodes_metrics_messages_qos2_received":  randFloat(),
		"nodes_metrics_packets_sent":            randFloat(),
		"nodes_metrics_packets_pubrel_received": randFloat(),
		"nodes_metrics_messages_qos0_received":  randFloat(),
		"nodes_connections":                     randFloat(),
		"nodes_load1":                           "1.26",
		"nodes_load15":                          "1.08",
		"nodes_load5":                           "1.19",
		"nodes_max_fds":                         1048576,
		"nodes_memory_total":                    155385856,
		"nodes_memory_used":                     114687840,
		"nodes_name":                            "emqx@172.17.0.2",
		"nodes_node_status":                     "Running",
		"nodes_otp_release":                     "R21/10.2.1",
		"nodes_process_available":               2097152,
		"nodes_process_used":                    388,
		"nodes_uptime":                          "3 hours, 46 minutes, 36 seconds",
		"nodes_version":                         "v3.0.1",
	}

	return
}

var _ = Describe("Exporter", func() {

	var (
		e *Exporter
		f *mockFetcher
	)

	BeforeEach(func() {
		f = &mockFetcher{}
		e = NewExporter(f)
	})

	It("should send description to the channel", func(done Done) {
		ch := make(chan *prometheus.Desc)

		go e.Describe(ch)
		Eventually(ch).Should(Receive())

		close(done)
	})

	It("should send metrics to the channel", func(done Done) {
		ch := make(chan prometheus.Metric)

		//run multiple Collect() goroutines to make sure:
		//1. no data race (go test . -race)
		//2. metrics are being updated properly
		for i := 0; i < 1000; i++ {
			go e.Collect(ch)
			Eventually(ch).Should(Receive())
		}

		close(done)
	}, 5)
})
