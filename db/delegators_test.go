// Package db
package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelegator_UniqueDelegators(t *testing.T) {
	dbClient, err := GetMgo()
	assert.Nil(t, err)
	totalDelegators, err := dbClient.UniqueDelegators(context.Background())
	assert.Nil(t, err)
	fmt.Println("total delegators", totalDelegators)
}

func TestDelegator_ValidatorsStakedAmount(t *testing.T) {
	dbClient, err := GetMgo()
	assert.Nil(t, err)

	validators, err := dbClient.Validators(context.Background(), ValidatorsFilter{})
	assert.Nil(t, err)
	var validatorAddresses []string
	for _, v := range validators {
		validatorAddresses = append(validatorAddresses, v.Address)
	}

	totalDelegators, err := dbClient.GetStakedOfAddresses(context.Background(), validatorAddresses)
	assert.Nil(t, err)
	fmt.Println("total delegators", totalDelegators)
}
