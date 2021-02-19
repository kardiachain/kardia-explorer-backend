// Package db
package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_mongoDB_GetAddressInfo(t *testing.T) {
	ctx := context.Background()
	mgo, err := GetMgo()
	assert.Nil(t, err)

	address, err := mgo.AddressByHash(ctx, "0x4f36A53DC32272b97Ae5FF511387E2741D727bdb")
	assert.Nil(t, err)
	fmt.Println("address info", address)
}
