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
	"github.com/ethereum/go-ethereum/common/math"

	"go-smilo/src/blockchain/smilobft/core"
	"go-smilo/src/blockchain/smilobft/params"
)

//go:generate gencodec -type SmiloGenesis -field-override genesisSpecMarshaling -out gen_smilo_genesis.go

// field type overrides for gencodec
type genesisSpecMarshaling struct {
	Nonce      math.HexOrDecimal64
	Timestamp  math.HexOrDecimal64
	ExtraData  hexutil.Bytes
	GasLimit   math.HexOrDecimal64
	GasUsed    math.HexOrDecimal64
	Number     math.HexOrDecimal64
	Difficulty *math.HexOrDecimal256
	Alloc      map[common.UnprefixedAddress]core.GenesisAccount
}

type SmiloChainConfig struct {
	*params.ChainConfig
	IsSmilo       bool `json:"isSmilo,omitempty"`
	IsGas         bool `json:"isGas,omitempty"`
	IsGasRefunded bool `json:"isGasRefunded,omitempty"`
}

type SmiloGenesis struct {
	Config     *SmiloChainConfig `json:"config"`
	Nonce      uint64            `json:"nonce"`
	Timestamp  uint64            `json:"timestamp"`
	ExtraData  []byte            `json:"extraData"`
	GasLimit   uint64            `json:"gasLimit"   gencodec:"required"`
	Difficulty *big.Int          `json:"difficulty" gencodec:"required"`
	Mixhash    common.Hash       `json:"mixHash"`
	Coinbase   common.Address    `json:"coinbase"`
	Alloc      core.GenesisAlloc `json:"alloc"      gencodec:"required"`

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Number     uint64      `json:"number"`
	GasUsed    uint64      `json:"gasUsed"`
	ParentHash common.Hash `json:"parentHash"`
}

// ToSmilo converts standard genesis to smilo genesis
func ToSmilo(g *core.Genesis, isSmilo bool) *SmiloGenesis {
	return &SmiloGenesis{
		Config: &SmiloChainConfig{
			ChainConfig:   g.Config,
			IsSmilo:       isSmilo,
			IsGas:         isSmilo,
			IsGasRefunded: isSmilo,
		},
		Nonce:      g.Nonce,
		Timestamp:  g.Timestamp,
		ExtraData:  g.ExtraData,
		GasLimit:   g.GasLimit,
		Difficulty: g.Difficulty,
		Mixhash:    g.Mixhash,
		Coinbase:   g.Coinbase,
		Alloc:      g.Alloc,
		//Number:     g.Number,
		//GasUsed:    g.GasUsed,
		ParentHash: g.ParentHash,
	}
}
