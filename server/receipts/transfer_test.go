package receipts

import (
	"context"
	kClient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/panjf2000/ants/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestProcessKRC721Logs(t *testing.T) {
	lgr, err := zap.NewDevelopment()
	assert.Nil(t, err)
	node, err := kClient.NewNode("https://dev.kardiachain.io", lgr)
	assert.Nil(t, err)
	srv := Server{
		db:     nil,
		node:   nil,
		cache:  nil,
		logger: lgr,
		p:      ants.PoolWithFunc{},
	}
	c := &types.Contract{}
	r, err := node.GetTransactionReceipt(context.Background(), "0x7edbf9c942e0908ce095d975d758894d7a082dc6d6b10281a53a4bb21d718085")
	assert.Nil(t, err)
	for _, l := range r.Logs {
		// Process if transfer event
		if l.Topics[0] == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
			lgr.Debug("PreprocessLog", zap.Any("Log", l))
			srv.onKRC721Transfer(context.Background(), c, l)
			assert.Nil(t, err)
		}

		// Process if mint/burn event

	}

}
