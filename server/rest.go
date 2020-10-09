// Package server
package server

import (
	"strconv"

	"github.com/bxcodec/faker/v3"
	"github.com/labstack/echo"

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
	return api.OK.Build(c)
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
	return api.OK.Build(c)
}

func (s *Server) Blocks(c echo.Context) error {
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

	var blocks []*types.Block

	for i := 0; i < limit; i++ {
		block := &types.Block{}
		if err := faker.FakeData(&block); err != nil {
			return err
		}
		blocks = append(blocks, block)
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
	block := &types.Block{}
	if err := faker.FakeData(&block); err != nil {
		return api.InternalServer.Build(c)
	}
	return api.OK.SetData(block).Build(c)
}

func (s *Server) BlockExist(c echo.Context) error {

	return api.OK.Build(c)
}

func (s *Server) BlockTxs(c echo.Context) error {
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
	// Random number of txs of block hash
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
		Limit int         `json:limit"`
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
	tx := &types.Transaction{}
	if err := faker.FakeData(&tx); err != nil {
		return err
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
