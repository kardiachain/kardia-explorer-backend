// Package api
package api

import (
	"bytes"
	"context"
	"io/ioutil"
	"math"
	"strconv"

	"github.com/kardiachain/go-kardia"
	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

func (s *Server) GetInternalTxs(c echo.Context) error {
	pagination, page, limit := getPagingOption(c)
	contractAddress := c.QueryParam("contractAddress")
	if contractAddress != "" {
		result, totalRecord, err := s.internalTxsOfContract(contractAddress, pagination)
		if err != nil {
			return Invalid.Build(c)
		}
		return OK.SetData(PagingResponse{
			Page:  page,
			Limit: limit,
			Total: totalRecord,
			Data:  result,
		}).Build(c)
	}
	address := c.QueryParam("address")
	if address != "" {
		result, totalRecord, err := s.internalTxsOfAddress(address, pagination)
		if err != nil {
			return Invalid.Build(c)
		}
		return OK.SetData(PagingResponse{
			Page:  page,
			Limit: limit,
			Total: totalRecord,
			Data:  result,
		}).Build(c)
	}

	transactionHash := c.QueryParam("txHash")
	if transactionHash != "" {
		result, totalRecord, err := s.internalTxsOfTransaction(transactionHash, pagination)
		if err != nil {
			return Invalid.Build(c)
		}
		return OK.SetData(PagingResponse{
			Page:  page,
			Limit: limit,
			Total: totalRecord,
			Data:  result,
		}).Build(c)
	}

	return Invalid.Build(c)

}

func (s *Server) internalTxsOfAddress(address string, pagination *types.Pagination) ([]*InternalTransaction, uint64, error) {
	ctx := context.Background()
	filterCrit := &types.InternalTxsFilter{
		Pagination: pagination,
		Address:    address,
	}
	iTxs, total, err := s.dbClient.GetListInternalTxs(ctx, filterCrit)
	if err != nil {
		s.logger.Warn("Cannot get internal txs from db", zap.Error(err))
	}
	var (
		result           = make([]*InternalTransaction, len(iTxs))
		fromInfo, toInfo *types.Address
	)
	for i := range iTxs {
		result[i] = &InternalTransaction{
			Log: &types.Log{
				Address: iTxs[i].Contract,
				Time:    iTxs[i].Time,
				TxHash:  iTxs[i].TransactionHash,
			},
			From:  iTxs[i].From,
			To:    iTxs[i].To,
			Value: iTxs[i].Value,
		}
		fromInfo, _ = s.getAddressDetail(ctx, iTxs[i].From)
		if fromInfo != nil {
			result[i].FromName = fromInfo.Name
		}
		toInfo, _ = s.getAddressDetail(ctx, iTxs[i].To)
		if toInfo != nil {
			result[i].ToName = toInfo.Name
		}
		krcTokenInfo, _ := s.getTokenInfo(ctx, iTxs[i].Contract)
		if krcTokenInfo != nil {
			result[i].KRCTokenInfo = krcTokenInfo
		}
	}
	return result, total, nil
}

func (s *Server) internalTxsOfTransaction(txHash string, pagination *types.Pagination) ([]*InternalTransaction, uint64, error) {
	ctx := context.Background()
	filterCrit := &types.InternalTxsFilter{
		Pagination:      pagination,
		TransactionHash: txHash,
	}
	iTxs, total, err := s.dbClient.GetListInternalTxs(ctx, filterCrit)
	if err != nil {
		s.logger.Warn("Cannot get internal txs from db", zap.Error(err))
	}
	var (
		result           = make([]*InternalTransaction, len(iTxs))
		fromInfo, toInfo *types.Address
	)
	for i := range iTxs {
		result[i] = &InternalTransaction{
			Log: &types.Log{
				Address: iTxs[i].Contract,
				Time:    iTxs[i].Time,
				TxHash:  iTxs[i].TransactionHash,
			},
			From:  iTxs[i].From,
			To:    iTxs[i].To,
			Value: iTxs[i].Value,
		}
		fromInfo, _ = s.getAddressDetail(ctx, iTxs[i].From)
		if fromInfo != nil {
			result[i].FromName = fromInfo.Name
		}
		toInfo, _ = s.getAddressDetail(ctx, iTxs[i].To)
		if toInfo != nil {
			result[i].ToName = toInfo.Name
		}
		krcTokenInfo, _ := s.getTokenInfo(ctx, iTxs[i].Contract)
		if krcTokenInfo != nil {
			result[i].KRCTokenInfo = krcTokenInfo
		}
	}
	return result, total, nil
}

func (s *Server) internalTxsOfContract(contractAddress string, pagination *types.Pagination) ([]*InternalTransaction, uint64, error) {
	ctx := context.Background()
	filterCrit := &types.InternalTxsFilter{
		Pagination: pagination,
		Contract:   contractAddress,
	}
	iTxs, total, err := s.dbClient.GetListInternalTxs(ctx, filterCrit)
	if err != nil {
		s.logger.Warn("Cannot get internal txs from db", zap.Error(err))
	}
	var (
		result           = make([]*InternalTransaction, len(iTxs))
		fromInfo, toInfo *types.Address
	)
	// Get contract info
	contractInfo, _, err := s.dbClient.Contract(ctx, contractAddress)
	if err != nil {
		return nil, 0, err
	}
	tokenInfo := &types.KRCTokenInfo{
		Address:     contractInfo.Address,
		Logo:        contractInfo.Logo,
		TokenName:   contractInfo.Name,
		TokenType:   contractInfo.Type,
		TokenSymbol: contractInfo.Symbol,
		Decimals:    int64(contractInfo.Decimals),
		TotalSupply: contractInfo.TotalSupply,
	}
	for i := range iTxs {
		result[i] = &InternalTransaction{
			Log: &types.Log{
				Address: iTxs[i].Contract,
				Time:    iTxs[i].Time,
				TxHash:  iTxs[i].TransactionHash,
			},
			From:    iTxs[i].From,
			To:      iTxs[i].To,
			Value:   iTxs[i].Value,
			TokenID: iTxs[i].TokenID,
		}
		fromInfo, _ = s.getAddressDetail(ctx, iTxs[i].From)
		if fromInfo != nil {
			result[i].FromName = fromInfo.Name
		}
		toInfo, _ = s.getAddressDetail(ctx, iTxs[i].To)
		if toInfo != nil {
			result[i].ToName = toInfo.Name
		}
		result[i].KRCTokenInfo = tokenInfo
	}
	return result, total, nil
}

func (s *Server) UpdateInternalTxs(c echo.Context) error {
	var (
		ctx             = context.Background()
		crit            *types.TxsFilter
		internalTxsCrit *types.InternalTxsFilter
		lgr             = s.logger.With(zap.String("api", "UpdateInternalTxs"))
		bodyBytes, _    = ioutil.ReadAll(c.Request().Body)
	)
	if c.Request().Header.Get("Authorization") != s.authorizationSecret {
		lgr.Warn("Cannot authorization request")
		return Unauthorized.Build(c)
	}
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	if err := c.Bind(&crit); err != nil {
		lgr.Error("cannot bind txs filter", zap.Error(err))
		return Invalid.Build(c)
	}
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	if err := c.Bind(&internalTxsCrit); err != nil {
		lgr.Error("cannot bind internal txs filter", zap.Error(err))
		return Invalid.Build(c)
	}

	// filter logs from this initial height to "latest" which satisfy the
	var (
		logs              []*types.Log
		latestBlockHeight uint64 = math.MaxUint64
		toBlock           uint64
	)
	fromBlock, err := strconv.ParseUint(c.QueryParam("from"), 10, 64)
	toBlock, err2 := strconv.ParseUint(c.QueryParam("to"), 10, 64)
	if err == nil && err2 == nil {
		criteria := kardia.FilterQuery{
			FromBlock: fromBlock,
			ToBlock:   toBlock,
			Addresses: []common.Address{common.HexToAddress(crit.ContractAddress)},
			Topics:    internalTxsCrit.Topics,
		}
		logs, err = s.kaiClient.GetLogs(ctx, criteria)
		if err != nil {
			lgr.Error("Cannot get contract logs from core", zap.Error(err), zap.Any("criteria", crit))
			return Invalid.Build(c)
		}
		lgr.Info("Filtering events", zap.Uint64("latestBlockHeight", latestBlockHeight), zap.Uint64("from", fromBlock), zap.Uint64("to", toBlock),
			zap.Any("criteria", criteria), zap.Int("number of logs", len(logs)))
	} else {
		// find the block height where this contract is deployed
		txs, _, err := s.dbClient.FilterTxs(ctx, crit)
		lgr.Info("UpdateInternalTxs", zap.Any("criteria", crit))
		if err != nil || len(txs) == 0 {
			lgr.Error("Cannot get the transaction where this contract was deployed", zap.Error(err))
			return Invalid.Build(c)
		}

		lgr.Info("Filtering events", zap.Uint64("from", txs[0].BlockNumber), zap.Uint64("to", latestBlockHeight))
		for i := txs[0].BlockNumber; i < latestBlockHeight; i += cfg.FilterLogsInterval {
			latestBlockHeight, err = s.kaiClient.LatestBlockNumber(ctx)
			if err != nil {
				lgr.Error("Cannot get latest block height from RPC", zap.Error(err), zap.Any("criteria", crit))
				return Invalid.Build(c)
			}
			if i+cfg.FilterLogsInterval > latestBlockHeight {
				toBlock = latestBlockHeight
			} else {
				toBlock = i + cfg.FilterLogsInterval
			}
			criteria := kardia.FilterQuery{
				FromBlock: i,
				ToBlock:   toBlock,
				Addresses: []common.Address{common.HexToAddress(crit.ContractAddress)},
				Topics:    internalTxsCrit.Topics,
			}
			partLogs, err := s.kaiClient.GetLogs(ctx, criteria)
			if err != nil {
				lgr.Error("Cannot get contract logs from core", zap.Error(err), zap.Any("criteria", crit))
				continue
			}
			lgr.Info("Filtering events", zap.Uint64("latestBlockHeight", latestBlockHeight), zap.Uint64("from", i), zap.Uint64("to", toBlock),
				zap.Any("criteria", criteria), zap.Int("number of logs", len(partLogs)))
			logs = append(logs, partLogs...)
		}
	}

	// parse logs to internal txs
	var internalTxs []*types.TokenTransfer
	smcABI, err := s.getSMCAbi(ctx, &types.Log{
		Address: cfg.SMCTypePrefix + cfg.SMCTypeKRC20,
	})
	if err != nil {
		lgr.Error("Cannot get contract ABI", zap.Error(err), zap.Any("smcAddr", crit.ContractAddress))
		return Invalid.Build(c)
	}
	for i := range logs {
		logs[i].Address = common.HexToAddress(logs[i].Address).Hex()
		decodedLog, err := s.kaiClient.UnpackLog(logs[i], smcABI)
		if err != nil {
			decodedLog = logs[i]
		}
		internalTx := s.getInternalTxs(ctx, decodedLog)
		if internalTx != nil {
			internalTxs = append(internalTxs, internalTx)
		}
	}
	// remove old internal txs satisfy this criteria
	isRemove, err := strconv.ParseInt(c.QueryParam("remove"), 10, 64)
	if err == nil && isRemove == 1 {
		if err = s.dbClient.RemoveInternalTxs(ctx, internalTxsCrit); err != nil {
			lgr.Error("Cannot delete old internal txs in db", zap.Error(err), zap.Any("criteria", internalTxsCrit))
			return Invalid.Build(c)
		}
	}

	// batch inserting to InternalTransactions collection in db
	lgr.Info("internalTxs ready to be batch inserted", zap.Any("iTxs", len(internalTxs)), zap.Any("logs", len(logs)))
	if err = s.dbClient.UpdateInternalTxs(ctx, internalTxs); err != nil {
		lgr.Error("Cannot batch inserting new internal txs in db", zap.Error(err))
		return Invalid.Build(c)
	}
	return OK.Build(c)
}
