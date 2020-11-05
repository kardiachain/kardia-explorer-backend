// Package server
package server

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/labstack/echo"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/api"
	"github.com/kardiachain/explorer-backend/types"
)

type PagingResponse struct {
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
	Total uint64      `json:"total"`
	Data  interface{} `json:"data"`
}

func (s *Server) Ping(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) Info(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) Search(c echo.Context) error {
	var (
		ctx = context.Background()
		err error
	)
	for paramName, paramValue := range c.QueryParams() {
		switch paramName {
		case "address":
			pageParams := c.QueryParam("page")
			limitParams := c.QueryParam("limit")
			page, err := strconv.Atoi(pageParams)
			if err != nil {
				page = 1
			}
			limit, err := strconv.Atoi(limitParams)
			if err != nil {
				limit = 20
			}
			pagination := &types.Pagination{
				Skip:  page * limit,
				Limit: limit,
			}
			txs, total, err := s.dbClient.TxsByAddress(ctx, paramValue[0], pagination)
			s.Logger.Info("search tx by hash:", zap.String("address", paramValue[0]))
			balance, err := s.kaiClient.BalanceAt(ctx, paramValue[0], nil)
			if err != nil {
				return err
			}
			s.logger.Debug("Balance", zap.String("address", paramValue[0]), zap.String("balance", balance))
			return api.OK.SetData(struct {
				Balance string         `json:"balance"`
				Txs     PagingResponse `json:"txs"`
			}{
				Balance: balance,
				Txs: PagingResponse{
					Page:  page,
					Limit: limit,
					Total: total,
					Data:  txs,
				},
			}).Build(c)
		case "txHash":
			s.Logger.Info("search tx by hash:", zap.String("txHash", paramValue[0]))
			tx, err := s.dbClient.TxByHash(ctx, paramValue[0])
			if err != nil {
				return api.Invalid.Build(c)
			}
			return api.OK.SetData(tx).Build(c)
		case "blockHash":
			s.Logger.Info("search block by hash:", zap.String("blockHash", paramValue[0]))
			block, err := s.dbClient.BlockByHash(ctx, paramValue[0])
			if err != nil {
				return api.Invalid.Build(c)
			}
			return api.OK.SetData(block).Build(c)
		case "blockHeight":
			blockHeight, err := strconv.ParseUint(paramValue[0], 10, 64)
			s.Logger.Info("search block by height:", zap.Uint64("blockHeight", blockHeight))
			if err != nil || blockHeight < 0 {
				return api.Invalid.Build(c)
			}
			block, err := s.dbClient.BlockByHeight(ctx, blockHeight)
			if err != nil {
				return api.Invalid.Build(c)
			}
			return api.OK.SetData(block).Build(c)
		default:
			if err != nil {
				return api.Invalid.Build(c)
			}
		}
	}
	return api.Invalid.Build(c)
}

func (s *Server) Stats(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	blocks, err := s.dbClient.Blocks(ctx, &types.Pagination{
		Skip:  0,
		Limit: 11,
	})
	if err != nil {
		return api.InternalServer.Build(c)
	}

	type Stat struct {
		NumTxs uint64 `json:"numTxs"`
		Time   uint64 `json:"time"`
	}

	var stats []*Stat
	for _, b := range blocks {
		stat := &Stat{
			NumTxs: b.NumTxs,
			Time:   uint64(b.Time.Unix()),
		}
		stats = append(stats, stat)
	}

	return api.OK.SetData(struct {
		Data interface{} `json:"data"`
	}{
		Data: stats,
	}).Build(c)
}

func (s *Server) TotalHolders(c echo.Context) error {
	ctx := context.Background()
	totalHolders := s.cacheClient.TotalHolders(ctx)
	return api.OK.SetData(totalHolders).Build(c)
}

func (s *Server) Nodes(c echo.Context) error {
	ctx := context.Background()
	nodes, err := s.kaiClient.NodesInfo(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(nodes).Build(c)
}

func (s *Server) TokenInfo(c echo.Context) error {
	ctx := context.Background()
	if !s.cacheClient.IsRequestToCoinMarket(ctx) {
		tokenInfo, err := s.cacheClient.TokenInfo(ctx)
		if err != nil {
			return api.InternalServer.Build(c)
		}
		return api.OK.SetData(tokenInfo).Build(c)
	}

	tokenInfo, err := s.infoServer.TokenInfo(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}

	return api.OK.SetData(tokenInfo).Build(c)
}

func (s *Server) TPS(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) ValidatorStats(c echo.Context) error {
	s.logger.Debug("ValidatorInfo", zap.Any("URL", c.Param("rpcURL")))
	ctx := context.Background()
	rpcURL := c.Param("rpcURL")
	validator, err := s.kaiClient.Validator(ctx, rpcURL)
	if err != nil {
		return api.Invalid.Build(c)
	}
	s.logger.Debug("ValidatorInfo", zap.Any("ValidatorInfo", validator))
	return api.OK.SetData(validator).Build(c)
}

func (s *Server) Validators(c echo.Context) error {
	ctx := context.Background()
	validators, err := s.kaiClient.Validators(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}
	s.logger.Debug("Validators", zap.Any("validators", validators))
	return api.OK.SetData(validators).Build(c)
}

func (s *Server) Blocks(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var (
		page, limit int
		err         error
		blocks      []*types.Block
	)
	pageParams := c.QueryParam("page")
	limitParams := c.QueryParam("limit")
	page, err = strconv.Atoi(pageParams)
	if err != nil {
		page = 1
	}
	limit, err = strconv.Atoi(limitParams)
	if err != nil {
		limit = 20
	}
	if limit > 100 {
		return api.Invalid.Build(c)
	}
	pagination := &types.Pagination{
		Skip:  page * limit,
		Limit: limit,
	}

	// todo @londnd: implement read from cache,
	blocks, err = s.cacheClient.LatestBlocks(ctx, pagination)
	if err != nil || blocks == nil {
		s.logger.Debug("Cannot get latest blocks from cache", zap.Error(err))
		blocks, err = s.dbClient.Blocks(ctx, pagination)
		if err != nil {
			s.logger.Debug("Cannot get latest blocks from db", zap.Error(err))
			return api.InternalServer.Build(c)
		}
		s.logger.Debug("Got latest blocks from db")
	}
	s.logger.Debug("Got latest blocks from cache")

	return api.OK.SetData(struct {
		Page  int         `json:"page"`
		Limit int         `json:"limit"`
		Data  interface{} `json:"data"`
	}{
		Page:  page,
		Limit: limit,
		Data:  blocks,
	}).Build(c)
}

func (s *Server) Block(c echo.Context) error {
	ctx := context.Background()
	blockHashOrHeightStr := c.Param("block")
	var (
		block *types.Block
		err   error
	)
	if strings.HasPrefix(blockHashOrHeightStr, "0x") {
		s.Logger.Info("get block by hash:", zap.String("blockHash", blockHashOrHeightStr))
		block, err = s.dbClient.BlockByHash(ctx, blockHashOrHeightStr)
		if err != nil {
			return api.Invalid.Build(c)
		}
	} else {
		blockHeight, err := strconv.ParseUint(blockHashOrHeightStr, 10, 64)
		s.Logger.Info("get block by height:", zap.Uint64("blockHeight", blockHeight))
		if err != nil || blockHeight <= 0 {
			return api.Invalid.Build(c)
		}
		block, err = s.dbClient.BlockByHeight(ctx, blockHeight)
		if err != nil {
			return api.Invalid.Build(c)
		}
	}

	return api.OK.SetData(block).Build(c)
}

func (s *Server) BlockExist(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) BlockTxs(c echo.Context) error {
	ctx := context.Background()
	var page, limit int
	var err error
	block := c.Param("block")
	pageParams := c.QueryParam("page")
	limitParams := c.QueryParam("limit")
	page, err = strconv.Atoi(pageParams)
	if err != nil {
		page = 1
	}
	limit, err = strconv.Atoi(limitParams)
	if err != nil {
		limit = 20
	}
	// Random number of txs of block hash

	var (
		txs   []*types.Transaction
		total uint64
	)
	pagination := &types.Pagination{
		Skip:  page * limit,
		Limit: limit,
	}
	if strings.HasPrefix(block, "0x") {
		s.logger.Debug("fetch block txs by hash", zap.String("hash", block))

		txs, total, err = s.dbClient.TxsByBlockHash(ctx, block, pagination)
		if err != nil {
			s.logger.Debug("cannot get txs by block hash", zap.String("blockHash", block))
			return api.InternalServer.Build(c)
		}
	} else {
		s.logger.Debug("fetch block txs by height", zap.String("height", block))
		height, err := strconv.Atoi(block)
		if err != nil {
			return api.Invalid.Build(c)
		}

		if height <= 0 {
			return api.Invalid.Build(c)
		}
		// Convert to height
		txs, total, err = s.dbClient.TxsByBlockHeight(ctx, uint64(height), pagination)
		if err != nil {
			return api.Invalid.Build(c)
		}
	}

	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  txs,
	}).Build(c)
}

func (s *Server) Txs(c echo.Context) error {
	ctx := context.Background()
	var (
		page, limit int
		err         error
	)
	pageParams := c.QueryParam("page")
	limitParams := c.QueryParam("limit")
	page, err = strconv.Atoi(pageParams)
	if err != nil {
		page = 1
	}
	limit, err = strconv.Atoi(limitParams)
	if err != nil {
		limit = 20
	}

	var txs []*types.Transaction
	pagination := &types.Pagination{
		Skip:  page * limit,
		Limit: limit,
	}

	txs, err = s.cacheClient.LatestTransactions(ctx, pagination)
	if err != nil || txs == nil {
		s.logger.Debug("Cannot get latest txs from cache", zap.Error(err))
		txs, err = s.dbClient.LatestTxs(ctx, pagination)
		if err != nil {
			s.logger.Debug("Cannot get latest txs from db", zap.Error(err))
			return api.Invalid.Build(c)
		}
		s.logger.Debug("Got latest txs from db")
	} else {
		s.logger.Debug("Got latest txs from cached")
	}

	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: s.cacheClient.TotalTxs(ctx),
		Data:  txs,
	}).Build(c)
}

func (s *Server) Addresses(c echo.Context) error {
	var page, limit int
	var err error
	//blockHash := c.Param("blockHash")
	pageParams := c.QueryParam("page")
	limitParams := c.QueryParam("limit")
	page, err = strconv.Atoi(pageParams)
	if err != nil {
		page = 1
	}
	limit, err = strconv.Atoi(limitParams)
	if err != nil {
		limit = 20
	}

	var addresses []*types.Address
	for i := 0; i < limit; i++ {
		address := &types.Address{}
		if err := faker.FakeData(address); err != nil {
			return err
		}
		addresses = append(addresses, address)
	}

	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: uint64(limit * 10),
		Data:  addresses,
	}).Build(c)
}

func (s *Server) Balance(c echo.Context) error {
	ctx := context.Background()
	address := c.Param("address")
	balance, err := s.kaiClient.BalanceAt(ctx, address, nil)
	if err != nil {
		return err
	}
	s.logger.Debug("Balance", zap.String("address", address), zap.String("balance", balance))

	return api.OK.SetData(balance).Build(c)
}

func (s *Server) AddressTxs(c echo.Context) error {
	ctx := context.Background()
	var page, limit int
	var err error
	address := c.Param("address")
	pageParams := c.QueryParam("page")
	limitParams := c.QueryParam("limit")
	page, err = strconv.Atoi(pageParams)
	if err != nil {
		page = 1
	}
	limit, err = strconv.Atoi(limitParams)
	if err != nil {
		limit = 20
	}
	pagination := &types.Pagination{
		Skip:  page * limit,
		Limit: limit,
	}

	txs, total, err := s.dbClient.TxsByAddress(ctx, address, pagination)
	if err != nil {
		s.logger.Debug("error while get address txs:", zap.Error(err))
		return err
	}

	s.logger.Info("address txs:", zap.String("address", address))
	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  txs,
	}).Build(c)
}

func (s *Server) AddressHolders(c echo.Context) error {
	var page, limit int
	var err error
	//blockHash := c.Param("blockHash")
	pageParams := c.QueryParam("page")
	limitParams := c.QueryParam("limit")
	page, err = strconv.Atoi(pageParams)
	if err != nil {
		page = 1
	}
	limit, err = strconv.Atoi(limitParams)
	if err != nil {
		limit = 20
	}

	var holders []*types.TokenHolder
	for i := 0; i < limit; i++ {
		holder := &types.TokenHolder{}
		if err := faker.FakeData(&holder); err != nil {
			return err
		}
		holders = append(holders, holder)
	}

	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: uint64(limit * 15),
		Data:  holders,
	}).Build(c)
}

func (s *Server) AddressOwnedTokens(c echo.Context) error {
	return api.OK.Build(c)
}

// AddressInternal
func (s *Server) AddressInternalTxs(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) AddressContract(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) AddressTxByNonce(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) AddressTxHashByNonce(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) TxByHash(c echo.Context) error {
	ctx := context.Background()
	txHash := c.Param("txHash")
	if txHash == "" {
		return api.Invalid.Build(c)
	}

	var tx *types.Transaction
	tx, err := s.dbClient.TxByHash(ctx, txHash)
	if err != nil {
		return api.InternalServer.Build(c)
	}

	return api.OK.SetData(tx).Build(c)
}

func (s *Server) TxExist(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) Contracts(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) BlockTime(c echo.Context) error {
	panic("implement me")
}
