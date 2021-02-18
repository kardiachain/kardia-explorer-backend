// Package types
package types

import (
	"math/big"
)

type UnbondedRecord struct {
	Balances        *big.Int `json:"balance"`
	CompletionTimes *big.Int `json:"completionTime"`
}
