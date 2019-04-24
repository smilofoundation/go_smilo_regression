package container

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/rlp"

	"go-smilo/src/blockchain/smilobft/consensus/sport"
	"go-smilo/src/blockchain/smilobft/core/types"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func sigHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewKeccak256()

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
