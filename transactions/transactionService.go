package transactions

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/xuxinlai2002/creda-celo-balance/client"
	"github.com/xuxinlai2002/creda-celo-balance/config"

	"github.com/celo-org/celo-blockchain/common/hexutil"
)

var celoClientRpc *client.Client

var outputFile *os.File

func Start(cfg *config.Config, results chan<- error) {
	cli, err := client.Dial(cfg.HTTP)
	if err != nil {
		fmt.Println(err)
		results <- err
	}
	celoClientRpc = cli

	outputFullPath := cfg.OutputDir + "/output" + ".txt"
	outputFile, err = os.OpenFile(outputFullPath, os.O_WRONLY|os.O_CREATE, 0644)

	outputFile.Write([]byte("from,to,value\n"))

	if err != nil {
		fmt.Println("can not open the file:", err)
		return
	}

	go func() {
		err = pullBlock(cfg)
		if err != nil {
			fmt.Println("pull block failed", "error", err)
			results <- err
		}
		defer outputFile.Close()
	}()

}

func pullBlock(cfg *config.Config) error {
	ctx := context.Background()
	for i := cfg.PullStartHeight; i <= cfg.PullEndHeight; i++ {
		b, err := celoClientRpc.BlockByNumber(ctx, big.NewInt(0).SetUint64(i))
		if err != nil {
			return err
		}

		info, err := celoClientRpc.TraceTx(ctx, b.Transactions()[7].Hash().String())
		if err != nil {
			return err
		}

		recursionInternalTx(info)
	}
	return nil
}

func interfaceToInternalTx(items []interface{}) {
	for i := 0; i < len(items); i++ {
		item := items[i].(map[string]interface{})
		var tx = &client.InternalTx{
			From: item["from"].(string),
			To:   item["to"].(string),
			Type: item["type"].(string),
		}
		if v, ok := item["value"]; ok {
			tx.Value, _ = hexutil.DecodeUint64(v.(string))
		}

		if item["calls"] != nil {
			recursionInternalTx(item)
		}

		fmt.Print(tx.From)
		line := fmt.Sprintf("%s,%s,%d\n", tx.From, tx.To, tx.Value)

		_, err := outputFile.Write([]byte(line))
		if err != nil {
			fmt.Println("write file error:", err)
			return
		}
	}
}

func recursionInternalTx(txs map[string]interface{}) error {
	tx := txs["calls"].([]interface{})
	interfaceToInternalTx(tx)
	return nil
}
