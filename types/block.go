package types

import (
	"encoding/json"

	"github.com/kardiachain/go-kardiamain/types"
)
type Header struct {
	Hash   string `json:"hash" bson:"blockHash"`
	Height uint64 `json:"height" bson:"height"`

	CommitHash string `json:"commitHash" bson:"commitHash"`
	GasLimit   uint64 `json:"gasLimit" bson:"gasLimit"`
	GasUsed    uint64 `json:"gasUsed" bson:"gasUsed"`
	NumTxs     uint64 `json:"numTxs" bson:"numTxs"`
	Time       uint64 `json:"time" bson:"time"`
	Validator  string `json:"validator" bson:"validator"`

	LastBlock string `json:"lastBlock" bson:"lastBlock"`

	DataHash     string      `json:"dataHash" bson:"dataHash"`
	ReceiptsRoot string      `json:"receiptsRoot" bson:"receiptsRoot"`
	LogsBloom    types.Bloom `json:"logsBloom" bson:"logsBloom"`

	ValidatorHash string `json:"validatorHash" bson:"validatorHash"`
	ConsensusHash string `json:"consensusHash" bson:"consensusHash"`
	AppHash       string `json:"appHash" bson:"appHash"`
	EvidenceHash  string `json:"evidenceHash" bson:"evidenceHash"`

	// Dual nodes
	NumDualEvents  uint64 `json:"numDualEvents" bson:"numDualEvents"`
	DualEventsHash string `json:"dualEventsHash" bson:"dualEventsHash"`
}

type Block struct {
	Hash   string `json:"hash" bson:"blockHash"`
	Height uint64 `json:"height" bson:"height"`

	CommitHash string `json:"commitHash" bson:"commitHash"`
	GasLimit   uint64 `json:"gasLimit" bson:"gasLimit"`
	GasUsed    uint64 `json:"gasUsed" bson:"gasUsed"`
	NumTxs     uint64 `json:"numTxs" bson:"numTxs"`
	Time       uint64 `json:"time" bson:"time"`
	Validator  string `json:"validator" bson:"validator"`

	LastBlock string `json:"lastBlock" bson:"lastBlock"`

	DataHash     string      `json:"dataHash" bson:"dataHash"`
	ReceiptsRoot string      `json:"receiptsRoot" bson:"receiptsRoot"`
	LogsBloom    types.Bloom `json:"logsBloom" bson:"logsBloom"`

	ValidatorHash string `json:"validatorHash" bson:"validatorHash"`
	ConsensusHash string `json:"consensusHash" bson:"consensusHash"`
	AppHash       string `json:"appHash" bson:"appHash"`
	EvidenceHash  string `json:"evidenceHash" bson:"evidenceHash"`

	// Dual nodes
	NumDualEvents  uint64 `json:"numDualEvents" bson:"numDualEvents"`
	DualEventsHash string `json:"dualEventsHash" bson:"dualEventsHash"`

	Txs      []*Transaction `json:"txs" bson:"-"`
	Receipts []*Receipt     `json:"receipts" bson:"-"`
}

func (b *Block) String() string {
	data, err := json.Marshal(b)
	if err != nil {
		return ""
	}
	return string(data)
}
