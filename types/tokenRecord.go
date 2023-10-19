package types

import (
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
)

type TokenRecord struct {
	CoinID      uint64
	BlockNumber uint64
	Timestamp   uint64
	TxHash      common.Hash
	From        common.Address
	To          common.Address
	Value       *big.Int
}
