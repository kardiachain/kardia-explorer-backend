package types

import (
	"github.com/kardiachain/go-kardiamain/lib/common"
)

type Signer struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Region string `json:"region"`
}

type SignerStats struct {
	SignerAddress common.Address `json:"signer_address"`
	BlocksCount   int            `json:"blocks_count"`
}

type BlockRange struct {
	StartBlock uint64 `json:"start_block"`
	EndBlock   uint64 `json:"end_block"`
}

type SignersStats struct {
	// front needs arr here
	SignerStats []SignerStats `json:"signer_stats"`
	BlockRange  BlockRange    `json:"block_range"`
	Range       string        `json:"range"`
}
