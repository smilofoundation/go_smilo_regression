package functional_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ethereum/go-ethereum/common"

	tests "go-smilo/src/blockchain/regression"
	"go-smilo/src/blockchain/regression/src/container"
	"go-smilo/src/blockchain/regression/src/genesis"
	"go-smilo/src/blockchain/smilobft/core/types"
)

var _ = Describe("QFS-01: General consensus", func() {
	const (
		numberOfFullnodes = 4
	)
	var (
		vaultNetwork container.VaultNetwork
		blockchain   container.Blockchain
		err 		 error
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

	It("QFS-01-01, QFS-01-02: Blockchain initialization and run", func() {
		errc := make(chan error, len(blockchain.Fullnodes()))
		valSet := make(map[common.Address]bool, numberOfFullnodes)
		for _, geth := range blockchain.Fullnodes() {
			valSet[geth.Address()] = true
		}
		for _, geth := range blockchain.Fullnodes() {
			go func(geth container.Ethereum) {
				// 1. Verify genesis block
				c := geth.NewClient()
				if c == nil {
					errc <- errors.New("could not start client")
					return
				}
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

	It("QFS-01-03: Peer connection", func(done Done) {
		expectedPeerCount := len(blockchain.Fullnodes()) - 1
		tests.WaitFor(blockchain.Fullnodes(), func(v container.Ethereum, wg *sync.WaitGroup) {
			Expect(v.WaitForPeersConnected(expectedPeerCount)).To(BeNil())
			wg.Done()
		})

		close(done)
	}, 50)

	It("QFS-01-04: Consensus progress", func(done Done) {
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
							diff := header.Time.Int64() - lastBlockTime
							if diff > maxBlockPeriod {
								errStr := fmt.Sprintf("Invaild block(%v) period, want:%v, got:%v", header.Number.Int64(), maxBlockPeriod, diff)
								errc <- errors.New(errStr)
								return
							}
						}
						lastBlockTime = header.Time.Int64()
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

	It("QFS-01-05: Round robin proposer selection", func(done Done) {
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
							errc <- errors.New("wrong times to be proposer.")
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
