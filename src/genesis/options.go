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
			Seal:          make([]byte, types.BFTExtraSeal),
			CommittedSeal: [][]byte{},
		}

		payload, err := rlp.EncodeToBytes(&ist)
		if err != nil {
			return
		}

		extraData := "0x" + common.Bytes2Hex(append(newVanity, payload...))
		genesis.ExtraData = hexutil.MustDecode(extraData)
		genesis.GasLimit = InitGasLimit
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
