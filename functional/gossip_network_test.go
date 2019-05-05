package functional

import (
	"context"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	tests "go-smilo/src/blockchain/regression"
	"go-smilo/src/blockchain/regression/src/container"
)

func TestGossip(t *testing.T) {
}

var _ = Describe("TFS-07: Gossip Network", func() {
	const (
		numberOfFullnodes = 4
	)
	var (
		blockchain container.Blockchain
	)

	BeforeEach(func() {
		blockchain = container.NewDefaultBlockchain(dockerNetwork, numberOfFullnodes)
		Expect(blockchain.Start(false)).To(BeNil())
	})

	AfterEach(func() {
		Expect(blockchain.Stop(true)).To(BeNil())
		blockchain.Finalize()
	})

	It("TFS-07-01: Gossip Network", func(done Done) {
		By("Check peer count", func() {
			for _, geth := range blockchain.Fullnodes() {
				c := geth.NewClient()
				peers, e := c.AdminPeers(context.Background())
				Expect(e).To(BeNil())
				Ω(len(peers)).Should(BeNumerically("<=", 2))
			}
		})

		By("Checking blockchain progress", func() {
			tests.WaitFor(blockchain.Fullnodes(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(3)).To(BeNil())
				wg.Done()
			})
		})

		close(done)
	}, 240)
})
