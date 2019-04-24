package genesis

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"time"

	"go-smilo/src/blockchain/smilobft/consensus/sport"
	"go-smilo/src/blockchain/smilobft/core"
	"go-smilo/src/blockchain/smilobft/core/types"
	"go-smilo/src/blockchain/smilobft/params"

	"go-smilo/src/blockchain/regression/src/common"
)

const (
	FileName       = "genesis.json"
	InitGasLimit   = 4700000
	InitDifficulty = 1
)

func New(options ...Option) *core.Genesis {
	genesis := &core.Genesis{
		Timestamp:  uint64(time.Now().Unix()),
		GasLimit:   InitGasLimit,
		Difficulty: big.NewInt(InitDifficulty),
		Alloc:      make(core.GenesisAlloc),
		Config: &params.ChainConfig{
			ChainID:        big.NewInt(2017),
			HomesteadBlock: big.NewInt(1),
			EIP150Block:    big.NewInt(2),
			EIP155Block:    big.NewInt(3),
			EIP158Block:    big.NewInt(3),
			Sport: &params.SportConfig{
				SpeakerPolicy: uint64(sport.DefaultConfig.SpeakerPolicy),
				Epoch:         sport.DefaultConfig.Epoch,
			},
		},
		Mixhash: types.SportDigest,
	}

	for _, opt := range options {
		opt(genesis)
	}

	return genesis
}

func NewFileAt(dir string, isSmilo bool, options ...Option) string {
	genesis := New(options...)
	if err := Save(dir, genesis, isSmilo); err != nil {
		log.Error("Failed to save genesis", "dir", dir, "err", err)
		return ""
	}

	return filepath.Join(dir, FileName)
}

func NewFile(isSmilo bool, options ...Option) string {
	dir, _ := common.GenerateRandomDir()
	return NewFileAt(dir, isSmilo, options...)
}

func Save(dataDir string, genesis *core.Genesis, isSmilo bool) error {
	filePath := filepath.Join(dataDir, FileName)

	var raw []byte
	var err error
	if isSmilo {
		raw, err = json.Marshal(ToSmilo(genesis, true))
	} else {
		raw, err = json.Marshal(genesis)
	}
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, raw, 0600)
}
