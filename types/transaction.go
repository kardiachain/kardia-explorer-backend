package types

import (
	"math/big"
	"time"

	"github.com/kardiachain/go-kardia/types"
)

const (
	TransactionStatusFailed  = 0
	TransactionStatusSuccess = 1
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

type TxTraceResult struct {
	UsedGas      uint64      `json:"usedGas"`
	Err          interface{} `json:"err"`
	ReturnData   string      `json:"returnData"`
	RevertReason string      `json:"revertReason"`
}
