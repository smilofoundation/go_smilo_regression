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

func TestNonByzantine(t *testing.T) {
}

var _ = Describe("TFS-04: Non-Byzantine Faulty", func() {
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

	It("TFS-04-01: Stop F fullnodes", func(done Done) {
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
