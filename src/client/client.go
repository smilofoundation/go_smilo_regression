package client

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"

	ethereum "go-smilo/src/blockchain/smilobft"
	"go-smilo/src/blockchain/smilobft/ethclient"
	"go-smilo/src/blockchain/smilobft/p2p"
)

type client struct {
	c         *rpc.Client
	ethClient *ethclient.Client
}

func Dial(rawurl string) (Client, error) {
	ic, err := rpc.Dial(rawurl)
	if err != nil {
		return nil, err
	}
	return &client{
		c:         ic,
		ethClient: ethclient.NewClient(ic),
	}, nil
}

func (ic *client) Close() {
	ic.c.Close()
}

// ----------------------------------------------------------------------------

func (ic *client) AddPeer(ctx context.Context, nodeURL string) error {
	var r bool
	// TODO: Result needs to be verified
	// The response data type are bytes, but we cannot parse...
	err := ic.c.CallContext(ctx, &r, "admin_addPeer", nodeURL)
	if err != nil {
		return err
	}
	return err
}

func (ic *client) AdminPeers(ctx context.Context) ([]*p2p.PeerInfo, error) {
	var r []*p2p.PeerInfo
	// The response data type are bytes, but we cannot parse...
	err := ic.c.CallContext(ctx, &r, "admin_peers")
	if err != nil {
		return nil, err
	}
	return r, err
}

func (ic *client) NodeInfo(ctx context.Context) (*p2p.PeerInfo, error) {
	var r *p2p.PeerInfo
	err := ic.c.CallContext(ctx, &r, "admin_nodeInfo")
	if err != nil {
		return nil, err
	}
	return r, err
}

// ----------------------------------------------------------------------------
func (ic *client) BlockNumber(ctx context.Context) (*big.Int, error) {
	var r string
	err := ic.c.CallContext(ctx, &r, "eth_blockNumber")
	if err != nil {
		return nil, err
	}
	h, err := hexutil.DecodeBig(r)
	return h, err
}

// ----------------------------------------------------------------------------

func (ic *client) StartMining(ctx context.Context) error {
	var r []byte
	// TODO: Result needs to be verified
	// The response data type are bytes, but we cannot parse...
	err := ic.c.CallContext(ctx, &r, "miner_start", nil)
	if err != nil {
		return err
	}
	return err
}

func (ic *client) StopMining(ctx context.Context) error {
	err := ic.c.CallContext(ctx, nil, "miner_stop", nil)
	if err != nil {
		return err
	}
	return err
}

// ----------------------------------------------------------------------------

func (ic *client) SendTransaction(ctx context.Context, from, to common.Address, value *big.Int) (txHash string, err error) {
	var hex hexutil.Bytes
	arg := map[string]interface{}{
		"from":  from,
		"to":    to,
		"value": (*hexutil.Big)(value),
	}
	if err = ic.c.CallContext(ctx, &hex, "eth_sendTransaction", arg); err == nil {
		txHash = hex.String()
	}
	return
}

func (ic *client) CreateContract(ctx context.Context, from common.Address, bytecode string, gas *big.Int) (txHash string, err error) {
	var hex hexutil.Bytes
	arg := map[string]interface{}{
		"from":     from,
		"gas":      (*hexutil.Big)(gas),
		"data":     bytecode,
		"gasPrice": (*hexutil.Big)(big.NewInt(0)),
	}
	if err = ic.c.CallContext(ctx, &hex, "eth_sendTransaction", arg); err == nil {
		txHash = hex.String()
	}
	return
}

func (ic *client) CreatePrivateContract(ctx context.Context, from common.Address, bytecode string, gas *big.Int, privateFor []string) (txHash string, err error) {
	var hex hexutil.Bytes
	arg := map[string]interface{}{
		"from":       from,
		"gas":        (*hexutil.Big)(gas),
		"data":       bytecode,
		"privateFor": privateFor,
		"gasPrice":   (*hexutil.Big)(big.NewInt(0)),
	}
	if err = ic.c.CallContext(ctx, &hex, "eth_sendTransaction", arg); err == nil {
		txHash = hex.String()
	}
	return
}

// ----------------------------------------------------------------------------

func (ic *client) ProposeFullnode(ctx context.Context, address common.Address, auth bool) error {
	var r []byte
	// TODO: Result needs to be verified with other method
	// The response data type are bytes, but we cannot parse...
	err := ic.c.CallContext(ctx, &r, "smilobft_propose", address, auth)
	if err != nil {
		fmt.Println("ERROR: test_client, Could not smilobft_propose", err)
		return ethereum.NotFound
	}
	return err
}

type addresses []common.Address

func (addrs addresses) Len() int {
	return len(addrs)
}

func (addrs addresses) Less(i, j int) bool {
	return strings.Compare(addrs[i].String(), addrs[j].String()) < 0
}

func (addrs addresses) Swap(i, j int) {
	addrs[i], addrs[j] = addrs[j], addrs[i]
}

func (ic *client) GetFullnodes(ctx context.Context, blockNumber *big.Int) ([]common.Address, error) {
	var r []common.Address
	err := ic.c.CallContext(ctx, &r, "smilobft_getFullnodes", toNumArg(blockNumber))
	if err == nil && r == nil {
		fmt.Println("ERROR: test_client, Could not smilobft_getFullnodes", err)
		return nil, ethereum.NotFound
	}

	sort.Sort(addresses(r))

	return r, err
}

func toNumArg(number *big.Int) string {
	if number == nil {
		return ""
	}
	return hexutil.EncodeBig(number)
}
