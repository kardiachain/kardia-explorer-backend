// Package db
package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_mongoDB_FindContractCreationTxs(t *testing.T) {
	ctx := context.Background()
	mgo, err := SetupTestMGO()
	assert.Nil(t, err)

	txs, err := mgo.FindContractCreationTxs(ctx)
	assert.Nil(t, err)
	fmt.Println("TotalTxs", len(txs))
	//for _, tx := range txs {
	//	//fmt.Println("tx", tx)
	//}
}
