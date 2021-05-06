// Package db
package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

func Test_mongoDB_GetAddressInfo(t *testing.T) {
	ctx := context.Background()
	mgo, err := GetMgo()
	assert.Nil(t, err)

	address, err := mgo.AddressByHash(ctx, "0x4f36A53DC32272b97Ae5FF511387E2741D727bdb")
	assert.Nil(t, err)
	fmt.Println("address info", address)
}

func TestMgo_GetAddress(t *testing.T) {
	ctx := context.Background()
	mgo, err := GetMgo()
	assert.Nil(t, err)

	address, err := mgo.AddressByHash(ctx, "0x4f36A53DC32272b97Ae5FF511387E2741D727bdb")
	assert.Nil(t, err)
	fmt.Println("address info", address)
}

func TestMgo_TxOfAddress(t *testing.T) {
	ctx := context.Background()
	mgo, err := GetMgo()
	assert.Nil(t, err)

	txs, total, err := mgo.TxsByAddress(ctx, "0x448388b598cf11A8a293d6e27B90C0Ee356F8e91", &types.Pagination{
		Skip:  0,
		Limit: 25,
	})
	assert.Nil(t, err)
	fmt.Println("Total", total)
	fmt.Println("Txs", txs)
}
