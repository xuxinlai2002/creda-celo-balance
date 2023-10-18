package transactions

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strconv"

	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/celo-org/celo-blockchain/params"
	"github.com/xuxinlai2002/creda-celo-balance/client"
	"github.com/xuxinlai2002/creda-celo-balance/config"

	"github.com/celo-org/celo-blockchain/common/hexutil"
	"github.com/xuxinlai2002/creda-celo-balance/utils"
)

type InternalTx struct {
	From  string
	To    string
	Value uint64
	Calls string
	Type  string
}

type BlockPull struct {
	client     *client.Client
	config     *config.Config
	coinID     string
	outputFile *os.File
}

func New(cfg *config.Config) (*BlockPull, error) {
	cli, err := client.Dial(cfg.HTTP)
	if err != nil {
		return nil, err
	}
	outputFullPath := cfg.OutputDir + "/output" + ".txt"
	outputFile, err := os.OpenFile(outputFullPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	info, err := outputFile.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return nil, err
	}
	title := "coinID,blockNumber,txHash,timestamp,from,to,value\n"
	if info.Size() > 0 {
		title = ""
	}
	outputFile.Write([]byte(title))
	pull := &BlockPull{
		client:     cli,
		config:     cfg,
		coinID:     "5567",
		outputFile: outputFile,
	}
	return pull, nil
}

func (p *BlockPull) Start(results chan<- error) {
	go func() {
		err := p.pullBlock()
		p.outputFile.Close()
		results <- err
	}()
}

func (p *BlockPull) pullBlock() error {
	startHeight := p.config.PullStartHeight
	progress, err := utils.GetCurrentHeight(p.config.OutputDir)
	if err == nil && progress > startHeight {
		startHeight = progress + 1
	}
	endHeight := p.config.PullEndHeight
	ctx := context.Background()
	for i := startHeight; i <= endHeight; i++ {
		b, err := p.client.BlockByNumber(ctx, big.NewInt(0).SetUint64(i))
		if err != nil {
			return err
		}

		signer := types.MakeSigner(params.MainnetChainConfig, b.Number())

		fmt.Println("getBlock", b.NumberU64())
		for _, tx := range b.Transactions() {
			fmt.Println("trace tx", tx.Hash().String())
			p.writeTxInfo(tx.Hash().String() + "\n")
			if tx.Value().Cmp(big.NewInt(0)) > 0 {
				from, errMsg := types.Sender(signer, tx)
				if errMsg == nil {
					content := p.coinID + "," +
						strconv.FormatUint(b.NumberU64(), 10) + "," +
						tx.Hash().String() + "," +
						strconv.FormatUint(b.Time(), 10) + "," +
						from.String() + "," +
						tx.To().String() + "," +
						tx.Value().String() + "\n"
					p.writeTxInfo(content)
				}
			}

			info, err := p.client.TraceTx(ctx, tx.Hash().String())
			if err != nil {
				return err
			}
			p.processInteralTxsInfo(info, tx.Hash().String(), b.NumberU64(), b.Time())
		}

		if b.NumberU64()%1 == 0 {
			utils.WriteCurrentHeight(p.config.OutputDir, b.NumberU64())
		}
	}
	return nil
}

func (p *BlockPull) writeTxInfo(content string) error {
	_, err := p.outputFile.Write([]byte(content))
	return err
}

func (p *BlockPull) processInteralTxsInfo(txInfo map[string]interface{}, txID string, blockHeight, timestamp uint64) {
	var tx = &InternalTx{
		From: txInfo["from"].(string),
		To:   txInfo["to"].(string),
		Type: txInfo["type"].(string),
	}
	if v, ok := txInfo["value"]; ok {
		tx.Value, _ = hexutil.DecodeUint64(v.(string))
	}
	if tx.Value != 0 && tx.Type == "CALL" {
		content := p.coinID + "," +
			strconv.FormatUint(blockHeight, 10) + "," +
			txID + "," +
			strconv.FormatUint(timestamp, 10) + "," +
			tx.From + "," +
			tx.To + "," +
			strconv.FormatUint(tx.Value, 10) + "\n"

		err := p.writeTxInfo(content)
		if err != nil {
			fmt.Println("write file error:", err)
			return
		}
	}

	if calls, ok := txInfo["calls"]; ok {
		var items = calls.([]interface{})
		for i := 0; i < len(items); i++ {
			p.processInteralTxsInfo(items[i].(map[string]interface{}), txID, blockHeight, timestamp)
		}
	}
}
