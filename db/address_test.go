// Package db
package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/utils"
)

func init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot create logger")
	}
	lgr = logger.With(zap.String("service", "unit_test"))
}

func TestAddress_Find(t *testing.T) {
	ctx := context.Background()
	mgo, err := newMongoDB(cfgTest)
	assert.Nil(t, err)

	filter := FindAddressFilter{IsHasName: true}
	addrs, err := mgo.FindAddress(ctx, filter)
	assert.Nil(t, err)

	addrMap := utils.ToAddressMap(addrs)
	fmt.Printf("Addr map %+v \n", addrMap)
}
