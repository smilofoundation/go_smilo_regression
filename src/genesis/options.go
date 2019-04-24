package genesis

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"go-smilo/src/blockchain/smilobft/core"

	"bytes"

	"github.com/ethereum/go-ethereum/rlp"

	"go-smilo/src/blockchain/smilobft/core/types"
)

type Option func(*core.Genesis)

func Fullnodes(addrs ...common.Address) Option {
	return func(genesis *core.Genesis) {

		newVanity, err := hexutil.Decode("0x00")
		if err != nil {
			return
		}

		if len(newVanity) < types.SportExtraVanity {
			newVanity = append(newVanity, bytes.Repeat([]byte{0x00}, types.SportExtraVanity-len(newVanity))...)
		}
		newVanity = newVanity[:types.SportExtraVanity]

		ist := &types.SportExtra{
			Fullnodes:     addrs,
			Seal:          make([]byte, types.SportExtraSeal),
			CommittedSeal: [][]byte{},
		}

		payload, err := rlp.EncodeToBytes(&ist)
		if err != nil {
			return
		}

		extraData := "0x" + common.Bytes2Hex(append(newVanity, payload...))
		if err != nil {
			log.Error("Failed to encode extra data", "err", err)
			return
		}
		genesis.ExtraData = hexutil.MustDecode(extraData)
	}
}

func GasLimit(limit uint64) Option {
	return func(genesis *core.Genesis) {
		genesis.GasLimit = limit
	}
}

func Alloc(addrs []common.Address, balance *big.Int) Option {
	return func(genesis *core.Genesis) {
		alloc := make(map[common.Address]core.GenesisAccount)
		for _, addr := range addrs {
			alloc[addr] = core.GenesisAccount{Balance: balance}
		}
		genesis.Alloc = alloc
	}
}
