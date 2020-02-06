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
		blockchain, err := container.NewDefaultBlockchain(dockerNetwork, numberOfFullnodes)
		Expect(err).To(BeNil())
		Expect(blockchain).ToNot(BeNil())
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
