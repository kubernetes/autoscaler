package integration

import (
	"testing"

	ginkgo "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Integration Suite")
}
