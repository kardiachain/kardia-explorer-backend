// Package receipts
package receipts

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	kClient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/go-kardia/lib/abi"
	"github.com/kardiachain/go-kardia/types/time"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/kardiachain/kardia-explorer-backend/utils"
	"go.uber.org/zap"
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

func (s *Server) insertKRC20Transfer(ctx context.Context, log *kClient.Log) error {
	lgr := s.logger
	var (
		from, to, value string
		ok              bool
	)
	from, ok = log.Arguments["from"].(string)
	if !ok {
		lgr.Error("cannot get from")
		return nil
	}
	to, ok = log.Arguments["to"].(string)
	if !ok {
		lgr.Error("cannot get to")
		return nil
	}
	value, ok = log.Arguments["value"].(string)
	if !ok {
		lgr.Error("cannot get value")
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
		Time:            log.Time,
	}
	lgr.Info("New KRC20 transfer", zap.Any("TX", internalTx))
	return s.db.InsertInternalTxs(ctx, internalTx)
}

func (s *Server) insertKRC721Transfer(ctx context.Context, log *kClient.Log) error {
	lgr := s.logger
	var (
		from, to, tokenId string
		ok                bool
	)
	from, ok = log.Arguments["from"].(string)
	if !ok {
		lgr.Error("cannot get from")
		return nil
	}
	to, ok = log.Arguments["to"].(string)
	if !ok {
		lgr.Error("cannot get to")
		return nil
	}
	tokenId, ok = log.Arguments["tokenId"].(string)
	if !ok {
		lgr.Error("cannot get tokenId")
		return nil
	}

	internalTx := &types.TokenTransfer{
		TransactionHash: log.TxHash,
		BlockHeight:     log.BlockHeight,
		Contract:        log.Address,
		From:            from,
		To:              to,
		TokenID:         tokenId,
		LogIndex:        log.Index,
		Time:            log.Time,
	}
	lgr.Info("New KRC721 transfer", zap.Any("TX", internalTx))
	return s.db.InsertInternalTxs(ctx, internalTx)
}

func (s *Server) upsertKRC20Holder(ctx context.Context, log *kClient.Log) error {
	var (
		from, to string
		ok       bool
	)
	from, ok = log.Arguments["from"].(string)
	if !ok {
		return errors.New("invalid from address")
	}
	to, ok = log.Arguments["to"].(string)
	if !ok {
		return errors.New("invalid to address")
	}
	holders := make([]*types.TokenHolder, 2)
	token, err := kClient.NewToken(s.node, log.Address)
	if err != nil {
		return err
	}
	krc20Info, err := token.KRC20Info(ctx)
	if err != nil {
		return err
	}
	fromBalance, err := token.HolderBalance(ctx, from)
	if err != nil {
		return err
	}
	toBalance, err := token.HolderBalance(ctx, to)
	if err != nil {
		return err
	}

	holders[0] = &types.TokenHolder{
		ContractAddress: log.Address,
		HolderAddress:   from,
		BalanceString:   fromBalance.String(),
		BalanceFloat:    utils.BalanceToFloatWithDecimals(fromBalance, int64(krc20Info.Decimals)),
		UpdatedAt:       time.Now().Unix(),
	}
	holders[1] = &types.TokenHolder{
		ContractAddress: log.Address,
		HolderAddress:   to,
		BalanceString:   toBalance.String(),
		BalanceFloat:    utils.BalanceToFloatWithDecimals(toBalance, int64(krc20Info.Decimals)),
		UpdatedAt:       time.Now().Unix(),
	}
	if err := s.db.UpsertHolders(ctx, holders); err != nil {
		return err
	}
	return nil
}

// todo: Update inventory for KRC721, now just ignore
func (s *Server) upsertKRC721Holder(ctx context.Context, log *kClient.Log) error {
	return nil
	var (
		from, to string
		ok       bool
	)
	from, ok = log.Arguments["from"].(string)
	if !ok {
		return errors.New("invalid from address")
	}
	to, ok = log.Arguments["to"].(string)
	if !ok {
		return errors.New("invalid to address")
	}
	holders := make([]*types.TokenHolder, 2)
	token, err := kClient.NewToken(s.node, log.Address)
	if err != nil {
		return err
	}
	krc20Info, err := token.KRC20Info(ctx)
	if err != nil {
		return err
	}
	fromBalance, err := token.HolderBalance(ctx, from)
	if err != nil {
		return err
	}
	toBalance, err := token.HolderBalance(ctx, to)
	if err != nil {
		return err
	}

	holders[0] = &types.TokenHolder{
		ContractAddress: log.Address,
		HolderAddress:   from,
		BalanceString:   fromBalance.String(),
		BalanceFloat:    utils.BalanceToFloatWithDecimals(fromBalance, int64(krc20Info.Decimals)),
		UpdatedAt:       time.Now().Unix(),
	}
	holders[1] = &types.TokenHolder{
		ContractAddress: log.Address,
		HolderAddress:   to,
		BalanceString:   toBalance.String(),
		BalanceFloat:    utils.BalanceToFloatWithDecimals(toBalance, int64(krc20Info.Decimals)),
		UpdatedAt:       time.Now().Unix(),
	}
	if err := s.db.UpsertHolders(ctx, holders); err != nil {
		return err
	}
	return nil
}

func (s *Server) decodeSMCABIFromBase64(ctx context.Context, abiStr, smcAddr string) (*abi.ABI, error) {
	abiData, err := base64.StdEncoding.DecodeString(abiStr)
	if err != nil {
		s.logger.Warn("Cannot decode smc abi", zap.Error(err))
		return nil, err
	}
	jsonABI, err := abi.JSON(bytes.NewReader(abiData))
	if err != nil {
		s.logger.Warn("Cannot convert decoded smc abi to JSON abi", zap.Error(err))
		return nil, err
	}
	return &jsonABI, nil
}
