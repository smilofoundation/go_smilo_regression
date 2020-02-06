// Copyright 2020 smilofoundation/regression Authors
// Copyright 2019 smilofoundation/regression Authors
// Copyright 2017 AMIS Technologies
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package functional_test

import (
	"context"
	"math/big"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	tests "go-smilo/src/blockchain/regression"
	"go-smilo/src/blockchain/regression/src/container"
)

var _ = Describe("Block synchronization testing", func() {
	const (
		numberOfFullnodes = 4
	)
	var (
		vaultNetwork container.VaultNetwork
		blockchain   container.Blockchain
	)

	BeforeEach(func() {
		vaultNetwork = container.NewDefaultVaultNetwork(dockerNetwork, numberOfFullnodes)
		Expect(vaultNetwork.Start()).To(BeNil())
		blockchain = container.NewDefaultSmiloBlockchain(dockerNetwork, vaultNetwork)
		Expect(blockchain.Start(true)).To(BeNil())
	})

	AfterEach(func() {
		blockchain.Stop(true)
		blockchain.Finalize()
		vaultNetwork.Stop()
		vaultNetwork.Finalize()
	})

	Describe("SFS-06: Block synchronization testing", func() {
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

		It("SFS-06-01: Node connection", func(done Done) {
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
		}, 50)

		It("SFS-06-02: Node synchronization", func(done Done) {
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
		}, 50)
	})
})
