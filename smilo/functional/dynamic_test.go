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
	"math"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	tests "go-smilo/src/blockchain/regression"
	"go-smilo/src/blockchain/regression/src/container"
)

var _ = Describe("SFS-02: Dynamic fullnodes addition/removal testing", func() {
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
		Expect(blockchain.Start(true)).To(BeNil())
	})

	AfterEach(func() {
		blockchain.Stop(true)
		blockchain.Finalize()
		vaultNetwork.Stop()
		vaultNetwork.Finalize()
	})

	It("SFS-02-01: Add fullnodes", func() {
		testFullnodes := 1

		By("Ensure the number of fullnodes is correct", func() {
			for _, v := range blockchain.Fullnodes() {
				client := v.NewClient()
				n, err := client.BlockNumber(context.Background())
				Expect(err).Should(BeNil())
				fullnodes, err := client.GetFullnodes(context.Background(), n)
				Expect(err).Should(BeNil())
				Expect(len(fullnodes)).Should(BeNumerically("==", numberOfFullnodes))
			}
		})

		By("Add fullnodes", func() {
			_, err := blockchain.AddFullnodes(testFullnodes)
			Expect(err).Should(BeNil())
		})

		By("Wait for several blocks", func() {
			tests.WaitFor(blockchain.Fullnodes(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(10)).To(BeNil())
				wg.Done()
			})
		})

		By("Ensure the number of fullnodes is correct", func() {
			for _, v := range blockchain.Fullnodes() {
				client := v.NewClient()
				n, err := client.BlockNumber(context.Background())
				Expect(err).Should(BeNil())
				fullnodes, err := client.GetFullnodes(context.Background(), n)
				Expect(err).Should(BeNil())
				Expect(len(fullnodes)).Should(BeNumerically("==", numberOfFullnodes+testFullnodes))
			}
		})
	})

	//It("SFS-02-02: New fullnodes consensus participation", func() {
	//	testFullnode := 1
	//
	//	newFullnodes, err := blockchain.AddFullnodes(testFullnode)
	//	Expect(err).Should(BeNil())
	//
	//	tests.WaitFor(blockchain.Fullnodes()[numberOfFullnodes:], func(eth container.Ethereum, wg *sync.WaitGroup) {
	//		Expect(eth.WaitForProposed(newFullnodes[0].Address(), 250*time.Second)).Should(BeNil())
	//		wg.Done()
	//	})
	//})

	It("SFS-02-03: Remove fullnodes", func() {
		numOfCandidates := 3

		By("Ensure that numbers of fullnode is equal to $numberOfFullnodes", func() {
			for _, v := range blockchain.Fullnodes() {
				client := v.NewClient()
				n, err := client.BlockNumber(context.Background())
				Expect(err).Should(BeNil())
				fullnodes, err := client.GetFullnodes(context.Background(), n)
				Expect(err).Should(BeNil())
				Expect(len(fullnodes)).Should(BeNumerically("==", numberOfFullnodes))
			}
		})

		By("Add fullnodes", func() {
			_, err := blockchain.AddFullnodes(numOfCandidates)
			Expect(err).Should(BeNil())
		})

		By("Ensure that consensus is working in 50 seconds", func() {
			Expect(blockchain.EnsureConsensusWorking(blockchain.Fullnodes(), 50*time.Second)).Should(BeNil())
		})

		By("Check if the number of fullnodes is correct", func() {
			for _, v := range blockchain.Fullnodes() {
				client := v.NewClient()
				n, err := client.BlockNumber(context.Background())
				Expect(err).Should(BeNil())
				fullnodes, err := client.GetFullnodes(context.Background(), n)
				Expect(err).Should(BeNil())
				Expect(len(fullnodes)).Should(BeNumerically("==", numberOfFullnodes+numOfCandidates))
			}
		})

		// remove fullnodes [1,2,3]
		By("Remove fullnodes", func() {
			removalCandidates := blockchain.Fullnodes()[:numOfCandidates]
			processingTime := time.Duration(math.Pow(2, float64(len(removalCandidates)))*7) * time.Second
			Expect(blockchain.RemoveFullnodes(removalCandidates, processingTime)).Should(BeNil())
		})

		By("Ensure that consensus is working in 20 seconds", func() {
			Expect(blockchain.EnsureConsensusWorking(blockchain.Fullnodes(), 20*time.Second)).Should(BeNil())
		})

		By("Check if the number of fullnodes is correct", func() {
			for _, v := range blockchain.Fullnodes() {
				client := v.NewClient()
				n, err := client.BlockNumber(context.Background())
				Expect(err).Should(BeNil())
				fullnodes, err := client.GetFullnodes(context.Background(), n)
				Expect(err).Should(BeNil())
				Expect(len(fullnodes)).Should(BeNumerically("==", numberOfFullnodes))
			}
		})

		By("Ensure that consensus is working in 30 seconds", func() {
			Expect(blockchain.EnsureConsensusWorking(blockchain.Fullnodes(), 30*time.Second)).Should(BeNil())
		})
	})

	It("SFS-02-04: Reduce fullnode network size below 2F+1", func() {
		By("Ensure that blocks are generated by fullnodes", func() {
			tests.WaitFor(blockchain.Fullnodes(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(5)).To(BeNil())
				wg.Done()
			})
		})

		By("Reduce fullnode network size but keep it more than 2F+1", func() {
			// stop fullnodes [3]
			stopCandidates := blockchain.Fullnodes()[numberOfFullnodes-1:]
			for _, candidates := range stopCandidates {
				c := candidates.NewClient()
				Expect(c.StopMining(context.Background())).Should(BeNil())
			}
		})

		By("Verify number of fullnodes", func() {
			for _, v := range blockchain.Fullnodes() {
				client := v.NewClient()
				n, err := client.BlockNumber(context.Background())
				Expect(err).Should(BeNil())
				fullnodes, err := client.GetFullnodes(context.Background(), n)
				Expect(err).Should(BeNil())
				Expect(len(fullnodes)).Should(BeNumerically("==", numberOfFullnodes))
			}
		})

		By("Ensure that blocks are generated by fullnodes", func() {
			tests.WaitFor(blockchain.Fullnodes()[:numberOfFullnodes-1], func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(5)).To(BeNil())
				wg.Done()
			})
		})
	})

	It("SFS-02-05: Reduce fullnode network size below 2F+1", func() {
		By("Ensure that blocks are generated by fullnodes", func() {
			tests.WaitFor(blockchain.Fullnodes(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(5)).To(BeNil())
				wg.Done()
			})
		})

		By("Reduce fullnode network size to less than 2F+1", func() {
			stopCandidates := blockchain.Fullnodes()[numberOfFullnodes-2:]
			// stop fullnodes [3,4]
			for _, candidates := range stopCandidates {
				c := candidates.NewClient()
				Expect(c.StopMining(context.Background())).Should(BeNil())
			}
		})

		By("Verify number of fullnodes", func() {
			for _, v := range blockchain.Fullnodes() {
				client := v.NewClient()
				n, err := client.BlockNumber(context.Background())
				Expect(err).Should(BeNil())
				fullnodes, err := client.GetFullnodes(context.Background(), n)
				Expect(err).Should(BeNil())
				Expect(len(fullnodes)).Should(BeNumerically("==", numberOfFullnodes))
			}
		})

		By("No block generated", func() {
			// REMARK: ErrNoBlock will return if fullnodes not generate block after 10 second.
			Expect(blockchain.EnsureConsensusWorking(blockchain.Fullnodes(), 11*time.Second)).Should(Equal(container.ErrNoBlock))
		})
	})
})
