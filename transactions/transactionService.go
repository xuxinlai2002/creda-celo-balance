package transactions

import (
	"fmt"
	client "github.com/xuxinlai2002/creda-celo-balance/celoClient"
	"github.com/xuxinlai2002/creda-celo-balance/config"
)

var celoClientRpc *client.Client

func Start(cfg *config.Config) error {
	cli, err := client.Dial(cfg.HTTP)
	if err != nil {
		fmt.Println(err)
		return err
	}
	celoClientRpc = cli

	b, err := celoClientRpc.GetBlockByNumber(10)
	fmt.Println("block", b, "error", err)
	return nil
}
