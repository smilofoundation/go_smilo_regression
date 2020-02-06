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
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	tests "go-smilo/src/blockchain/regression"
	"go-smilo/src/blockchain/regression/src/container"
)

var _ = Describe("SFS-04: Non-Byzantine Faulty", func() {
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

	It("SFS-04-01: Stop F fullnodes", func(done Done) {
		By("Generating blockchain progress before stopping fullnode", func() {
			tests.WaitFor(blockchain.Fullnodes(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(3)).To(BeNil())
				wg.Done()
			})
		})

		By("Stopping fullnode 0", func() {
			v0 := blockchain.Fullnodes()[0]
			e := v0.Stop()
			Expect(e).To(BeNil())
			ticker := time.NewTicker(time.Millisecond * 100)
			for range ticker.C {
				e := v0.Stop()
				// Wait for e to be non-nil to make sure the container is down
				if e != nil {
					ticker.Stop()
					break
				}
			}
		})

		By("Checking blockchain progress after stopping fullnode", func() {
			tests.WaitFor(blockchain.Fullnodes()[1:], func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlocks(3)).To(BeNil())
				wg.Done()
			})
		})

		close(done)
	}, 120)
})
