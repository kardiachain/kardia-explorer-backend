package types

import (
	"time"

	"github.com/kardiachain/go-kardia/types"
)

type Log struct {
	Address       string                 `json:"address,omitempty" bson:"address"`
	MethodName    string                 `json:"methodName,omitempty" bson:"methodName"`
	ArgumentsName string                 `json:"argumentsName,omitempty" bson:"argumentsName"`
	Arguments     map[string]interface{} `json:"arguments,omitempty" bson:"arguments"`
	Topics        []string               `json:"topics,omitempty" bson:"topics"`
	Data          string                 `json:"data,omitempty" bson:"data"`
	BlockHeight   uint64                 `json:"blockHeight,omitempty" bson:"blockHeight"`
	Time          time.Time              `json:"time" bson:"time"`
	TxHash        string                 `json:"transactionHash"  bson:"transactionHash"`
	TxIndex       uint                   `json:"transactionIndex,omitempty" bson:"transactionIndex"`
	BlockHash     string                 `json:"blockHash,omitempty" bson:"blockHash"`
	Index         uint                   `json:"logIndex,omitempty" bson:"logIndex"`
	Removed       bool                   `json:"removed,omitempty" bson:"removed"`
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
