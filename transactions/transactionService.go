package transactions

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/common/hexutil"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/celo-org/celo-blockchain/params"
	"github.com/xuxinlai2002/creda-celo-balance/client"
	"github.com/xuxinlai2002/creda-celo-balance/config"
	"github.com/xuxinlai2002/creda-celo-balance/db"
	ctypes "github.com/xuxinlai2002/creda-celo-balance/types"
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
	pullTxList map[string][]*ctypes.TokenRecord
	dataBase   *db.PostgresDB
}

func New(cfg *config.Config) (*BlockPull, error) {
	cli, err := client.Dial(cfg.HTTP)
	if err != nil {
		return nil, err
	}
	database, err := db.NewDB(cfg.PostgresDBName, cfg.PostgresUser, cfg.PostgresPassword)
	if err != nil {
		return nil, err
	}
	pull := &BlockPull{
		client:   cli,
		config:   cfg,
		coinID:   "5567",
		dataBase: database,
	}
	return pull, nil
}

func (p *BlockPull) Start(results chan<- error) {
	go func() {
		err := p.pullBlock()
		p.persistToDB(p.pullTxList)
		results <- err
	}()
}

func (p *BlockPull) getTableNameByTimeStamp(timestamp uint64) string {
	t := time.Unix(int64(timestamp), 0)
	date := fmt.Sprintf("tx_%04d%02d%02d", t.Year(), int(t.Month()), t.Day())
	return date
}
func (p *BlockPull) persistToDB(records map[string][]*ctypes.TokenRecord) {
	//p.dataBase.CreateRecordTable()
	for filename, datas := range records {
		fmt.Println("filename", filename, "datas", datas)
		err := p.dataBase.CreateRecordTable(filename)
		if err != nil {
			fmt.Println("persistToDB CreateRecordTable", "error", err)
			panic(any(err.Error()))
		}
		err = p.dataBase.InsertRecords(filename, datas)
		if err != nil {
			fmt.Println("persistToDB failed", "error", err)
			panic(any(err.Error()))
		}
	}
}

//
//func (p *BlockPull) saveTxToFile(datas map[string][]*tokens.TokenRecord) {
//	for k, v := range datas {
//		if err := p.saveToFile(k, v); err != nil {
//			fmt.Printf("save transaction list to file err: %v\n", err)
//		}
//		delete(datas, k)
//	}
//}
//
//func (p *BlockPull) saveToFile(fileName string, data []*tokens.TokenRecord) error {
//	filePath := p.config.OutputDir + fileName + ".txt"
//	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_SYNC, 0666)
//	if err != nil {
//		return err
//	}
//	defer f.Close()
//	info, err := f.Stat()
//	if err != nil {
//		fmt.Println("Error getting file info:", "fileName", fileName, "error", err)
//		return err
//	}
//	title := "coinID,blockNumber,timestamp, txHash,from,to,value\n"
//	if info.Size() > 0 {
//		title = ""
//	}
//
//	_, err = f.WriteString(title)
//	if err != nil {
//		return err
//	}
//	for _, d := range data {
//		line := fmt.Sprintf("%d,%d,%d,%s,%s,%s,%d\n", d.CoinID, d.BlockNumber, d.Timestamp, d.TxHash, d.From, d.To, d.Value)
//		_, err = f.WriteString(line)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}

func (p *BlockPull) pullBlock() error {
	p.pullTxList = make(map[string][]*ctypes.TokenRecord)
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
		filePath := p.getTableNameByTimeStamp(b.Time())
		if _, ok := p.pullTxList[filePath]; !ok {
			if i > startHeight {
				p.persistToDB(p.pullTxList)
				p.pullTxList = make(map[string][]*ctypes.TokenRecord)
			}
		}
		signer := types.MakeSigner(params.MainnetChainConfig, b.Number())
		fmt.Println("getBlock", b.NumberU64())
		for _, tx := range b.Transactions() {
			fmt.Println("trace tx", tx.Hash().String())
			if tx.Value().Cmp(big.NewInt(0)) > 0 {
				from, errMsg := types.Sender(signer, tx)
				if errMsg == nil {
					coinID, ok := big.NewInt(0).SetString(p.coinID, 10)
					if !ok {
						fmt.Println("CoinID is not correct", "coinID", p.coinID)
					}
					tr := &ctypes.TokenRecord{
						CoinID:      coinID.Uint64(),
						BlockNumber: b.NumberU64(),
						Timestamp:   b.Header().Time,
						TxHash:      tx.Hash(),
						From:        from,
						To:          *tx.To(),
						Value:       tx.Value(),
					}
					p.addPullTxRecord(filePath, tr)
				}
			}

			info, err := p.client.TraceTx(ctx, tx.Hash().String())
			if err != nil {
				return err
			}
			p.processInteralTxsInfo(info, tx.Hash(), b.NumberU64(), b.Time(), filePath)
		}

		if b.NumberU64()%1000 == 0 {
			utils.WriteCurrentHeight(p.config.OutputDir, b.NumberU64())
		}
	}
	return nil
}

func (p *BlockPull) addPullTxRecord(filePath string, tr *ctypes.TokenRecord) {
	if _, ok := p.pullTxList[filePath]; ok {
		p.pullTxList[filePath] = append(p.pullTxList[filePath], tr)
	} else {
		p.pullTxList[filePath] = []*ctypes.TokenRecord{tr}
	}
}

func (p *BlockPull) processInteralTxsInfo(txInfo map[string]interface{}, txID common.Hash, blockHeight, timestamp uint64, filePath string) {
	var tx = &InternalTx{
		From: txInfo["from"].(string),
		To:   txInfo["to"].(string),
		Type: txInfo["type"].(string),
	}
	if v, ok := txInfo["value"]; ok {
		tx.Value, _ = hexutil.DecodeUint64(v.(string))
	}
	if tx.Value != 0 && tx.Type == "CALL" {
		coinID, ok := big.NewInt(0).SetString(p.coinID, 10)
		if !ok {
			fmt.Println("CoinID is not correct", "coinID", p.coinID)
		}
		tr := &ctypes.TokenRecord{
			CoinID:      coinID.Uint64(),
			BlockNumber: blockHeight,
			Timestamp:   timestamp,
			TxHash:      txID,
			From:        common.HexToAddress(tx.From),
			To:          common.HexToAddress(tx.To),
			Value:       big.NewInt(0).SetUint64(tx.Value),
		}
		p.addPullTxRecord(filePath, tr)
	}

	if calls, ok := txInfo["calls"]; ok {
		var items = calls.([]interface{})
		for i := 0; i < len(items); i++ {
			p.processInteralTxsInfo(items[i].(map[string]interface{}), txID, blockHeight, timestamp, filePath)
		}
	}
}
