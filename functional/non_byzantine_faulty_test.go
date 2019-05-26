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
		blockchain, err := container.NewDefaultBlockchain(dockerNetwork, numberOfFullnodes)
		Expect(err).To(BeNil())
		Expect(blockchain).ToNot(BeNil())
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
