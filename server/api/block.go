// Package api
package api

import (
	"context"
	"strconv"
	"strings"

	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

type IBlock interface {
	Blocks(c echo.Context) error
	Block(c echo.Context) error
	BlockTxs(c echo.Context) error
	BlocksByProposer(c echo.Context) error
}

func bindBlocksAPIs(gr *echo.Group, srv RestServer) {
	apis := []restDefinition{
		{
			method: echo.GET,
			// Query params: ?page=0&limit=10
			path: "/blocks",
			fn:   srv.Blocks,
		},
		{
			method: echo.GET,
			path:   "/blocks/:block",
			fn:     srv.Block,
		},
		{
			method: echo.GET,
			// Params: proposer address
			// Query params: ?page=0&limit=10
			path:        "/blocks/proposer/:address",
			fn:          srv.BlocksByProposer,
			middlewares: nil,
		},
		{
			method: echo.GET,
			// Params: block's hash
			// Query params: ?page=0&limit=10
			path:        "/block/:block/txs",
			fn:          srv.BlockTxs,
			middlewares: []echo.MiddlewareFunc{checkPagination()},
		},
	}
	for _, api := range apis {
		gr.Add(api.method, api.path, api.fn, api.middlewares...)
	}
}

func (s *Server) Blocks(c echo.Context) error {
	ctx := context.Background()
	var (
		err    error
		blocks []*types.Block
	)
	pagination, page, limit := getPagingOption(c)

	blocks, err = s.cacheClient.LatestBlocks(ctx, pagination)
	if err != nil || blocks == nil {
		blocks, err = s.dbClient.Blocks(ctx, pagination)
		if err != nil {
			s.logger.Info("Cannot get latest blocks from db", zap.Error(err))
			return InternalServer.Build(c)
		}
	}

	smcAddress := map[string]*valInfoResponse{}
	vals, err := s.getValidators(ctx)
	if err != nil {
		smcAddress = make(map[string]*valInfoResponse)
		vals = []*types.Validator{}
	}

	for _, v := range vals {
		smcAddress[v.Address] = &valInfoResponse{
			Name: v.Name,
			Role: v.Role,
		}
	}
	s.logger.Debug("List validators", zap.Any("smc", smcAddress))
	var result Blocks
	for _, block := range blocks {
		b := SimpleBlock{
			Height:          block.Height,
			Hash:            block.Hash,
			Time:            block.Time,
			ProposerAddress: common.HexToAddress(block.ProposerAddress).String(),
			NumTxs:          block.NumTxs,
			GasLimit:        block.GasLimit,
			GasUsed:         block.GasUsed,
			Rewards:         block.Rewards,
		}
		p, ok := smcAddress[b.ProposerAddress]
		if ok && p != nil {
			b.ProposerName = smcAddress[b.ProposerAddress].Name
		}
		result = append(result, b)
	}
	total := s.cacheClient.LatestBlockHeight(ctx)
	return OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Data:  result,
		Total: total,
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
		// get block in cache if exist
		block, err = s.cacheClient.BlockByHash(ctx, blockHashOrHeightStr)
		if err != nil {
			// otherwise, get from db
			block, err = s.dbClient.BlockByHash(ctx, blockHashOrHeightStr)
			if err != nil {
				s.logger.Warn("got block by hash from db error", zap.Any("block", block), zap.Error(err))
				// try to get from RPC at last
				block, err = s.kaiClient.BlockByHash(ctx, blockHashOrHeightStr)
				if err != nil {
					s.logger.Warn("got block by hash from RPC error", zap.Any("block", block), zap.Error(err))
					return Invalid.Build(c)
				}
			}
		}
	} else {
		blockHeight, err := strconv.ParseUint(blockHashOrHeightStr, 10, 64)
		if err != nil || blockHeight <= 0 {
			return Invalid.Build(c)
		}
		// get block in cache if exist
		block, err = s.cacheClient.BlockByHeight(ctx, blockHeight)
		if err != nil {
			// otherwise, get from db
			block, err = s.dbClient.BlockByHeight(ctx, blockHeight)
			if err != nil {
				s.logger.Warn("got block by height from db error", zap.Uint64("blockHeight", blockHeight), zap.Error(err))
				// try to get from RPC at last
				block, err = s.kaiClient.BlockByHeight(ctx, blockHeight)
				if err != nil {
					s.logger.Warn("got block by height from RPC error", zap.Uint64("blockHeight", blockHeight), zap.Error(err))
					return Invalid.Build(c)
				}
			}
		}
	}

	smcAddress := map[string]*valInfoResponse{}
	validators, err := s.getValidators(ctx)
	if err != nil {
		smcAddress = make(map[string]*valInfoResponse)
		validators = []*types.Validator{}
	}
	for _, v := range validators {
		smcAddress[v.Address] = &valInfoResponse{
			Name: v.Name,
			Role: v.Role,
		}
	}
	var proposerName string
	p, ok := smcAddress[block.ProposerAddress]
	if ok && p != nil {
		proposerName = p.Name
	}

	result := &Block{
		Block:        *block,
		ProposerName: proposerName,
	}
	return OK.SetData(result).Build(c)
}

func (s *Server) BlockTxs(c echo.Context) error {
	ctx := context.Background()
	block := c.Param("block")
	pagination, page, limit := getPagingOption(c)

	var (
		txs   []*types.Transaction
		total uint64
		err   error
	)

	if strings.HasPrefix(block, "0x") {
		// get block txs in block if exist
		txs, total, err = s.cacheClient.TxsByBlockHash(ctx, block, pagination)
		if err != nil {
			// otherwise, get from db
			txs, total, err = s.dbClient.TxsByBlockHash(ctx, block, pagination)
			if err != nil {
				s.logger.Warn("cannot get block txs by hash from db", zap.String("blockHash", block), zap.Error(err))
				// try to get block txs from RPC
				blockRPC, err := s.kaiClient.BlockByHash(ctx, block)
				if err != nil {
					s.logger.Warn("cannot get block txs by hash from RPC", zap.String("blockHash", block), zap.Error(err))
					return InternalServer.Build(c)
				}
				txs = blockRPC.Txs
				if pagination.Skip > len(txs) {
					txs = []*types.Transaction(nil)
				} else if pagination.Skip+pagination.Limit > len(txs) {
					txs = blockRPC.Txs[pagination.Skip:len(txs)]
				} else {
					txs = blockRPC.Txs[pagination.Skip : pagination.Skip+pagination.Limit]
				}
				total = blockRPC.NumTxs
			}
		}
	} else {
		height, err := strconv.ParseUint(block, 10, 64)
		if err != nil || height <= 0 {
			return Invalid.Build(c)
		}
		// get block txs in block if exist
		txs, total, err = s.cacheClient.TxsByBlockHeight(ctx, height, pagination)
		if err != nil {
			// otherwise, get from db
			txs, total, err = s.dbClient.TxsByBlockHeight(ctx, height, pagination)
			if err != nil {
				s.logger.Warn("cannot get block txs by height from db", zap.String("blockHeight", block), zap.Error(err))
				// try to get block txs from RPC
				blockRPC, err := s.kaiClient.BlockByHeight(ctx, height)
				if err != nil {
					s.logger.Warn("cannot get block txs by height from RPC", zap.String("blockHeight", block), zap.Error(err))
					return InternalServer.Build(c)
				}
				txs = blockRPC.Txs
				if pagination.Skip > len(txs) {
					txs = []*types.Transaction(nil)
				} else if pagination.Skip+pagination.Limit > len(txs) {
					txs = blockRPC.Txs[pagination.Skip:len(txs)]
				} else {
					txs = blockRPC.Txs[pagination.Skip : pagination.Skip+pagination.Limit]
				}
				total = blockRPC.NumTxs
			}
		}
	}

	smcAddress := s.getValidatorsAddressAndRole(ctx)
	var result Transactions
	for _, tx := range txs {
		t := SimpleTransaction{
			Hash:             tx.Hash,
			BlockNumber:      tx.BlockNumber,
			Time:             tx.Time,
			From:             tx.From,
			To:               tx.To,
			ContractAddress:  tx.ContractAddress,
			Value:            tx.Value,
			TxFee:            tx.TxFee,
			Status:           tx.Status,
			DecodedInputData: tx.DecodedInputData,
			InputData:        tx.InputData,
		}
		if smcAddress[tx.To] != nil {
			t.Role = smcAddress[tx.To].Role
			t.IsInValidatorsList = true
		}
		addrInfo, _ := s.getAddressInfo(ctx, tx.From)
		if addrInfo != nil {
			t.FromName = addrInfo.Name
		}
		addrInfo, _ = s.getAddressInfo(ctx, tx.To)
		if addrInfo != nil {
			t.ToName = addrInfo.Name
		}
		result = append(result, t)
	}

	return OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  result,
	}).Build(c)
}

func (s *Server) BlocksByProposer(c echo.Context) error {
	ctx := context.Background()
	pagination, page, limit := getPagingOption(c)
	blocks, total, err := s.dbClient.BlocksByProposer(ctx, c.Param("address"), pagination)
	if err != nil {
		return Invalid.Build(c)
	}

	smcAddress := map[string]*valInfoResponse{}
	validators, err := s.getValidators(ctx)
	if err != nil {
		smcAddress = make(map[string]*valInfoResponse)
		validators = []*types.Validator{}
	}

	for _, v := range validators {
		smcAddress[v.Address] = &valInfoResponse{
			Name: v.Name,
			Role: v.Role,
		}
	}
	var result Blocks
	for _, block := range blocks {
		b := SimpleBlock{
			Height:          block.Height,
			Hash:            block.Hash,
			Time:            block.Time,
			ProposerAddress: block.ProposerAddress,
			NumTxs:          block.NumTxs,
			GasLimit:        block.GasLimit,
			GasUsed:         block.GasUsed,
			Rewards:         block.Rewards,
		}

		p, ok := smcAddress[b.ProposerAddress]
		if ok && p != nil {
			b.ProposerName = smcAddress[b.ProposerAddress].Name
		}

		result = append(result, b)
	}
	return OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  result,
	}).Build(c)
}
