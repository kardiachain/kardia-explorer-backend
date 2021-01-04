// Package db
package db

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/zap"
	"gotest.tools/assert"
)

func Test_mongoDB_FindBlock(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return
	}
	cfgTest.Logger = logger
	ctx := context.Background()
	mgo, err := newMongoDB(cfgTest)
	assert.NilError(t, err)
	filter := FindBlockFilter{
		StartDate: "2020-01-02",
		EndDate:   "2020-01-03",
	}
	blocks, err := mgo.FindBlock(ctx, filter)
	assert.NilError(t, err)
	fmt.Println("Blocks size", len(blocks))
	for _, b := range blocks {
		fmt.Println("BLock", b)
	}
}
