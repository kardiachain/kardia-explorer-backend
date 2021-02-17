// Package db
package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/kardiachain/explorer-backend/types"
)

var (
	seedValidators = []*types.Validator{
		{},
	}
)

func Test_mongoDB_UpdateValidators(t *testing.T) {
	ctx := context.Background()
	// Do insert seed validators
	seedValidators[0].Name = "updatedValidatorName0"
	seedValidators[1].Name = "updatedValidatorName1"
	// Append new test validator since we going to upsert
	newValidator := &types.Validator{}
	seedValidators = append(seedValidators, newValidator)
	size := len(seedValidators)

	mgo, err := GetTestMgo()
	assert.Nil(t, err)

	assert.Nil(t, mgo.UpsertValidators(ctx, seedValidators))

	// Make sure if upsert success
	filter := ValidatorsFilter{}
	validators, err := mgo.Validators(ctx, filter)
	assert.Nil(t, err)
	assert.Equal(t, size, len(validators))
	assert.Equal(t, validators[0].Name, "updatedValidatorName0", "first validator should be")
	assert.Equal(t, validators[1].Name, "updatedValidatorName1", "second validator should be")

	// Clean validators
	_, err = mgo.wrapper.C(cValidators).RemoveAll(bson.M{})
	assert.Nil(t, err)
}
