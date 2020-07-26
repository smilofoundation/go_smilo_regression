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

package client

import (
	"context"
	"go-smilo/src/blockchain/smilobft/accounts/abi/bind"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	ethereum "go-smilo/src/blockchain/smilobft"
	"go-smilo/src/blockchain/smilobft/core/types"
)

// Blockchain Access

// BlockByHash returns the given full block.
//
// Note that loading full blocks requires two requests. Use HeaderByHash
// if you don't need all transactions or uncle headers.
func (ic *client) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return ic.ethClient.BlockByHash(ctx, hash)
}

// BlockByNumber returns a block from the current canonical chain. If number is nil, the
// latest known block is returned.
//
// Note that loading full blocks requires two requests. Use HeaderByNumber
// if you don't need all transactions or uncle headers.
func (ic *client) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return ic.ethClient.BlockByNumber(ctx, number)
}

// HeaderByHash returns the block header with the given hash.
func (ic *client) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return ic.ethClient.HeaderByHash(ctx, hash)
}

// HeaderByNumber returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (ic *client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return ic.ethClient.HeaderByNumber(ctx, number)
}

// TransactionByHash returns the transaction with the given hash.
func (ic *client) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	return ic.ethClient.TransactionByHash(ctx, hash)
}

// TransactionCount returns the total number of transactions in the given block.
func (ic *client) TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error) {
	return ic.ethClient.TransactionCount(ctx, blockHash)
}

// TransactionInBlock returns a single transaction at index in the given block.
func (ic *client) TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*types.Transaction, error) {
	return ic.ethClient.TransactionInBlock(ctx, blockHash, index)
}

// TransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (ic *client) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return ic.ethClient.TransactionReceipt(ctx, txHash)
}

// SyncProgress retrieves the current progress of the sync algorithm. If there's
// no sync currently running, it returns nil.
func (ic *client) SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
	return ic.ethClient.SyncProgress(ctx)
}

// SubscribeNewHead subscribes to notifications about the current blockchain head
// on the given channel.
func (ic *client) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	return ic.ethClient.SubscribeNewHead(ctx, ch)
}

// State Access

// NetworkID returns the network ID (also known as the chain ID) for this chain.
func (ic *client) NetworkID(ctx context.Context) (*big.Int, error) {
	return ic.ethClient.NetworkID(ctx)
}

// BalanceAt returns the wei balance of the given account.
// The block number can be nil, in which case the balance is taken from the latest known block.
func (ic *client) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	return ic.ethClient.BalanceAt(ctx, account, blockNumber)
}

// StorageAt returns the value of key in the contract storage of the given account.
// The block number can be nil, in which case the value is taken from the latest known block.
func (ic *client) StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	return ic.ethClient.StorageAt(ctx, account, key, blockNumber)
}

// CodeAt returns the contract code of the given account.
// The block number can be nil, in which case the code is taken from the latest known block.
func (ic *client) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	return ic.ethClient.CodeAt(ctx, account, blockNumber)
}

// NonceAt returns the account nonce of the given account.
// The block number can be nil, in which case the nonce is taken from the latest known block.
func (ic *client) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	return ic.ethClient.NonceAt(ctx, account, blockNumber)
}

// Filters

// FilterLogs executes a filter query.
func (ic *client) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	return ic.ethClient.FilterLogs(ctx, q)
}

// SubscribeFilterLogs subscribes to the results of a streaming filter query.
func (ic *client) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	return ic.ethClient.SubscribeFilterLogs(ctx, q, ch)
}

// Pending State

// PendingBalanceAt returns the wei balance of the given account in the pending state.
func (ic *client) PendingBalanceAt(ctx context.Context, account common.Address) (*big.Int, error) {
	return ic.ethClient.PendingBalanceAt(ctx, account)
}

// PendingStorageAt returns the value of key in the contract storage of the given account in the pending state.
func (ic *client) PendingStorageAt(ctx context.Context, account common.Address, key common.Hash) ([]byte, error) {
	return ic.ethClient.PendingStorageAt(ctx, account, key)
}

// PendingCodeAt returns the contract code of the given account in the pending state.
func (ic *client) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	return ic.ethClient.PendingCodeAt(ctx, account)
}

// PendingNonceAt returns the account nonce of the given account in the pending state.
// This is the nonce that should be used for the next transaction.
func (ic *client) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return ic.ethClient.PendingNonceAt(ctx, account)
}

// PendingTransactionCount returns the total number of transactions in the pending state.
func (ic *client) PendingTransactionCount(ctx context.Context) (uint, error) {
	return ic.ethClient.PendingTransactionCount(ctx)
}

// Contract Calling

// CallContract executes a message call transaction, which is directly executed in the VM
// of the node, but never mined into the blockchain.
//
// blockNumber selects the block height at which the call runs. It can be nil, in which
// case the code is taken from the latest known block. Note that state from very old
// blocks might not be available.
func (ic *client) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return ic.ethClient.CallContract(ctx, msg, blockNumber)
}

// PendingCallContract executes a message call transaction using the EVM.
// The state seen by the contract call is the pending state.
func (ic *client) PendingCallContract(ctx context.Context, msg ethereum.CallMsg) ([]byte, error) {
	return ic.ethClient.PendingCallContract(ctx, msg)
}

// SuggestGasPrice retrieves the currently suggested gas price to allow a timely
// execution of a transaction.
func (ic *client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return ic.ethClient.SuggestGasPrice(ctx)
}

// EstimateGas tries to estimate the gas needed to execute a specific transaction based on
// the current pending state of the backend blockchain. There is no guarantee that this is
// the true gas limit requirement as other transactions may be added or removed by miners,
// but it should provide a basis for setting a reasonable default.
func (ic *client) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	return ic.ethClient.EstimateGas(ctx, msg)
}

// SendRawTransaction injects a signed transaction into the pending pool for execution.
//
// If the transaction was a contract creation use the TransactionReceipt method to get the
// contract address after the transaction has been mined.
func (ic *client) SendRawTransaction(ctx context.Context, tx *types.Transaction) error {
	return ic.ethClient.SendTransaction(ctx, tx, bind.PrivateTxArgs{})
}
