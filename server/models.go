package server

import "time"

type PagingResponse struct {
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
	Total uint64      `json:"total"`
	Data  interface{} `json:"data"`
}

type Blocks []SimpleBlock

type SimpleBlock struct {
	Height          uint64    `json:"height,omitempty" bson:"height"`
	Time            time.Time `json:"time,omitempty" bson:"time"`
	ProposerAddress string    `json:"proposerAddress,omitempty" bson:"proposerAddress"`
	NumTxs          uint64    `json:"numTxs" bson:"numTxs"`
	GasLimit        uint64    `json:"gasLimit,omitempty" bson:"gasLimit"`
	GasUsed         uint64    `json:"gasUsed" bson:"gasUsed"`
	Rewards         string    `json:"rewards" bson:"rewards"`
}

type Transactions []SimpleTransaction

type SimpleTransaction struct {
	Hash        string    `json:"hash" bson:"hash"`
	BlockNumber uint64    `json:"blockNumber" bson:"blockNumber"`
	Time        time.Time `json:"time" bson:"time"`
	From        string    `json:"from" bson:"from"`
	To          string    `json:"to" bson:"to"`
	Value       string    `json:"value" bson:"value"`
	TxFee       string    `json:"txFee"`
	Status      uint      `json:"status" bson:"status"`
}
