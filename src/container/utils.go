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

package container

import (
	"golang.org/x/crypto/sha3"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/rlp"

	"go-smilo/src/blockchain/smilobft/consensus/sport"
	"go-smilo/src/blockchain/smilobft/core/types"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func sigHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()

	// Clean seal is required for calculating proposer seal.
	rlp.Encode(hasher, types.SportFilteredHeader(header, false))
	hasher.Sum(hash[:0])
	return hash
}

func GetSpeaker(header *types.Header) common.Address {
	if header == nil {
		return common.Address{}
	}

	extra, err := hexutil.Decode(common.ToHex(header.Extra))
	if err != nil {
		log.Error("Decode, hexutil.Decode error %v", err)
		return common.Address{}
	}

	sportExtra, err := types.ExtractSportExtra(&types.Header{Extra: extra})
	if err != nil {
		log.Error("Decode, ExtractSportExtra error %v", err)
		return common.Address{}
	}

	addr, err := sport.GetSignatureAddress(sigHash(header).Bytes(), sportExtra.Seal)
	if err != nil {
		return common.Address{}
	}
	return addr
}
