package types

import (
	"github.com/kardiachain/go-kardia/types"
)

type Log struct {
	Address       string                 `json:"address" bson:"address"`
	MethodName    string                 `json:"methodName" bson:"methodName"`
	ArgumentsName string                 `json:"argumentsName" bson:"argumentsName"`
	Arguments     map[string]interface{} `json:"arguments" bson:"arguments"`
	Topics        []string               `json:"topics" bson:"topics"`
	Data          string                 `json:"data" bson:"data"`
	BlockHeight   uint64                 `json:"blockHeight" bson:"blockHeight"`
	TxHash        string                 `json:"transactionHash"  bson:"transactionHash"`
	TxIndex       uint                   `json:"transactionIndex" bson:"transactionIndex"`
	BlockHash     string                 `json:"blockHash" bson:"blockHash"`
	Index         uint                   `json:"logIndex" bson:"logIndex"`
	Removed       bool                   `json:"removed" bson:"removed"`
}

type Receipt struct {
	TransactionHash  string `json:"transactionHash" bson:"transactionHash"`
	TransactionIndex uint64 `json:"transactionIndex" bson:"transactionIndex"`

	BlockHash         string      `json:"blockHash" bson:"blockHash"`
	BlockHeight       uint64      `json:"blockHeight" bson:"blockHeight"`
	From              string      `json:"from" bson:"from"`
	To                string      `json:"to" bson:"to"`
	GasUsed           uint64      `json:"gasUsed" bson:"gasUsed"`
	CumulativeGasUsed uint64      `json:"cumulativeGasUsed" bson:"cumulativeGasUsed"`
	ContractAddress   string      `json:"contractAddress" bson:"contractAddress"`
	Logs              []Log       `json:"logs" bson:"logs"`
	LogsBloom         types.Bloom `json:"logsBloom"`
	Root              string      `json:"root" bson:"root"`
	Status            uint        `json:"status" bson:"status"`
}
