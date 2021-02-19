package server

import (
	"time"

	coreTypes "github.com/kardiachain/go-kardia/types"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

type PagingResponse struct {
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
	Total uint64      `json:"total"`
	Data  interface{} `json:"data"`
}

type Blocks []SimpleBlock

type SimpleBlock struct {
	Height          uint64    `json:"height,omitempty"`
	Hash            string    `json:"hash,omitempty"`
	Time            time.Time `json:"time,omitempty"`
	ProposerAddress string    `json:"proposerAddress,omitempty"`
	ProposerName    string    `json:"proposerName"`
	NumTxs          uint64    `json:"numTxs"`
	GasLimit        uint64    `json:"gasLimit,omitempty"`
	GasUsed         uint64    `json:"gasUsed"`
	Rewards         string    `json:"rewards"`
}

type Block struct {
	types.Block
	ProposerName string `json:"proposerName"`
}

type Transactions []SimpleTransaction

type SimpleTransaction struct {
	Hash               string              `json:"hash"`
	BlockNumber        uint64              `json:"blockNumber"`
	Time               time.Time           `json:"time"`
	From               string              `json:"from"`
	FromName           string              `json:"fromName"`
	To                 string              `json:"to"`
	ToName             string              `json:"toName"`
	IsInValidatorsList bool                `json:"isInValidatorsList"`
	Role               int                 `json:"role"`
	ContractAddress    string              `json:"contractAddress,omitempty"`
	Value              string              `json:"value"`
	TxFee              string              `json:"txFee"`
	Status             uint                `json:"status"`
	DecodedInputData   *types.FunctionCall `json:"decodedInputData,omitempty"`
	InputData          string              `json:"input"`
}

type Transaction struct {
	BlockHash   string `json:"blockHash"`
	BlockNumber uint64 `json:"blockNumber"`

	Hash               string              `json:"hash"`
	From               string              `json:"from"`
	To                 string              `json:"to"`
	ToName             string              `json:"toName"`
	IsInValidatorsList bool                `json:"isInValidatorsList"`
	Role               int                 `json:"role"`
	Status             uint                `json:"status"`
	ContractAddress    string              `json:"contractAddress"`
	Value              string              `json:"value"`
	GasPrice           uint64              `json:"gasPrice"`
	GasLimit           uint64              `json:"gas"`
	GasUsed            uint64              `json:"gasUsed"`
	TxFee              string              `json:"txFee"`
	Nonce              uint64              `json:"nonce"`
	Time               time.Time           `json:"time"`
	InputData          string              `json:"input"`
	DecodedInputData   *types.FunctionCall `json:"decodedInputData,omitempty"`
	Logs               []types.Log         `json:"logs"`
	TransactionIndex   uint                `json:"transactionIndex"`
	LogsBloom          coreTypes.Bloom     `json:"logsBloom"`
	Root               string              `json:"root"`
}

type NodeInfo struct {
	ID         string `json:"id"`
	Moniker    string `json:"moniker"`
	PeersCount int    `json:"peersCount"`
}

type Addresses []SimpleAddress

type SimpleAddress struct {
	Address            string `json:"address"` // low precise balance for sorting purposes
	BalanceString      string `json:"balance"` // high precise balance for API
	IsContract         bool   `json:"isContract"`
	Name               string `json:"name"`
	IsInValidatorsList bool   `json:"isInValidatorsList"`
	Role               int    `json:"role"`
	Rank               uint64 `json:"rank"`
}

type valInfoResponse struct {
	Name string
	Role int
}
