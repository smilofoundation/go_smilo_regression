package functional_test

import (
	"context"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	tests "go-smilo/src/blockchain/regression"
	"go-smilo/src/blockchain/regression/src/container"
)

var _ = Describe("QFS-07: Gossip Network", func() {
	const (
		numberOfFullnodes = 4
	)
	var (
		vaultNetwork container.VaultNetwork
		blockchain   container.Blockchain
		err          error
	)

	BeforeEach(func() {
		vaultNetwork, err = container.NewDefaultVaultNetwork(dockerNetwork, numberOfFullnodes)
		Expect(err).To(BeNil())
		Expect(vaultNetwork).ToNot(BeNil())
		Expect(vaultNetwork.Start()).To(BeNil())
		blockchain, err = container.NewDefaultSmiloBlockchain(dockerNetwork, vaultNetwork)
		Expect(err).To(BeNil())
		Expect(blockchain).ToNot(BeNil())
		Expect(blockchain.Start(false)).To(BeNil())
	})

	AfterEach(func() {
		blockchain.Stop(true)
		blockchain.Finalize()
		vaultNetwork.Stop()
		vaultNetwork.Finalize()
	})

	It("QFS-07-01: Gossip Network", func(done Done) {
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
