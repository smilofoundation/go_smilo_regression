package functional

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"go-smilo/src/blockchain/regression/src/container"
)

var dockerNetwork *container.DockerNetwork

func TestSport(t *testing.T) {
	//t.SkipNow()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sport Test Suite")
}

var _ = BeforeSuite(func() {
	var err error
	dockerNetwork, err = container.NewDockerNetwork()
	Expect(err).To(BeNil())
})

var _ = AfterSuite(func() {
	err := dockerNetwork.Remove()
	Expect(err).To(BeNil())
})
