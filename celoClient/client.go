package client

import (
	"context"
	"errors"
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

func (c *Client) GetBlockByNumber(height uint64) (*types.Block, error) {
	var block *types.Block

	err := c.rpcClient.CallContext(context.Background(), block, "eth_getBlockByNumber", height, false)
	if err == nil || block == nil {
		err = errors.New("not found")
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return block, err
}
