package types

import (
	"math/big"
	"time"

	"github.com/kardiachain/go-kardia/types"
)

type Transaction struct {
	BlockHash   string `json:"blockHash" bson:"blockHash"`
	BlockNumber uint64 `json:"blockNumber" bson:"blockNumber"`

	Hash             string        `json:"hash" bson:"hash"`
	From             string        `json:"from" bson:"from"`
	To               string        `json:"to" bson:"to"`
	Status           uint          `json:"status" bson:"status"`
	ContractAddress  string        `json:"contractAddress" bson:"contractAddress"`
	Value            string        `json:"value" bson:"value"`
	GasPrice         uint64        `json:"gasPrice" bson:"gasPrice"`
	GasLimit         uint64        `json:"gas" bson:"gas"`
	GasUsed          uint64        `json:"gasUsed"`
	TxFee            string        `json:"txFee"`
	Nonce            uint64        `json:"nonce" bson:"nonce"`
	Time             time.Time     `json:"time" bson:"time"`
	InputData        string        `json:"input" bson:"input"`
	DecodedInputData *FunctionCall `json:"decodedInputData,omitempty" bson:"decodedInputData"`
	Logs             []Log         `json:"logs" bson:"logs"`
	TransactionIndex uint          `json:"transactionIndex"`
	LogsBloom        types.Bloom   `json:"logsBloom"`
	Root             string        `json:"root"`
}

type FunctionCall struct {
	Function   string                 `json:"function"`
	MethodID   string                 `json:"methodID"`
	MethodName string                 `json:"methodName"`
	Arguments  map[string]interface{} `json:"arguments"`
}

type Log struct {
	ContractAddress string                 `json:"address" bson:"address"`
	MethodName      string                 `json:"methodName" bson:"methodName"`
	ArgumentsName   string                 `json:"argumentsName" bson:"argumentsName"`
	Arguments       map[string]interface{} `json:"arguments" bson:"arguments"`
	Topics          []string               `json:"topics" bson:"topics"`
	Data            string                 `json:"data" bson:"data"`
	BlockHeight     uint64                 `json:"blockHeight" bson:"blockHeight"`
	TxHash          string                 `json:"transactionHash"  bson:"transactionHash"`
	TxIndex         uint                   `json:"transactionIndex" bson:"transactionIndex"`
	BlockHash       string                 `json:"blockHash" bson:"blockHash"`
	Index           uint                   `json:"logIndex" bson:"logIndex"`
	Removed         bool                   `json:"removed" bson:"removed"`
}

type Receipt struct {
	BlockHash   string `json:"blockHash"`
	BlockHeight uint64 `json:"blockHeight"`

	TransactionHash  string `json:"transactionHash"`
	TransactionIndex uint64 `json:"transactionIndex"`

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

type TransactionByAddress struct {
	Address string    `json:"address" bson:"address"`
	TxHash  string    `json:"txHash" bson:"txHash"`
	Time    time.Time `json:"time" bson:"time"`
}

type CallArgsJSON struct {
	From     string   `json:"from"`     // the sender of the 'transaction'
	To       *string  `json:"to"`       // the destination contract (nil for contract creation)
	Gas      uint64   `json:"gas"`      // if 0, the call executes with near-infinite gas
	GasPrice *big.Int `json:"gasPrice"` // HYDRO <-> gas exchange ratio
	Value    *big.Int `json:"value"`    // amount of HYDRO sent along with the call
	Data     string   `json:"data"`     // input data, usually an ABI-encoded contract method invocation
}
