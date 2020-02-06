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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sync"

	tests "go-smilo/src/blockchain/regression"
	"go-smilo/src/blockchain/regression/src/container"
)

var _ = Describe("SFS-03: Recoverability testing", func() {
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

	It("SFS-03-01: Add fullnodes in a network with < 2F+1 fullnodes to > 2F+1", func(done Done) {
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

		//By("The consensus should not work after resuming", func() {
		//	tests.WaitFor(blockchain.Fullnodes(), func(geth container.Ethereum, wg *sync.WaitGroup) {
		//		// container.ErrNoBlock should be returned if we didn't see any block in 10 seconds
		//		Expect(geth.WaitForBlocks(1, 30*time.Second)).To(BeEquivalentTo(container.ErrNoBlock))
		//		wg.Done()
		//	})
		//})

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
