package functional

import (
	"context"
	"math/big"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	tests "go-smilo/src/blockchain/regression"
	"go-smilo/src/blockchain/regression/src/container"
)

func TestBlockSync(t *testing.T) {
}

var _ = Describe("Block synchronization testing", func() {
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
		Expect(blockchain.Stop(true)).To(BeNil())
		blockchain.Finalize()
	})

	Describe("TFS-06: Block synchronization testing", func() {
		const numberOfNodes = 2
		var nodes []container.Ethereum

		BeforeEach(func() {
			var err error

			incubator, ok := blockchain.(container.NodeIncubator)
			Expect(ok).To(BeTrue())

			nodes, err = incubator.CreateNodes(numberOfNodes,
				container.ImageRepository("quay.io/smilo/go-smilo"),
				container.ImageTag("latest"),
				container.DataDir("/data"),
				container.WebSocket(),
				container.WebSocketAddress("0.0.0.0"),
				container.WebSocketAPI("admin,eth,net,web3,personal,miner"),
				container.WebSocketOrigin("*"),
				container.NAT("any"),
			)

			Expect(err).To(BeNil())

			for _, n := range nodes {
				err = n.Start()
				Expect(err).To(BeNil())
			}
		})

		AfterEach(func() {
			for _, n := range nodes {
				n.Stop()
			}
		})

		It("TFS-06-01: Node connection", func(done Done) {
			By("Connect all nodes to the fullnodes", func() {
				for _, n := range nodes {
					for _, v := range blockchain.Fullnodes() {
						Expect(n.AddPeer(v.NodeAddress())).To(BeNil())
					}
				}
			})

			By("Wait for p2p connection", func() {
				tests.WaitFor(nodes, func(node container.Ethereum, wg *sync.WaitGroup) {
					Expect(node.WaitForPeersConnected(numberOfFullnodes)).To(BeNil())
					wg.Done()
				})
			})

			close(done)
		}, 15)

		It("TFS-06-02: Node synchronization", func(done Done) {
			const targetBlockHeight = 10

			By("Wait for blocks", func() {
				tests.WaitFor(blockchain.Fullnodes(), func(geth container.Ethereum, wg *sync.WaitGroup) {
					Expect(geth.WaitForBlocks(targetBlockHeight)).To(BeNil())
					wg.Done()
				})
			})

			By("Stop consensus", func() {
				for _, v := range blockchain.Fullnodes() {
					client := v.NewClient()
					Expect(client).NotTo(BeNil())
					err := client.StopMining(context.Background())
					Expect(err).To(BeNil())
					client.Close()
				}
			})

			By("Connect all nodes to the fullnodes", func() {
				for _, n := range nodes {
					for _, v := range blockchain.Fullnodes() {
						Expect(n.AddPeer(v.NodeAddress())).To(BeNil())
					}
				}
			})

			By("Wait for p2p connection", func() {
				tests.WaitFor(nodes, func(node container.Ethereum, wg *sync.WaitGroup) {
					Expect(node.WaitForPeersConnected(numberOfFullnodes)).To(BeNil())
					wg.Done()
				})
			})

			By("Wait for block synchronization between nodes and fullnodes", func() {
				tests.WaitFor(nodes, func(geth container.Ethereum, wg *sync.WaitGroup) {
					Expect(geth.WaitForBlockHeight(targetBlockHeight)).To(BeNil())
					wg.Done()
				})
			})

			By("Check target block hash of nodes", func() {
				expectedBlock, err := blockchain.Fullnodes()[0].NewClient().BlockByNumber(context.Background(), big.NewInt(targetBlockHeight))
				Expect(err).To(BeNil())
				Expect(expectedBlock).NotTo(BeNil())

				for _, n := range nodes {
					nodeClient := n.NewClient()
					block, err := nodeClient.BlockByNumber(context.Background(), big.NewInt(targetBlockHeight))

					Expect(err).To(BeNil())
					Expect(block).NotTo(BeNil())
					Expect(expectedBlock.Hash()).To(BeEquivalentTo(block.Hash()))
				}
			})

			close(done)
		}, 30)
	})
})
