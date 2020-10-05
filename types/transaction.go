package types

import (
	"github.com/kardiachain/go-kardiamain/types"
)

type Transaction struct {
	TxHash           string `json:"hash" bson:"hash"`
	To               string `json:"to" bson:"to"`
	From             string `json:"from" bson:"from"`
	Status           bool   `json:"status" bson:"status"`
	ContractAddress  string `json:"contract_address" bson:"contract_address"`
	Value            string `json:"value" bson:"value"`
	GasPrice         uint64 `json:"gasPrice" bson:"gasPrice"`
	GasFee           uint64 `json:"gas" bson:"gas"`
	GasLimit         uint64 `json:"gasLimit" bson:"gasLimit"`
	BlockNumber      uint64 `json:"blockNumber" bson:"blockNumber"`
	Nonce            uint64 `json:"nonce" bson:"nonce"`
	BlockHash        string `json:"blockHash" bson:"blockHash"`
	Time             uint64 `json:"time" bson:"time"`
	InputData        string `json:"input" bson:"input"`
	Logs             string `json:"logs" bson:"logs"`
	TransactionIndex uint   `json:"transactionIndex"`

	ReceiptReceived bool `json:"-" bson:"receipt_received"`
}

type Log struct {
	Address     string   `json:"address"`
	Topics      []string `json:"topics"`
	Data        string   `json:"data"`
	BlockHeight uint64   `json:"blockHeight"`
	TxHash      string   `json:"transactionHash"`
	TxIndex     uint     `json:"transactionIndex"`
	BlockHash   string   `json:"blockHash"`
	Index       uint     `json:"logIndex"`
	Removed     bool     `json:"removed"`
}

type Receipt struct {
	BlockHash         string      `json:"blockHash"`
	BlockHeight       uint64      `json:"blockHeight"`
	TransactionHash   string      `json:"transactionHash"`
	TransactionIndex  uint64      `json:"transactionIndex"`
	From              string      `json:"from"`
	To                string      `json:"to"`
	GasUsed           uint64      `json:"gasUsed"`
	CumulativeGasUsed uint64      `json:"cumulativeGasUsed"`
	ContractAddress   string      `json:"contractAddress"`
	Logs              []Log       `json:"logs"`
	LogsBloom         types.Bloom `json:"logsBloom"`
	Root              string      `json:"root"`
	Status            uint        `json:"status"`
}

type TransactionList struct {
	Transactions []*Transaction `json:"txs"`
}
