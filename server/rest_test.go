// Package server
package server

import (
	"fmt"
	"testing"

	"github.com/kardiachain/explorer-backend/types"
)

func TestTValidators(t *testing.T) {
	val := &types.Validators{}

	for _, v := range val.Validators {
		fmt.Printf("a %+v \n", v)
	}
}
