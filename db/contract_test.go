// Package db
package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContracts_GetKRC20(t *testing.T) {
	ctx := context.Background()
	mgo, err := GetMgo()
	assert.Nil(t, err)

	contracts, _, err := mgo.Contract(ctx, "0xee165121948878745E17593519d61aE04c8f07c6")
	assert.Nil(t, err)
	fmt.Println("Contract", contracts)
}
