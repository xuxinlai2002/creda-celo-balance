package tokens

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/xuxinlai2002/creda-celo-balance/client"
	"github.com/xuxinlai2002/creda-celo-balance/config"

	"context"
	"fmt"
	"math/big"
)

var rpcClient *client.Client

type LogTransfer struct {
	From   common.Address
	To     common.Address
	Tokens *big.Int
}

func Start(cfg *config.Config) error {
	cli, err := client.Dial(cfg.HTTP)
	if err != nil {
		return err
	}
	rpcClient = cli

	processERC20Tokens(cfg)
	return nil
}

func processERC20Tokens(cfg *config.Config) {
	tokens := ERC20Tokens
	distance := uint64(10000)
	toBlock := uint64(0)
	logTransferSig := []byte("Transfer(address,address,uint256)")

	for i := cfg.StartBlock; i < cfg.EndBlock; i = toBlock + 1 {
		if i+distance < cfg.EndBlock {
			toBlock = i + distance
		} else {
			toBlock = cfg.EndBlock
		}
		fmt.Println("block", i, "toblock", toBlock)
		for address, _ := range tokens {
			query := rpcClient.BuildQuery(address, logTransferSig, big.NewInt(0).SetUint64(i), big.NewInt(0).SetUint64(toBlock))
			logs, err := rpcClient.FilterLogs(context.Background(), query)
			if err != nil {
				fmt.Println("filter logs failed", "error", err)
			}
			if len(logs) > 0 {
				fmt.Println("address", address, "block", i, "toblock", toBlock)
				fmt.Println("logs==", logs)
				for _, vlog := range logs {
					var transferEvent LogTransfer
					transferEvent.Tokens = big.NewInt(0).SetBytes(vlog.Data)
					transferEvent.From = common.HexToAddress(vlog.Topics[1].Hex())
					transferEvent.To = common.HexToAddress(vlog.Topics[2].Hex())

					fmt.Printf("From: %s\n", transferEvent.From.Hex())
					fmt.Printf("To: %s\n", transferEvent.To.Hex())
					fmt.Printf("Tokens: %d\n", transferEvent.Tokens)
				}

			}

		}
	}

}
