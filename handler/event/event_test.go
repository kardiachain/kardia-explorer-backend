// Package event
package event

import (
	"context"
	"testing"

	"github.com/kardiachain/go-kaiclient/kardia"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/db"
)

func TestEvent_Log(t *testing.T) {
	lgr, err := zap.NewDevelopment()
	assert.Nil(t, err)
	wsURL := "wss://ws-dev.kardiachain.io"
	node, err := kardia.NewNode(wsURL, lgr)
	assert.Nil(t, err)
	ctx := context.Background()
	// KRC20 transfer
	// TxHash 0x8c8bc88a68bacaa4e3f99f19b10d8cd4544151177a9956e8c90386e9714a01a0
	r, err := node.GetTransactionReceipt(ctx, "0x8c8bc88a68bacaa4e3f99f19b10d8cd4544151177a9956e8c90386e9714a01a0")
	assert.Nil(t, err)

	dbCfg := db.Config{
		DbAdapter: "",
		DbName:    "",
		URL:       "",
		MinConn:   0,
		MaxConn:   0,
		FlushDB:   false,
		Logger:    lgr,
	}
	mgo, err := db.NewClient(dbCfg)
	cacheCfg := cache.Config{
		Adapter:            "",
		URL:                "",
		DB:                 0,
		IsFlush:            false,
		BlockBuffer:        0,
		DefaultExpiredTime: 0,
		Logger:             lgr,
	}
	c, err := cache.New(cacheCfg)

	e := &Event{
		node:   node,
		db:     mgo,
		cache:  c,
		logger: lgr,
	}
	for _, l := range r.Logs {
		e.ProcessNewEventLog(ctx, l)

	}
}
