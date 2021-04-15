// Package db
package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

func TestContracts_GetKRC20(t *testing.T) {
	ctx := context.Background()
	mgo, err := GetMgo()
	assert.Nil(t, err)
	filter := &types.ContractsFilter{
		Type:       "KRC20",
		Pagination: &types.Pagination{Limit: 100, Skip: 0},
	}
	contracts, total, err := mgo.Contracts(ctx, filter)
	assert.Nil(t, err)
	fmt.Println("TotalRecords:", total)
	for _, c := range contracts {
		fmt.Println("Contract info: ", c)
	}
}
