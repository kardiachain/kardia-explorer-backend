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

func (s *Server) Ping(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) Info(c echo.Context) error {
	return api.OK.Build(c)
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
			Time:   b.Time,
		}
		stats = append(stats, stat)
	}

	return api.OK.SetData(struct {
		Data interface{} `json:"data"`
	}{
		Data: stats,
	}).Build(c)
}

func (s *Server) Nodes(c echo.Context) error {
	ctx := context.Background()
	nodes, err := s.kaiClient.NodeInfo(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(nodes).Build(c)
}

func (s *Server) TokenInfo(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) TPS(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) ValidatorStats(c echo.Context) error {
	return api.OK.Build(c)
}

func (s *Server) Validators(c echo.Context) error {
	ctx := context.Background()
	validators := s.kaiClient.Validators(ctx)
	s.logger.Debug("Validators", zap.Any("validators", validators))
	return api.OK.SetData(validators).Build(c)
}

func (s *Server) Blocks(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var page, limit int
	var err error
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

	// todo @londnd: implement read from cache,
	blocks, err := s.dbClient.Blocks(ctx, &types.Pagination{
		Skip:  page*limit - limit,
		Limit: limit,
	})
	if err != nil {
		return api.InternalServer.Build(c)
	}

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
	blockHash := c.QueryParam("hash")

	var blockHeight uint64
	blockHeightStr := c.QueryParam("height")
	if blockHeightStr != "" {
		height, err := strconv.Atoi(blockHeightStr)
		if err != nil || height <= 0 {
			return api.Invalid.Build(c)
		}
		blockHeight = uint64(height)
	}

	var block *types.Block
	var err error
	if blockHash != "" {
		block, err = s.dbClient.BlockByHash(ctx, blockHash)
		if err != nil {
			return api.Invalid.Build(c)
		}
	}

	if blockHeight > 0 {
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

	var txs []*types.Transaction
	pagination := &types.Pagination{
		Skip:  page * limit,
		Limit: limit,
	}
	if strings.HasPrefix(block, "0x") {
		s.logger.Debug("fetch block txs by hash", zap.String("hash", block))

		txs, err = s.dbClient.TxsByBlockHash(ctx, block, pagination)
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
		txs, err = s.dbClient.TxsByBlockHeight(ctx, uint64(height), pagination)
		if err != nil {
			return api.Invalid.Build(c)
		}
	}

	return api.OK.SetData(struct {
		Page  int         `json:"page"`
		Limit int         `json:"limit"`
		Total int         `json:"total"`
		Data  interface{} `json:"data"`
	}{
		Page:  page,
		Limit: limit,
		Total: limit * 15,
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

	return api.OK.SetData(struct {
		Page  int         `json:"page"`
		Limit int         `json:"limit"`
		Total int         `json:"total"`
		Data  interface{} `json:"data"`
	}{
		Page:  page,
		Limit: limit,
		Total: limit * 10,
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

	var txs []*types.Transaction
	for i := 0; i < limit; i++ {
		tx := &types.Transaction{}
		if err := faker.FakeData(&tx); err != nil {
			return err
		}
		txs = append(txs, tx)
	}

	return api.OK.SetData(struct {
		Page  int         `json:"page"`
		Limit int         `json:"limit"`
		Total int         `json:"total"`
		Data  interface{} `json:"data"`
	}{
		Page:  page,
		Limit: limit,
		Total: limit * 25,
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

	return api.OK.SetData(struct {
		Page  int         `json:"page"`
		Limit int         `json:"limit"`
		Total int         `json:"total"`
		Data  interface{} `json:"data"`
	}{
		Page:  page,
		Limit: limit,
		Total: limit * 15,
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
