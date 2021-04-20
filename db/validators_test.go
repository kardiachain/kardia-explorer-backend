// Package db
package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_mongoDB_UpdateProposers(t *testing.T) {
	mgo, err := GetMgo()
	assert.Nil(t, err)

	assert.Nil(t, mgo.UpdateProposers(context.Background(), []string{"0x50a26DF56fC91eECF7f25D52eFB4eFAB56Dacf08"}))
}
