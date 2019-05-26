package functional_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ethereum/go-ethereum/common"

	"go-smilo/src/blockchain/regression/src/client"
	"go-smilo/src/blockchain/regression/src/container"
	"go-smilo/src/blockchain/smilobft/core/types"
)

const (
	testBaseByteCode = "0x6060604052341561000f57600080fd5b604051602080610149833981016040528080519060200190919050505b806000819055505b505b610104806100456000396000f30060606040526000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680632a1afcd914605157806360fe47b11460775780636d4ce63c146097575b600080fd5b3415605b57600080fd5b606160bd565b6040518082815260200191505060405180910390f35b3415608157600080fd5b6095600480803590602001909190505060c3565b005b341560a157600080fd5b60a760ce565b6040518082815260200191505060405180910390f35b60005481565b806000819055505b50565b6000805490505b905600a165627a7a72305820278d34fe369cdf9e578c4d5cdbbdffa7e964a8e34060e68788e08e52c20181c10029"
)

var _ = Describe("QFS-08: Private transaction", func() {
	const (
		numberOfFullnodes = 4
	)
	var (
		vaultNetwork container.VaultNetwork
		blockchain   container.Blockchain
		err			 error
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
//		blockchain.Finalize()
		vaultNetwork.Stop()
//		vaultNetwork.Finalize()
	})

	It("QFS–08–01: Vault connection", func() {
		//Skip for now, vault network will be ensured in BeforeEach.
	})

	It("QFS–08–02: Sending common contract transaction", func() {
		const storedValue = 1
		var txHash common.Hash

		By("Sending tx should be successful", func() {
			geth0 := blockchain.Fullnodes()[0]
			acc := geth0.Accounts()[0]
			client0 := geth0.NewClient()
			byteCode := genByteCodeWithValue(storedValue)
			hash, err := client0.CreateContract(context.Background(), acc, byteCode, big.NewInt(0x47b760))
			Expect(err).To(BeNil())
			txHash = common.HexToHash(hash)
		})

		By("All geth nodes should can get value of contract storage", func() {
			errc := make(chan error, len(blockchain.Fullnodes()))
			for _, geth := range blockchain.Fullnodes() {
				go func(geth container.Ethereum) {
					ethClient := geth.NewClient()
					errc <- checkContractValue(ethClient, txHash, storedValue)
				}(geth)
			}
			for i := 0; i < len(blockchain.Fullnodes()); i++ {
				err := <-errc
				Expect(err).To(BeNil())
			}
		})
	})

	It("QFS–08–03: Sending private contract transaction", func() {
		var (
			txHash      common.Hash
			storedValue = 1
		)

		By("Sending tx by geth#0 and private for geth#1 should be successful", func() {
			geth0 := blockchain.Fullnodes()[0]
			acc := geth0.Accounts()[0]
			client0 := geth0.NewClient()
			ct1 := vaultNetwork.GetVault(1)
			pubKey1 := ct1.PublicKeys()
			byteCode := genByteCodeWithValue(storedValue)
			hash, err := client0.CreatePrivateContract(context.Background(), acc, byteCode, big.NewInt(0x47b760), pubKey1)
			Expect(err).To(BeNil())
			txHash = common.HexToHash(hash)
		})

		By("Only geth#0 and geth#1 can get value of contract storage", func() {
			errc := make(chan error, len(blockchain.Fullnodes()))
			for i, geth := range blockchain.Fullnodes() {
				var expValue = 0
				//TODO: get nodes by pub and make sure the target node has/not the info before checkContractValue
				if i == 0 || i == 1 {
					expValue = storedValue
				}
				go func(geth container.Ethereum, expValue int) {
					ethClient := geth.NewClient()
					errc <- checkContractValue(ethClient, txHash, expValue)
				}(geth, expValue)
			}
			for i := 0; i < len(blockchain.Fullnodes()); i++ {
				err := <-errc
				Expect(err).To(BeNil())
			}
		})

		storedValue = 2
		By("Sending tx by geth#2 and private for geth#3 should be successful", func() {
			geth2 := blockchain.Fullnodes()[2]
			acc := geth2.Accounts()[0]
			client2 := geth2.NewClient()
			ct3 := vaultNetwork.GetVault(3)
			pubKey3 := ct3.PublicKeys()
			byteCode := genByteCodeWithValue(storedValue)
			hash, err := client2.CreatePrivateContract(context.Background(), acc, byteCode, big.NewInt(0x47b760), pubKey3)
			Expect(err).To(BeNil())
			txHash = common.HexToHash(hash)
		})

		By("Only geth#2 and geth#3 can get value of contract storage", func() {
			errc := make(chan error, len(blockchain.Fullnodes()))
			for i, geth := range blockchain.Fullnodes() {
				var expValue = 0
				if i == 2 || i == 3 {
					expValue = storedValue
				}
				go func(geth container.Ethereum, expValue int) {
					ethClient := geth.NewClient()
					errc <- checkContractValue(ethClient, txHash, expValue)
				}(geth, expValue)
			}
			for i := 0; i < len(blockchain.Fullnodes()); i++ {
				err := <-errc
				Expect(err).To(BeNil())
			}
		})

		storedValue = 3
		By("Sending common tx after private tx should be successful", func() {
			geth0 := blockchain.Fullnodes()[0]
			acc := geth0.Accounts()[0]
			client0 := geth0.NewClient()
			byteCode := genByteCodeWithValue(storedValue)
			hash, err := client0.CreateContract(context.Background(), acc, byteCode, big.NewInt(0x47b760))
			Expect(err).To(BeNil())
			txHash = common.HexToHash(hash)
		})

		By("All geth nodes should can get value of contract storage", func() {
			errc := make(chan error, len(blockchain.Fullnodes()))
			for _, geth := range blockchain.Fullnodes() {
				go func(geth container.Ethereum) {
					ethClient := geth.NewClient()
					errc <- checkContractValue(ethClient, txHash, storedValue)
				}(geth)
			}
			for i := 0; i < len(blockchain.Fullnodes()); i++ {
				err := <-errc
				Expect(err).To(BeNil())
			}
		})
	})
})

func genByteCodeWithValue(v int) string {
	return fmt.Sprintf("%s%064x", testBaseByteCode, v)
}

func getTxReceipt(ethClient client.Client, txHash common.Hash, timeout time.Duration) (*types.Receipt, error) {
	timer := time.After(timeout)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timer:
			return nil, errors.New("Get TxReceipt timeout.")
		case <-ticker.C:
			r, err := ethClient.TransactionReceipt(context.Background(), txHash)
			if err == nil {
				return r, nil
			}
		}
	}
}

func checkContractValue(ethClient client.Client, txHash common.Hash, expValue int) error {
	receipt, err := getTxReceipt(ethClient, txHash, 10*time.Second)
	if err != nil {
		return err
	}

	emptyAddress := common.Address{}
	if receipt.ContractAddress == emptyAddress {
		return errors.New("invalid contract address")
	}

	v, err := ethClient.StorageAt(context.Background(),
		receipt.ContractAddress,
		common.HexToHash("0x0"),
		nil)
	if err != nil {
		return err
	}

	if value := new(big.Int).SetBytes(v).Int64(); value != int64(expValue) {
		errMsg := fmt.Sprintf("Wrong value of contract storage, got:%v, expected:%v", value, expValue)
		return errors.New(errMsg)
	}
	return nil
}
