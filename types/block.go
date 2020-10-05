package types

import (
	kai "github.com/kardiachain/go-kardiamain/mainchain"
	"github.com/kardiachain/go-kardiamain/types"
)

type Block struct {
	// basic block info
	BlockHash string `json:"hash" bson:"hash"`
	Height    uint64 `json:"height" bson:"height"`
	Time      uint64 `json:"time" bson:"time"`
	NumTxs    int    `json:"numTxs" bson:"numTxs"`

	NumDualEvents uint64 `json:"numDualEvents" bson:"numDualEvents"`

	GasLimit uint64 `json:"gasLimit" bson:"gasLimit"`
	GasUsed  uint64 `json:"gasUsed" bson:"gasUsed"`

	// prev block info
	LastBlock string `json:"lastBlock" bson:"lastBlock"`

	CommitHash string `json:"commitHash" bson:"commitHash"`
	TxHash     string `json:"dataHash" bson:"dataHash"` // transactions

	DualEventsHash string      `json:"dualEventsHash" bson:"dualEventsHash"`
	Root           string      `json:"stateRoot"  bson:"stateRoot"`
	ReceiptHash    string      `json:"receiptsRoot"     bson:"receiptsRoot"`
	Bloom          types.Bloom `json:"logsBloom"    bson:"logsBloom"`

	Validator string `json:"validator" bson:"validator"`
	// hashes from the app output from the prev block
	ValidatorHash string              `json:"validatorHash" bson:"validatorHash"`
	ConsensusHash string              `json:"consensusHash" bson:"consensusHash"`
	AppHash       string              `json:"appHash" bson:"appHash"`
	EvidenceHash  string              `json:"evidenceHash" bson:"evidenceHash"`
	Txs           []*Transaction      `json:"-" bson:"txs"`
	Receipts      []*kai.BasicReceipt `json:"-" bson:"receipts"`

	NonceBool bool `json:"nonce_bool" bson:"nonce_bool"`
}

type Header struct {
	BlockHash     string      `json:"hash"`
	Height        uint64      `json:"height"`
	LastBlock     string      `json:"lastBlock"`
	CommitHash    string      `json:"commitHash"`
	Time          uint64      `json:"time"`
	NumTxs        uint64      `json:"numTxs"`
	GasLimit      uint64      `json:"gasLimit"`
	GasUsed       uint64      `json:"gasUsed"`
	Validator     string      `json:"validator"`
	TxHash        string      `json:"dataHash"`     // transactions
	ReceiptHash   string      `json:"receiptsRoot"` // receipt root
	Bloom         types.Bloom `json:"logsBloom"`
	ValidatorHash string      `json:"validatorHash"` // validators for the current block
	ConsensusHash string      `json:"consensusHash"` // hash of current consensus
	AppHash       string      `json:"appHash"`       // txs state
	EvidenceHash  string      `json:"evidenceHash"`  // hash of evidence
}

type BlockList struct {
	Blocks []*Block `json:"blocks" bson:"blocks"`
}
