package db

import (
	"context"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_mongoDB_UpdateInternalTxs(t *testing.T) {
	arrInternalTransactions := []*types.TokenTransfer{&types.TokenTransfer{
		TransactionHash: "0x475ab5810f35704e30fbb3501b3c26b74fd6fa830d6df6e5bbee86f81940acdc",
		Contract:        "0x087C82ea812a450C517D55961Dd76ED2cAc7D469",
		From:            "0xf64C35a3d5340B8493cE4CD988B3c1e890B2bD68",
		To:              "0xE09913f6Ecf7b64C6A14A8145b4ac2B51111774c",
		Value:           "2999999999999999",
		LogIndex:        "3",
		Time:            time.Time{},
	}}

	dbClient, err := GetMgo()
	assert.Nil(t, err)

	err = dbClient.UpdateInternalTxs(context.Background(), arrInternalTransactions)
	assert.Nil(t, err)
}
