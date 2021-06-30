// Package receipts
package receipts

import (
	"context"
	"errors"
	"fmt"

	kClient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/kardia-explorer-backend/types"
)

func tryKRC20(l *kClient.Log) (*kClient.Log, error) {
	// Try with KRC20
	krc20ABI, err := kClient.KRC20ABI()
	if err != nil {
		return nil, err
	}
	unpackLog, err := kClient.UnpackLog(l, krc20ABI)
	if err != nil {
		return nil, err
	}
	fmt.Printf("UnpackLogs: %+v \n ", unpackLog)
	if l.ArgumentsName != "index_topic_1 address from, index_topic_2 address to, uint256 value" {
		return nil, errors.New("not valid krc20 transfer")
	}

	return unpackLog, nil
}

func tryKRC721(l *kClient.Log) (*kClient.Log, error) {
	// Try with KRC20
	krc721ABI, err := kClient.KRC721ABI()
	if err != nil {
		return nil, err
	}
	unpackLog, err := kClient.UnpackLog(l, krc721ABI)
	if err != nil {
		return nil, err
	}
	if l.ArgumentsName != "index_topic_1 address from, index_topic_2 address to, index_topic_3 uint256 tokenId" {
		return nil, errors.New("not valid krc20 transfer")
	}

	return unpackLog, nil
}

func (s *Server) insertTokenTransfer(ctx context.Context, log *kClient.Log) error {
	var (
		from, to, value string
		ok              bool
	)
	from, ok = log.Arguments["from"].(string)
	if !ok {
		return nil
	}
	to, ok = log.Arguments["to"].(string)
	if !ok {
		return nil
	}
	value, ok = log.Arguments["value"].(string)
	if !ok {
		return nil
	}

	internalTx := &types.TokenTransfer{
		TransactionHash: log.TxHash,
		BlockHeight:     log.BlockHeight,
		Contract:        log.Address,
		From:            from,
		To:              to,
		Value:           value,
		LogIndex:        log.Index,
	}

	return s.db.InsertInternalTxs(ctx, internalTx)
}

func (s *Server) upsertTokenHolder(ctx context.Context, log *kClient.Log) error {
	var (
		from, to, value string
		ok              bool
	)
	from, ok = log.Arguments["from"].(string)
	if !ok {
		return nil
	}
	to, ok = log.Arguments["to"].(string)
	if !ok {
		return nil
	}
	value, ok = log.Arguments["value"].(string)
	if !ok {
		return nil
	}

	internalTx := &types.TokenTransfer{
		TransactionHash: log.TxHash,
		BlockHeight:     log.BlockHeight,
		Contract:        log.Address,
		From:            from,
		To:              to,
		Value:           value,
		LogIndex:        log.Index,
	}

	return s.db.InsertInternalTxs(ctx, internalTx)
}
