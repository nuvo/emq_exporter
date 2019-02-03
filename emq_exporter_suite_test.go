package main

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEmqExporter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EmqExporter Suite")
}
