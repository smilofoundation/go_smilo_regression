package functional

import (
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	tests "go-smilo/src/blockchain/regression"
	"go-smilo/src/blockchain/regression/src/container"
)

func TestRecover(t *testing.T) {
}

var _ = Describe("TFS-03: Recoverability testing", func() {
	const (
		numberOfFullnodes = 4
	)
	var (
		blockchain container.Blockchain
	)

	BeforeEach(func() {
		blockchain = container.NewDefaultBlockchain(dockerNetwork, numberOfFullnodes)
		Expect(blockchain.Start(true)).To(BeNil())
	})

	AfterEach(func() {
		blockchain.Stop(true) // This will return container not found error since we stop one
		blockchain.Finalize()
	})

	It("TFS-03-01: Add fullnodes in a network with < 2F+1 fullnodes to > 2F+1", func(done Done) {
		By("The consensus should work at the beginning", func() {
			tests.WaitFor(blockchain.Fullnodes(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(5)).To(BeNil())
				wg.Done()
			})
		})

		numOfFullnodesToBeStopped := 2

		By("Stop several fullnodes until there are less than 2F+1 fullnodes", func() {
			tests.WaitFor(blockchain.Fullnodes()[:numOfFullnodesToBeStopped], func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.StopMining()).To(BeNil())
				wg.Done()
			})
		})

		By("The consensus should not work after resuming", func() {
			tests.WaitFor(blockchain.Fullnodes(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				// container.ErrNoBlock should be returned if we didn't see any block in 10 seconds
				Expect(geth.WaitForBlocks(1, 10*time.Second)).To(BeEquivalentTo(container.ErrNoBlock))
				wg.Done()
			})
		})

		By("Resume the stopped fullnodes", func() {
			tests.WaitFor(blockchain.Fullnodes()[:numOfFullnodesToBeStopped], func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.StartMining()).To(BeNil())
				wg.Done()
			})
		})

		By("The consensus should work after resuming", func() {
			tests.WaitFor(blockchain.Fullnodes(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(5)).To(BeNil())
				wg.Done()
			})
		})

		close(done)
	}, 120)
})
