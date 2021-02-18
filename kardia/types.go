// Package kardia
package kardia

import (
	"math/big"
)

type UnbondedRecord struct {
	Balance        *big.Int
	CompletionTime *big.Int
}
