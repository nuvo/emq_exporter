package main

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Main", func() {
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

	Describe("Utility functions", func() {
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
	})
})
