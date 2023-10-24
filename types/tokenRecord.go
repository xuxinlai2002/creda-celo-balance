package types

import (
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
)

const CELO_COINID = 5567

type TokenRecord struct {
	CoinID      uint64
	BlockNumber uint64
	Timestamp   uint64
	TxHash      common.Hash
	From        common.Address
	To          common.Address
	Value       *big.Int
}

type COINID uint64
type ADDRESS string
type DATE string
