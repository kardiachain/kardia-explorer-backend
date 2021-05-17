// Package handler
package handler

import (
	"context"
	"fmt"
	"testing"

	"github.com/kardiachain/go-kaiclient/kardia"
	"github.com/stretchr/testify/assert"
)

func TestHandler_onNewLogEvent(t *testing.T) {
	h, err := setupTestHandler()
	assert.Nil(t, err)
	fmt.Println()
	node, err := kardia.NewNode("https://dev-1.kardiachain.io", h.logger)
	assert.Nil(t, err)
	r, err := node.GetTransactionReceipt(context.Background(), "0x8287652f698e49d85bb1d5631b9d27fc472583c4bcd3dbda3683f0b2d0bea9bb")
	//r, err :=  node.GetTransactionReceipt(context.Background(), "0x8ea49e12f566a4902ce7ca06c91adf9ef05634ee8aa84635875f70e1684a9cf8")
	assert.Nil(t, err)
	fmt.Printf("Receipt: %+v \n", r)
	for _, l := range r.Logs {
		fmt.Printf("L: %+v \n", l)
		normalLogs := &kardia.Log{
			Address:     l.Address,
			Topics:      l.Topics,
			Data:        l.Data,
			BlockHeight: l.BlockHeight,
			TxHash:      l.TxHash,
			TxIndex:     l.TxIndex,
			BlockHash:   l.BlockHash,
			Index:       l.Index,
			Removed:     l.Removed,
		}
		assert.Nil(t, h.onNewLogEvent(context.Background(), normalLogs))
	}
}
