package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/celo-org/celo-blockchain/crypto"
	"github.com/celo-org/celo-blockchain/eth/tracers"
	"math/big"

	"github.com/celo-org/celo-blockchain"
	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/common/hexutil"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/celo-org/celo-blockchain/rpc"
)

type Client struct {
	rpcClient *rpc.Client
}

func Dial(rawurl string) (*Client, error) {
	rpcClient, err := rpc.DialContext(context.TODO(), rawurl)
	if err != nil {
		return nil, err
	}
	return &Client{
		rpcClient: rpcClient,
	}, nil
}

func (c *Client) ChainID(ctx context.Context) (*big.Int, error) {
	var result hexutil.Big
	err := c.rpcClient.CallContext(ctx, &result, "eth_chainId")
	if err != nil {
		return nil, err
	}
	return (*big.Int)(&result), err
}

type headerNumber struct {
	Number *big.Int `json:"number"           gencodec:"required"`
}

func (h *headerNumber) UnmarshalJSON(input []byte) error {
	type headerNumber struct {
		Number *hexutil.Big `json:"number" gencodec:"required"`
	}
	var dec headerNumber
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Number == nil {
		return errors.New("missing required field 'number' for Header")
	}
	h.Number = (*big.Int)(dec.Number)
	return nil
}

func (c *Client) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return c.getBlock(ctx, "eth_getBlockByNumber", toBlockNumArg(number), true)
}

type InternalTx struct {
	From  string `json:"from,omitempty"`
	To    string `json:"to,omitempty"`
	Value uint64 `json:"value,omitempty"`
	Calls string `json:"calls,omitempty"`

	GasUsed uint64 `json:"gasUsed,omitempty"`
	Output  string `json:"output,omitempty"`
	Input   string `json:"input,omitempty"`
	Type    string `json:"type,omitempty"`
	Gas     uint64 `json:"gas,omitempty"`
}

func (c *Client) TraceTx(ctx context.Context, txHash string) (map[string]interface{}, error) {
	var result map[string]interface{} = make(map[string]interface{}, 0)
	tracerStr := "callTracer"
	err := c.rpcClient.CallContext(ctx, &result, "debug_traceTransaction", txHash, tracers.TraceConfig{Tracer: &tracerStr})
	if err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) LatestBlock() (*big.Int, error) {
	var head *headerNumber

	err := c.rpcClient.CallContext(context.Background(), &head, "eth_getBlockByNumber", "latest", false)
	if err == nil && head == nil {
		err = errors.New("not found")
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return head.Number, err
}

// TransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (c *Client) TransactionReceipt(txHash common.Hash) (*types.Receipt, error) {
	var r *types.Receipt
	err := c.rpcClient.CallContext(context.Background(), &r, "eth_getTransactionReceipt", txHash)
	if err == nil {
		if r == nil {
			return nil, celo.NotFound
		}
	}
	return r, err
}

func (c *Client) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	var result hexutil.Big
	err := c.rpcClient.CallContext(ctx, &result, "eth_getBalance", account, toBlockNumArg(blockNumber))
	return (*big.Int)(&result), err
}

func (c *Client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	var hex hexutil.Big
	if err := c.rpcClient.CallContext(ctx, &hex, "eth_gasPrice"); err != nil {
		return nil, err
	}
	return (*big.Int)(&hex), nil
}

func (c *Client) CallContract(ctx context.Context, callArgs map[string]interface{}, blockNumber *big.Int) ([]byte, error) {
	var hex hexutil.Bytes
	err := c.rpcClient.CallContext(ctx, &hex, "eth_call", callArgs, toBlockNumArg(blockNumber))
	if err != nil {
		return nil, err
	}
	return hex, nil
}

func (c *Client) BuildQuery(contractAddress string, sig []byte, startBlock *big.Int, endBlock *big.Int) celo.FilterQuery {
	query := celo.FilterQuery{
		FromBlock: startBlock,
		ToBlock:   endBlock,
		Addresses: []common.Address{common.HexToAddress(contractAddress)},
		Topics: [][]common.Hash{
			{crypto.Keccak256Hash(sig)},
		},
	}
	return query
}

// FilterLogs executes a filter query.
func (c *Client) FilterLogs(ctx context.Context, q celo.FilterQuery) ([]types.Log, error) {
	var result []types.Log
	arg, err := toFilterArg(q)
	if err != nil {
		return nil, err
	}
	err = c.rpcClient.CallContext(ctx, &result, "eth_getLogs", arg)
	return result, err
}

func toFilterArg(q celo.FilterQuery) (interface{}, error) {
	arg := map[string]interface{}{
		"address": q.Addresses,
		"topics":  q.Topics,
	}
	if q.BlockHash != nil {
		arg["blockHash"] = *q.BlockHash
		if q.FromBlock != nil || q.ToBlock != nil {
			return nil, fmt.Errorf("cannot specify both BlockHash and FromBlock/ToBlock")
		}
	} else {
		if q.FromBlock == nil {
			arg["fromBlock"] = "0x0"
		} else {
			arg["fromBlock"] = toBlockNumArg(q.FromBlock)
		}
		arg["toBlock"] = toBlockNumArg(q.ToBlock)
	}
	return arg, nil
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	return hexutil.EncodeBig(number)
}

type rpcBlock struct {
	Hash           common.Hash           `json:"hash"`
	Transactions   []rpcTransaction      `json:"transactions"`
	Randomness     *types.Randomness     `json:"randomness"`
	EpochSnarkData *types.EpochSnarkData `json:"epochSnarkData"`
}

type rpcTransaction struct {
	tx *types.Transaction
	txExtraInfo
}

type txExtraInfo struct {
	BlockNumber *string         `json:"blockNumber,omitempty"`
	BlockHash   *common.Hash    `json:"blockHash,omitempty"`
	From        *common.Address `json:"from,omitempty"`
}

func (tx *rpcTransaction) UnmarshalJSON(msg []byte) error {
	if err := json.Unmarshal(msg, &tx.tx); err != nil {
		return err
	}
	return json.Unmarshal(msg, &tx.txExtraInfo)
}

func (c *Client) getBlock(ctx context.Context, method string, args ...interface{}) (*types.Block, error) {
	var raw json.RawMessage
	err := c.rpcClient.CallContext(ctx, &raw, method, args...)
	if err != nil {
		return nil, err
	} else if len(raw) == 0 {
		return nil, errors.New("not found")
	}
	// Decode header and transactions.
	var head *types.Header
	var body rpcBlock
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	if head.TxHash == types.EmptyRootHash && len(body.Transactions) > 0 {
		return nil, fmt.Errorf("server returned non-empty transaction list but block header indicates no transactions")
	}
	if head.TxHash != types.EmptyRootHash && len(body.Transactions) == 0 {
		return nil, fmt.Errorf("server returned empty transaction list but block header indicates transactions")
	}
	// Fill the sender cache of transactions in the block.
	txs := make([]*types.Transaction, len(body.Transactions))
	for i, tx := range body.Transactions {
		if tx.From != nil {
			setSenderFromServer(tx.tx, *tx.From, body.Hash)
		}
		txs[i] = tx.tx
	}
	return types.NewBlockWithHeader(head).WithBody(txs, body.Randomness, body.EpochSnarkData), nil
}

// senderFromServer is a types.Signer that remembers the sender address returned by the RPC
// server. It is stored in the transaction's sender address cache to avoid an additional
// request in TransactionSender.
type senderFromServer struct {
	addr      common.Address
	blockhash common.Hash
}

var errNotCached = errors.New("sender not cached")

func setSenderFromServer(tx *types.Transaction, addr common.Address, block common.Hash) {
	// Use types.Sender for side-effect to store our signer into the cache.
	types.Sender(&senderFromServer{addr, block}, tx)
}

func (s *senderFromServer) Equal(other types.Signer) bool {
	os, ok := other.(*senderFromServer)
	return ok && os.blockhash == s.blockhash
}

func (s *senderFromServer) Sender(tx *types.Transaction) (common.Address, error) {
	if s.blockhash == (common.Hash{}) {
		return common.Address{}, errNotCached
	}
	return s.addr, nil
}

func (s *senderFromServer) ChainID() *big.Int {
	panic("can't sign with senderFromServer")
}
func (s *senderFromServer) Hash(tx *types.Transaction) common.Hash {
	panic("can't sign with senderFromServer")
}
func (s *senderFromServer) SignatureValues(tx *types.Transaction, sig []byte) (R, S, V *big.Int, err error) {
	panic("can't sign with senderFromServer")
}
