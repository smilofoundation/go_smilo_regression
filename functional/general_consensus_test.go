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
	"errors"
	"fmt"
	"math/big"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"github.com/ethereum/go-ethereum/common"

	tests "go-smilo/src/blockchain/regression"
	"go-smilo/src/blockchain/regression/src/container"
	"go-smilo/src/blockchain/regression/src/genesis"
	"go-smilo/src/blockchain/smilobft/core/types"
)

func TestConsensus(t *testing.T) {
}

var _ = Describe("TFS-01: General consensus", func() {
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

	It("TFS-01-01, TFS-01-02: Blockchain initialization and run", func() {
		errc := make(chan error, len(blockchain.Fullnodes()))
		valSet := make(map[common.Address]bool, numberOfFullnodes)
		for _, geth := range blockchain.Fullnodes() {
			valSet[geth.Address()] = true
		}
		for _, geth := range blockchain.Fullnodes() {
			go func(geth container.Ethereum) {
				// 1. Verify genesis block
				c := geth.NewClient()
				header, err := c.HeaderByNumber(context.Background(), big.NewInt(0))
				if err != nil {
					errc <- err
					return
				}

				if header.GasLimit != genesis.InitGasLimit {
					errStr := fmt.Sprintf("Invalid genesis gas limit. want:%v, got:%v", genesis.InitGasLimit, header.GasLimit)
					errc <- errors.New(errStr)
					return
				}

				if header.Difficulty.Int64() != genesis.InitDifficulty {
					errStr := fmt.Sprintf("Invalid genesis difficulty. want:%v, got:%v", genesis.InitDifficulty, header.Difficulty.Int64())
					errc <- errors.New(errStr)
					return
				}

				if header.MixDigest != types.SportDigest {
					errStr := fmt.Sprintf("Invalid block mixhash. want:%v, got:%v", types.SportDigest, header.MixDigest)
					errc <- errors.New(errStr)
					return

				}

				// 2. Check fullnode set
				istClient := geth.NewClient()
				n, err := istClient.BlockNumber(context.Background())
				Expect(err).Should(BeNil())

				vals, err := istClient.GetFullnodes(context.Background(), n)
				if err != nil {
					errc <- err
					return
				}

				for _, val := range vals {
					if _, ok := valSet[val]; !ok {
						errc <- errors.New("Invalid fullnode address.")
						return
					}
				}

				errc <- nil
			}(geth)
		}

		for i := 0; i < len(blockchain.Fullnodes()); i++ {
			err := <-errc
			Expect(err).To(BeNil())
		}

	})

	It("TFS-01-03: Peer connection", func(done Done) {
		expectedPeerCount := len(blockchain.Fullnodes()) - 1
		tests.WaitFor(blockchain.Fullnodes(), func(v container.Ethereum, wg *sync.WaitGroup) {
			Expect(v.WaitForPeersConnected(expectedPeerCount)).To(BeNil())
			wg.Done()
		})

		close(done)
	}, 50)

	It("TFS-01-04: Consensus progress", func(done Done) {
		const (
			targetBlockHeight = 10
			maxBlockPeriod    = 3
		)

		By("Wait for consensus progress", func() {
			tests.WaitFor(blockchain.Fullnodes(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlockHeight(targetBlockHeight)).To(BeNil())
				wg.Done()
			})
		})

		By("Check the block period should less than 3 seconds", func() {
			errc := make(chan error, len(blockchain.Fullnodes()))
			for _, geth := range blockchain.Fullnodes() {
				go func(geth container.Ethereum) {
					c := geth.NewClient()
					lastBlockTime := int64(0)
					// The reason to verify block period from block#2 is that
					// the block period from block#1 to block#2 might take long time due to
					// encounter several round changes at the beginning of the consensus progress.
					for i := 2; i <= targetBlockHeight; i++ {
						header, err := c.HeaderByNumber(context.Background(), big.NewInt(int64(i)))
						if err != nil {
							errc <- err
							return
						}
						if lastBlockTime != 0 {
							diff := int64(header.Time) - lastBlockTime
							if diff > maxBlockPeriod {
								errStr := fmt.Sprintf("Invaild block(%v) period, want:%v, got:%v", header.Number.Int64(), maxBlockPeriod, diff)
								errc <- errors.New(errStr)
								return
							}
						}
						lastBlockTime = int64(header.Time)
					}
					errc <- nil
				}(geth)
			}

			for i := 0; i < len(blockchain.Fullnodes()); i++ {
				err := <-errc
				Expect(err).To(BeNil())
			}
		})
		close(done)
	}, 60)

	It("TFS-01-05: Round robin proposer selection", func(done Done) {
		var (
			timesOfBeSpeaker  = 3
			targetBlockHeight = timesOfBeSpeaker * numberOfFullnodes
			emptySpeaker      = common.Address{}
		)

		By("Wait for consensus progress", func() {
			tests.WaitFor(blockchain.Fullnodes(), func(geth container.Ethereum, wg *sync.WaitGroup) {
				Expect(geth.WaitForBlockHeight(targetBlockHeight)).To(BeNil())
				wg.Done()
			})
		})

		By("Block proposer selection should follow round-robin policy", func() {
			errc := make(chan error, len(blockchain.Fullnodes()))
			for _, geth := range blockchain.Fullnodes() {
				go func(geth container.Ethereum) {
					c := geth.NewClient()
					istClient := geth.NewClient()

					// get initial fullnode set
					n, err := istClient.BlockNumber(context.Background())
					Expect(err).Should(BeNil())
					vals, err := istClient.GetFullnodes(context.Background(), n)
					if err != nil {
						errc <- err
						return
					}

					lastSpeakerIdx := -1
					counts := make(map[common.Address]int, numberOfFullnodes)
					// initial count map
					for _, addr := range vals {
						counts[addr] = 0
					}
					for i := 1; i <= targetBlockHeight; i++ {
						header, err := c.HeaderByNumber(context.Background(), big.NewInt(int64(i)))
						if err != nil {
							errc <- err
							return
						}

						p := container.GetSpeaker(header)
						if p == emptySpeaker {
							errStr := fmt.Sprintf("Empty block(%v) proposer", header.Number.Int64())
							errc <- errors.New(errStr)
							return
						}
						// count the times to be the proposer
						if count, ok := counts[p]; ok {
							counts[p] = count + 1
						}
						// check if the proposer is valid
						if lastSpeakerIdx == -1 {
							for i, val := range vals {
								if p == val {
									lastSpeakerIdx = i
									break
								}
							}
						} else {
							proposerIdx := (lastSpeakerIdx + 1) % len(vals)
							if p != vals[proposerIdx] {
								errStr := fmt.Sprintf("Invaild block(%v) proposer, want:%v, got:%v", header.Number.Int64(), vals[proposerIdx], p)
								errc <- errors.New(errStr)
								return
							}
							lastSpeakerIdx = proposerIdx
						}
					}
					// check times to be proposer
					for _, count := range counts {
						if count != timesOfBeSpeaker {
							errc <- errors.New("Wrong times to be proposer.")
							return
						}
					}
					errc <- nil
				}(geth)
			}

			for i := 0; i < len(blockchain.Fullnodes()); i++ {
				err := <-errc
				Expect(err).To(BeNil())
			}
		})
		close(done)
	}, 120)
})
