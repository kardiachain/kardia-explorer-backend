// Package server
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"go.uber.org/zap"

	"github.com/kardiachain/go-kardia"
	"github.com/kardiachain/go-kardia/lib/common"

	"github.com/kardiachain/kardia-explorer-backend/api"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/kardiachain/kardia-explorer-backend/utils"
)

func (s *Server) Ping(c echo.Context) error {
	type pingStat struct {
		Version string `json:"version"`
	}
	stats := &pingStat{Version: cfg.ServerVersion}
	return api.OK.SetData(stats).Build(c)
}

func (s *Server) ServerStatus(c echo.Context) error {
	lgr := s.Logger.With(zap.String("method", "ServerStatus"))
	ctx := context.Background()
	var status *types.ServerStatus
	var err error
	status, err = s.cacheClient.ServerStatus(ctx)
	if err != nil {
		lgr.Warn("Cannot get cache, return default instead")
		status = &types.ServerStatus{
			Status:        "ONLINE",
			AppVersion:    "1.0.0",
			ServerVersion: "1.0.0",
			DexStatus:     "ONLINE",
		}
		// Return default
	}

	return api.OK.SetData(status).Build(c)
}

func (s *Server) UpdateServerStatus(c echo.Context) error {
	lgr := s.Logger.With(zap.String("method", "UpdateServerStatus"))
	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		lgr.Warn("Cannot authorization request")
		return api.Unauthorized.Build(c)
	}
	var serverStatus *types.ServerStatus
	if err := c.Bind(&serverStatus); err != nil {
		lgr.Error("cannot bind server status", zap.Error(err))
		return api.Invalid.Build(c)
	}
	ctx := context.Background()
	if err := s.cacheClient.UpdateServerStatus(ctx, serverStatus); err != nil {
		lgr.Error("cannot update server status", zap.Error(err))
		return api.Invalid.Build(c)
	}

	return api.OK.SetData(nil).Build(c)

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
	totalHolders, totalContracts := s.cacheClient.TotalHolders(ctx)
	return api.OK.SetData(struct {
		TotalHolders   uint64 `json:"totalHolders"`
		TotalContracts uint64 `json:"totalContracts"`
	}{
		TotalHolders:   totalHolders,
		TotalContracts: totalContracts,
	}).Build(c)
}

func (s *Server) Nodes(c echo.Context) error {
	ctx := context.Background()
	nodes, err := s.kaiClient.NodesInfo(ctx)
	if err != nil {
		s.logger.Warn("cannot get nodes info from RPC", zap.Error(err))
		return api.Invalid.Build(c)
	}
	var result []*NodeInfo
	for _, node := range nodes {
		result = append(result, &NodeInfo{
			ID:         node.ID,
			Moniker:    node.Moniker,
			PeersCount: len(node.Peers),
		})
	}
	customNodes, err := s.dbClient.Nodes(ctx)
	if err != nil {
		// If cannot read nodes from db
		// then return network nodes only
		return api.OK.SetData(result).Build(c)
	}

	for _, n := range customNodes {
		result = append(result, &NodeInfo{
			ID:         n.ID,
			Moniker:    n.Moniker,
			PeersCount: len(n.Peers),
		})
	}

	return api.OK.SetData(result).Build(c)
}

func (s *Server) TokenInfo(c echo.Context) error {
	ctx := context.Background()
	if !s.cacheClient.IsRequestToCoinMarket(ctx) {
		tokenInfo, err := s.cacheClient.TokenInfo(ctx)
		if err != nil {
			tokenInfo, err = s.infoServer.TokenInfo(ctx)
			if err != nil {
				return api.Invalid.Build(c)
			}
		}
		cirSup, err := s.kaiClient.GetCirculatingSupply(ctx)
		if err != nil {
			return api.Invalid.Build(c)
		}
		cirSup = new(big.Int).Div(cirSup, new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
		tokenInfo.MainnetCirculatingSupply = cirSup.Int64() - 4500000000
		return api.OK.SetData(tokenInfo).Build(c)
	}

	tokenInfo, err := s.infoServer.TokenInfo(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}
	cirSup, err := s.kaiClient.GetCirculatingSupply(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}
	cirSup = new(big.Int).Div(cirSup, new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	tokenInfo.MainnetCirculatingSupply = cirSup.Int64() - 4500000000
	return api.OK.SetData(tokenInfo).Build(c)
}

func (s *Server) UpdateSupplyAmounts(c echo.Context) error {
	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		return api.Unauthorized.Build(c)
	}
	var supplyInfo *types.SupplyInfo
	if err := c.Bind(&supplyInfo); err != nil {
		return api.Invalid.Build(c)
	}
	if err := s.cacheClient.UpdateSupplyAmounts(ctx, supplyInfo); err != nil {
		return api.Invalid.Build(c)
	}
	return api.OK.Build(c)
}

func (s *Server) GetProposalsList(c echo.Context) error {
	ctx := context.Background()
	pagination, page, limit := getPagingOption(c)
	dbResult, dbTotal, dbErr := s.dbClient.GetListProposals(ctx, pagination)
	if dbErr != nil {
		return api.Invalid.Build(c)
	}
	rpcResult, rpcTotal, rpcErr := s.kaiClient.GetProposals(ctx, pagination)
	if rpcErr != nil {
		fmt.Println("GetProposals err: ", rpcErr)
		return api.Invalid.Build(c)
	}
	if dbTotal != rpcTotal { // try to find out and insert missing proposals to db
		isFound := false
		for _, rpcProposal := range rpcResult {
			isFound = false
			for _, dbProposal := range dbResult {
				if dbProposal.ID == rpcProposal.ID {
					isFound = true
					break
				}
			}
			if isFound {
				continue
			}
			dbResult = append(dbResult, rpcProposal) // include new proposal in response
			s.logger.Info("Inserting new proposal", zap.Any("proposal", rpcProposal))
			err := s.dbClient.UpsertProposal(ctx, rpcProposal) // insert missing proposal to db
			if err != nil {
				s.logger.Debug("Cannot insert new proposal to DB", zap.Error(err))
			}
		}
	}
	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Data:  dbResult,
		Total: rpcTotal,
	}).Build(c)
}

func (s *Server) GetProposalDetails(c echo.Context) error {
	ctx := context.Background()
	proposalID, ok := new(big.Int).SetString(c.Param("id"), 10)
	if !ok {
		return api.Invalid.Build(c)
	}
	result, err := s.dbClient.ProposalInfo(ctx, proposalID.Uint64())
	if err == nil {
		return api.OK.SetData(result).Build(c)
	}
	result, err = s.kaiClient.GetProposalDetails(ctx, proposalID)
	if err != nil {
		fmt.Println("GetProposalDetails err: ", err)
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(result).Build(c)
}

func (s *Server) RemoveDuplicateEvents(c echo.Context) error {
	ctx := context.Background()
	data, err := s.dbClient.RemoveDuplicateEvents(ctx)
	if err != nil {
		return api.InternalServer.Build(c)
	}
	return api.OK.SetData(data).Build(c)
}

func (s *Server) GetParams(c echo.Context) error {
	ctx := context.Background()
	params, err := s.kaiClient.GetParams(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}
	result := make(map[string]interface{})
	for _, param := range params {
		result[param.LabelName] = param.FromValue
	}
	return api.OK.SetData(result).Build(c)
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
			return api.InternalServer.Build(c)
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
	total := s.cacheClient.LatestBlockHeight(ctx)
	return api.OK.SetData(PagingResponse{
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
					return api.Invalid.Build(c)
				}
			}
		}
	} else {
		blockHeight, err := strconv.ParseUint(blockHashOrHeightStr, 10, 64)
		if err != nil || blockHeight <= 0 {
			return api.Invalid.Build(c)
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
					return api.Invalid.Build(c)
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
	return api.OK.SetData(result).Build(c)
}

func (s *Server) PersistentErrorBlocks(c echo.Context) error {
	ctx := context.Background()
	heights, err := s.cacheClient.PersistentErrorBlockHeights(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(heights).Build(c)
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
					return api.InternalServer.Build(c)
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
			return api.Invalid.Build(c)
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
					return api.InternalServer.Build(c)
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

	return api.OK.SetData(PagingResponse{
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
		return api.Invalid.Build(c)
	}

	smcAddress := map[string]*valInfoResponse{}
	validators, err := s.getValidators(ctx)
	if err != nil {
		smcAddress = make(map[string]*valInfoResponse)
		validators = []*types.Validator{}
	}
	//vals, err := s.cacheClient.Validators(ctx)
	//if err != nil {
	//
	//}

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
	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  result,
	}).Build(c)
}

func (s *Server) Txs(c echo.Context) error {
	ctx := context.Background()
	pagination, page, limit := getPagingOption(c)
	var (
		err error
		txs []*types.Transaction
	)

	txs, err = s.cacheClient.LatestTransactions(ctx, pagination)
	if err != nil || txs == nil || len(txs) < limit {
		txs, err = s.dbClient.LatestTxs(ctx, pagination)
		if err != nil {
			return api.Invalid.Build(c)
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

	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: s.cacheClient.TotalTxs(ctx),
		Data:  result,
	}).Build(c)
}

func (s *Server) Addresses(c echo.Context) error {
	ctx := context.Background()
	pagination, page, limit := getPagingOption(c)
	sortDirectionStr := c.QueryParam("sort")
	sortDirection, err := strconv.Atoi(sortDirectionStr)
	if err != nil || (sortDirection != 1 && sortDirection != -1) {
		sortDirection = -1 // DESC
	}
	addrs, err := s.dbClient.GetListAddresses(ctx, sortDirection, pagination)
	if err != nil {
		return api.Invalid.Build(c)
	}
	totalHolders, totalContracts := s.cacheClient.TotalHolders(ctx)
	smcAddress := s.getValidatorsAddressAndRole(ctx)
	var result Addresses
	for _, addr := range addrs {
		addrInfo := SimpleAddress{
			Address:       addr.Address,
			BalanceString: addr.BalanceString,
			IsContract:    addr.IsContract,
			Name:          addr.Name,
			Rank:          addr.Rank,
		}
		if smcAddress[addr.Address] != nil {
			addrInfo.IsInValidatorsList = true
			addrInfo.Role = smcAddress[addr.Address].Role
		}
		// double check with balance from RPC
		balance, err := s.kaiClient.GetBalance(ctx, addr.Address)
		if err != nil {
			return err
		}
		if balance != addr.BalanceString {
			addr.BalanceString = balance
			_ = s.dbClient.UpdateAddresses(ctx, []*types.Address{addr})
		}
		result = append(result, addrInfo)
	}
	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: totalHolders + totalContracts,
		Data:  result,
	}).Build(c)
}

func (s *Server) AddressInfo(c echo.Context) error {
	ctx := context.Background()
	// Convert to addr and get back string to avoid wrong checksum
	address := common.HexToAddress(c.Param("address")).String()
	smcAddress := s.getValidatorsAddressAndRole(ctx)
	addrInfo, err := s.dbClient.AddressByHash(ctx, address)
	if err == nil {
		balance, err := s.kaiClient.GetBalance(ctx, address)
		if err != nil {
			s.logger.Warn("Cannot get address balance from RPC", zap.String("address", address), zap.Error(err))
			return api.Invalid.Build(c)
		}
		code, err := s.kaiClient.GetCode(ctx, address)
		if err != nil {
			s.logger.Warn("Cannot get address code from RPC", zap.String("address", address), zap.Error(err))
			return api.Invalid.Build(c)
		}
		if balance != addrInfo.BalanceString || addrInfo.IsContract != (len(code) > 0) {
			addrInfo.BalanceString = balance
			addrInfo.IsContract = len(code) > 0
			_ = s.dbClient.UpdateAddresses(ctx, []*types.Address{addrInfo})
		}
		result := SimpleAddress{
			Address:       addrInfo.Address,
			BalanceString: addrInfo.BalanceString,
			IsContract:    addrInfo.IsContract,
			Name:          addrInfo.Name,
		}
		if smcAddress[result.Address] != nil {
			result.IsInValidatorsList = true
			result.Role = smcAddress[result.Address].Role
		}
		return api.OK.SetData(result).Build(c)
	}
	s.logger.Warn("address not found in db, getting from RPC instead...", zap.Error(err))
	// try to get balance and code at this address to determine whether we should write this address info to database or not
	newAddr, err := s.newAddressInfo(ctx, address)
	if err != nil {
		return api.Invalid.Build(c)
	}
	result := &SimpleAddress{
		Address:       newAddr.Address,
		BalanceString: newAddr.BalanceString,
		IsContract:    newAddr.IsContract,
	}
	if smcAddress[result.Address] != nil {
		result.IsInValidatorsList = true
		result.Role = smcAddress[result.Address].Role
		result.Name = smcAddress[result.Address].Name
	}
	return api.OK.SetData(result).Build(c)
}

func (s *Server) newAddressInfo(ctx context.Context, address string) (*types.Address, error) {
	balance, err := s.kaiClient.GetBalance(ctx, address)
	if err != nil {
		return nil, err
	}
	balanceInBigInt, _ := new(big.Int).SetString(balance, 10)
	balanceFloat, _ := new(big.Float).SetPrec(100).Quo(new(big.Float).SetInt(balanceInBigInt), new(big.Float).SetInt(cfg.Hydro)).Float64() //converting to KAI from HYDRO
	addrInfo := &types.Address{
		Address:       address,
		BalanceFloat:  balanceFloat,
		BalanceString: balance,
		IsContract:    false,
	}
	code, err := s.kaiClient.GetCode(ctx, address)
	if err == nil && len(code) > 0 {
		addrInfo.IsContract = true
	}
	// write this address to db if its balance is larger than 0 or it's a SMC or it holds KRC token
	tokens, _, _ := s.dbClient.GetListHolders(ctx, &types.HolderFilter{
		HolderAddress: address,
	})
	if balance != "0" || addrInfo.IsContract || len(tokens) > 0 {
		_ = s.dbClient.InsertAddress(ctx, addrInfo) // insert this address to database
	}
	return &types.Address{
		Address:       addrInfo.Address,
		BalanceString: addrInfo.BalanceString,
		IsContract:    addrInfo.IsContract,
	}, nil
}

func (s *Server) AddressTxs(c echo.Context) error {
	ctx := context.Background()
	var err error
	address := c.Param("address")
	pagination, page, limit := getPagingOption(c)

	txs, total, err := s.dbClient.TxsByAddress(ctx, address, pagination)
	if err != nil {
		return err
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

	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  result,
	}).Build(c)
}

func (s *Server) AddressHolders(c echo.Context) error {
	ctx := context.Background()
	var (
		page, limit int
		err         error
	)
	pagination, page, limit := getPagingOption(c)
	filterCrit := &types.HolderFilter{
		Pagination:      pagination,
		ContractAddress: c.QueryParam("contractAddress"),
		HolderAddress:   c.Param("address"),
	}
	holders, total, err := s.dbClient.GetListHolders(ctx, filterCrit)
	if err != nil {
		s.logger.Warn("Cannot get events from db", zap.Error(err))
	}
	for i := range holders {
		holderInfo, _ := s.getAddressInfo(ctx, holders[i].HolderAddress)
		if holderInfo != nil {
			holders[i].HolderName = holderInfo.Name
		}
		krcTokenInfo, _ := s.getKRCTokenInfo(ctx, holders[i].ContractAddress)
		if krcTokenInfo != nil {
			holders[i].Logo = krcTokenInfo.Logo
			// get holder balance from RPC for rechecking
			smcABIStr, err := s.cacheClient.SMCAbi(ctx, cfg.SMCTypePrefix+krcTokenInfo.TokenType)
			if err != nil {
				s.logger.Warn("Cannot get KRC20 token ABI", zap.Error(err), zap.String("smcAddress", holders[i].ContractAddress))
				continue
			}
			smcABI, err := s.decodeSMCABIFromBase64(ctx, smcABIStr, holders[i].ContractAddress)
			if err != nil {
				s.logger.Warn("Cannot decode KRC20 token ABI", zap.Error(err), zap.String("smcAddress", holders[i].ContractAddress))
				continue
			}
			balance, err := s.kaiClient.GetKRC20BalanceByAddress(ctx, smcABI, common.HexToAddress(holders[i].ContractAddress), common.HexToAddress(holders[i].HolderAddress))
			if err != nil {
				s.logger.Warn("Cannot get KRC20 balance of address", zap.Error(err), zap.String("holderAddress", holders[i].HolderAddress),
					zap.String("smcAddress", holders[i].ContractAddress))
				continue
			}
			s.logger.Info("RPC balance vs db balance", zap.String("RPC", balance.String()), zap.String("DB", holders[i].BalanceString))
			if !strings.EqualFold(balance.String(), holders[i].BalanceString) {
				// update correct balance to database and return to client
				holders[i].BalanceString = balance.String()
				holders[i].BalanceFloat = s.calculateKRC20BalanceFloat(balance, krcTokenInfo.Decimals)
				err = s.dbClient.UpdateHolders(ctx, []*types.TokenHolder{holders[i]})
				if err != nil {
					s.logger.Warn("Cannot update KRC20 holder with new balance", zap.Error(err), zap.Any("holder", holders[i]))
				}
			}
		}
	}
	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  holders,
	}).Build(c)
}

func (s *Server) TxByHash(c echo.Context) error {
	ctx := context.Background()
	txHash := c.Param("txHash")
	if txHash == "" {
		return api.Invalid.Build(c)
	}

	receipt, err := s.kaiClient.GetTransactionReceipt(ctx, txHash)
	if err != nil {
		s.Logger.Warn("cannot get receipt by hash from RPC:", zap.String("txHash", txHash))
	}

	var tx *types.Transaction
	tx, err = s.dbClient.TxByHash(ctx, txHash)
	if err != nil {
		// try to get tx by hash through RPC
		s.Logger.Info("cannot get tx by hash from db:", zap.String("txHash", txHash))
		tx, err = s.kaiClient.GetTransaction(ctx, txHash)
		if err != nil {
			s.Logger.Warn("cannot get tx by hash from RPC:", zap.String("txHash", txHash))
			return api.Invalid.Build(c)
		}
		if receipt != nil {
			tx.Logs = receipt.Logs
			tx.Root = receipt.Root
			tx.Status = receipt.Status
			tx.GasUsed = receipt.GasUsed
			tx.ContractAddress = receipt.ContractAddress
			tx.TxFee = new(big.Int).Mul(new(big.Int).SetUint64(tx.GasPrice), new(big.Int).SetUint64(receipt.GasUsed)).String()
		} else { // will be improved later when core blockchain support pending txs API
			if tx.Time.Sub(time.Now()) < 20*time.Second {
				tx.Status = 2 // marked as pending transaction if the duration between now and tx.Time is less than 20 seconds
			} else {
				tx.Status = 0 // marked as failed tx if this tx is submitted for too long
			}
		}
	}

	// Get contract details
	var (
		functionCall *types.FunctionCall
		krcTokenInfo *types.KRCTokenInfo
	)
	smcABI, err := s.getSMCAbi(ctx, &types.Log{
		Address: tx.To,
	})
	if err != nil || smcABI == nil {
		decoded, err := s.kaiClient.DecodeInputData(tx.To, tx.InputData)
		if err == nil {
			functionCall = decoded
		}
	} else {
		decoded, err := s.kaiClient.DecodeInputWithABI(tx.To, tx.InputData, smcABI)
		if err == nil {
			functionCall = decoded
		}
	}

	if functionCall != nil {
		tx.DecodedInputData = functionCall
	}

	filter := &types.InternalTxsFilter{
		TransactionHash: txHash,
	}
	iTxs, _, err := s.dbClient.GetListInternalTxs(ctx, filter)
	if err != nil {
		s.logger.Warn("Cannot get internal transactions from db", zap.Error(err))
	}
	internalTxs := make([]*InternalTransaction, len(iTxs))
	for i := range iTxs {
		logIndex, _ := iTxs[i].LogIndex.(uint)
		internalTxs[i] = &InternalTransaction{
			Log: &types.Log{
				Address:    iTxs[i].Contract,
				MethodName: cfg.KRCTransferMethodName,
				Arguments: map[string]interface{}{
					"from":  iTxs[i].From,
					"to":    iTxs[i].To,
					"value": iTxs[i].Value,
				},
				BlockHeight: iTxs[i].BlockHeight,
				Time:        iTxs[i].Time,
				TxHash:      iTxs[i].TransactionHash,
				Index:       logIndex,
			},
			From:  iTxs[i].From,
			To:    iTxs[i].To,
			Value: iTxs[i].Value,
		}
		fromInfo, _ := s.getAddressInfo(ctx, iTxs[i].From)
		if fromInfo != nil {
			internalTxs[i].FromName = fromInfo.Name
		}
		toInfo, _ := s.getAddressInfo(ctx, iTxs[i].To)
		if toInfo != nil {
			internalTxs[i].ToName = toInfo.Name
		}
		krcTokenInfo, err = s.getKRCTokenInfo(ctx, iTxs[i].Contract)
		if err != nil {
			s.logger.Info("Cannot get KRC Token Info", zap.String("smcAddress", iTxs[i].Contract), zap.Error(err))
			continue
		}
		internalTxs[i].KRCTokenInfo = krcTokenInfo
	}

	result := &Transaction{
		BlockHash:        tx.BlockHash,
		BlockNumber:      tx.BlockNumber,
		Hash:             tx.Hash,
		From:             tx.From,
		To:               tx.To,
		Status:           tx.Status,
		ContractAddress:  tx.ContractAddress,
		Value:            tx.Value,
		GasPrice:         tx.GasPrice,
		GasLimit:         tx.GasLimit,
		GasUsed:          tx.GasUsed,
		TxFee:            tx.TxFee,
		Nonce:            tx.Nonce,
		Time:             tx.Time,
		InputData:        tx.InputData,
		DecodedInputData: tx.DecodedInputData,
		Logs:             internalTxs,
		TransactionIndex: tx.TransactionIndex,
		LogsBloom:        tx.LogsBloom,
		Root:             tx.Root,
	}
	addrInfo, _ := s.getAddressInfo(ctx, tx.From)
	if addrInfo != nil {
		result.FromName = addrInfo.Name
	}
	addrInfo, _ = s.getAddressInfo(ctx, tx.To)
	if addrInfo != nil {
		result.ToName = addrInfo.Name
	}
	smcAddress := s.getValidatorsAddressAndRole(ctx)
	if smcAddress[result.To] != nil {
		result.Role = smcAddress[result.To].Role
		result.IsInValidatorsList = true
		return api.OK.SetData(result).Build(c)
	}
	if result.Status == 0 {
		txTraceResult, err := s.kaiClient.TraceTransaction(ctx, result.Hash)
		if err != nil {
			s.logger.Warn("Cannot trace tx hash", zap.Error(err), zap.String("txHash", result.Hash))
			return api.OK.SetData(result).Build(c)
		}
		result.RevertReason = txTraceResult.RevertReason
	}
	return api.OK.SetData(result).Build(c)
}

//getValidators
func (s *Server) getValidators(ctx context.Context) ([]*types.Validator, error) {
	//validators, err := s.cacheClient.Validators(ctx)
	//if err == nil && len(validators.Validators) != 0 {
	//	return validators, nil
	//}
	// Try from db
	dbValidators, err := s.dbClient.Validators(ctx, db.ValidatorsFilter{})
	if err == nil {
		//s.logger.Debug("get validators from storage", zap.Any("Validators", dbValidators))
		stats, err := s.CalculateValidatorStats(ctx, dbValidators)
		if err == nil && len(dbValidators) != 0 {
			s.logger.Debug("stats ", zap.Any("stats", stats))
		}
		return dbValidators, nil
		//return dbValidators, nil
	}

	s.logger.Debug("Load validator from network")
	validators, err := s.kaiClient.Validators(ctx)
	if err != nil {
		s.logger.Warn("cannot get validators list from RPC", zap.Error(err))
		return nil, err
	}
	//err = s.cacheClient.UpdateValidators(ctx, vasList)
	//if err != nil {
	//	s.logger.Warn("cannot store validators list to cache", zap.Error(err))
	//}
	return validators, nil
}

func (s *Server) CalculateValidatorStats(ctx context.Context, validators []*types.Validator) (*types.ValidatorStats, error) {
	var stats types.ValidatorStats
	var (
		ErrParsingBigIntFromString = errors.New("cannot parse big.Int from string")
		proposersStakedAmount      = big.NewInt(0)
		delegatorsMap              = make(map[string]bool)
		totalProposers             = 0
		totalValidators            = 0
		totalCandidates            = 0
		totalStakedAmount          = big.NewInt(0)
		totalDelegatorStakedAmount = big.NewInt(0)
		totalDelegators            = 0

		valStakedAmount *big.Int
		delStakedAmount *big.Int
		ok              bool
	)
	for _, val := range validators {
		// Calculate total staked amount
		valStakedAmount, ok = new(big.Int).SetString(val.StakedAmount, 10)
		if !ok {
			return nil, ErrParsingBigIntFromString
		}
		totalStakedAmount = new(big.Int).Add(totalStakedAmount, valStakedAmount)

		for _, d := range val.Delegators {
			if !delegatorsMap[d.Address] {
				delegatorsMap[d.Address] = true
				totalDelegators++
			}
			delStakedAmount, ok = new(big.Int).SetString(d.StakedAmount, 10)
			if !ok {
				return nil, ErrParsingBigIntFromString
			}
			if d.Address == val.Address {
				proposersStakedAmount = new(big.Int).Add(proposersStakedAmount, delStakedAmount)
			} else {

				totalDelegatorStakedAmount = new(big.Int).Add(totalDelegatorStakedAmount, delStakedAmount)
			}
		}
		//val.Role = ec.getValidatorRole(valsSet, val.Address, val.Status)
		// validator who started a node and not in validators set is a normal validator
		if val.Role == 2 {
			totalProposers++
			totalValidators++
		} else if val.Role == 1 {
			totalValidators++
		} else if val.Role == 0 {
			totalCandidates++
		}
	}
	stats.TotalStakedAmount = totalStakedAmount.String()
	stats.TotalDelegatorStakedAmount = totalDelegatorStakedAmount.String()
	stats.TotalValidatorStakedAmount = proposersStakedAmount.String()
	stats.TotalDelegators = totalDelegators
	stats.TotalCandidates = totalCandidates
	stats.TotalValidators = totalValidators
	stats.TotalProposers = totalProposers
	return &stats, nil
}

func getPagingOption(c echo.Context) (*types.Pagination, int, int) {
	pageParams := c.QueryParam("page")
	limitParams := c.QueryParam("limit")
	if pageParams == "" && limitParams == "" {
		return nil, 0, 0
	}
	page, err := strconv.Atoi(pageParams)
	if err != nil {
		page = 1
	}
	page = page - 1
	limit, err := strconv.Atoi(limitParams)
	if err != nil {
		limit = 25
	}
	pagination := &types.Pagination{
		Skip:  page * limit,
		Limit: limit,
	}
	pagination.Sanitize()
	return pagination, page + 1, limit
}

func (s *Server) getValidatorsAddressAndRole(ctx context.Context) map[string]*valInfoResponse {
	validators, err := s.getValidators(ctx)
	if err != nil {
		return make(map[string]*valInfoResponse)
	}

	smcAddress := map[string]*valInfoResponse{}
	for _, v := range validators {
		smcAddress[v.SmcAddress] = &valInfoResponse{
			Name: v.Name,
			Role: v.Role,
		}
	}
	return smcAddress
}

func (s *Server) UpsertNetworkNodes(c echo.Context) error {
	//ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		return api.Unauthorized.Build(c)
	}
	var nodeInfo *types.NodeInfo
	if err := c.Bind(&nodeInfo); err != nil {
		return api.Invalid.Build(c)
	}
	if nodeInfo.ID == "" || nodeInfo.Moniker == "" {
		return api.Invalid.Build(c)
	}
	ctx := context.Background()
	if err := s.dbClient.UpsertNode(ctx, nodeInfo); err != nil {
		return api.InternalServer.Build(c)
	}

	return api.OK.Build(c)
}

func (s *Server) RemoveNetworkNodes(c echo.Context) error {
	//ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		return api.Unauthorized.Build(c)
	}
	nodesID := c.Param("nodeID")
	if nodesID == "" {
		return api.Invalid.Build(c)
	}

	ctx := context.Background()
	if err := s.dbClient.RemoveNode(ctx, nodesID); err != nil {
		return api.InternalServer.Build(c)
	}

	return api.OK.Build(c)
}

func (s *Server) ReloadAddressesBalance(c echo.Context) error {
	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		return api.Unauthorized.Build(c)
	}

	addresses, err := s.dbClient.Addresses(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}

	for id, a := range addresses {
		balance, err := s.kaiClient.GetBalance(ctx, a.Address)
		if err != nil {
			continue
		}
		addresses[id].BalanceString = balance
	}

	if err := s.dbClient.UpdateAddresses(ctx, addresses); err != nil {
		return api.Invalid.Build(c)
	}

	return api.OK.Build(c)
}

func (s *Server) UpdateAddressName(c echo.Context) error {
	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		return api.Unauthorized.Build(c)
	}
	var addressName types.UpdateAddress
	if err := c.Bind(&addressName); err != nil {
		fmt.Println("cannot bind ", err)
		return api.Invalid.Build(c)
	}
	addressInfo, err := s.dbClient.AddressByHash(ctx, addressName.Address)
	if err != nil {
		return api.Invalid.Build(c)
	}

	addressInfo.Name = addressName.Name

	if err := s.dbClient.UpdateAddresses(ctx, []*types.Address{addressInfo}); err != nil {
		fmt.Println("cannot update ", err)
		return api.Invalid.Build(c)
	}
	_ = s.cacheClient.UpdateAddressInfo(ctx, addressInfo)
	return api.OK.Build(c)
}

func (s *Server) ReloadValidators(c echo.Context) error {
	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		return api.Unauthorized.Build(c)
	}

	//todo longnd: rework reload validator API
	//validators, err := s.kaiClient.Validators(ctx)
	//if err != nil {
	//	return api.Invalid.Build(c)
	//}
	//
	//if err := s.dbClient.UpsertValidators(ctx, validators); err != nil {
	//	return api.Invalid.Build(c)
	//}

	return api.OK.Build(c)
}

func (s *Server) ContractEvents(c echo.Context) error {
	ctx := context.Background()
	var (
		page, limit  int
		err          error
		krcTokenInfo *types.KRCTokenInfo
		events       []*types.Log
		total        uint64
	)
	pagination, page, limit := getPagingOption(c)
	filter := &types.EventsFilter{
		Pagination:      pagination,
		ContractAddress: c.QueryParam("contractAddress"),
		MethodName:      c.QueryParam("methodName"),
		TxHash:          c.QueryParam("txHash"),
	}
	if filter.MethodName != "" || filter.ContractAddress != "" {
		events, total, err = s.dbClient.GetListEvents(ctx, filter)
		if err != nil {
			s.logger.Warn("Cannot get events from db", zap.Error(err))
		}
	} else {
		receipt, err := s.kaiClient.GetTransactionReceipt(ctx, filter.TxHash)
		if err != nil {
			s.logger.Warn("Cannot get receipt from RPC", zap.Error(err))
			return api.Invalid.Build(c)
		}
		for i := range receipt.Logs {
			smcABI, err := s.getSMCAbi(ctx, &receipt.Logs[i])
			if err != nil {
				events = append(events, &receipt.Logs[i])
				continue
			}
			unpackedLog, err := s.kaiClient.UnpackLog(&receipt.Logs[i], smcABI)
			if err != nil {
				events = append(events, &receipt.Logs[i])
				continue
			}
			events = append(events, unpackedLog)
		}
	}
	result := make([]*InternalTransaction, len(events))
	for i := range events {
		krcTokenInfo, err = s.getKRCTokenInfo(ctx, events[i].Address)
		if err != nil {
			continue
		}
		result[i] = &InternalTransaction{
			Log:          events[i],
			KRCTokenInfo: krcTokenInfo,
		}
	}
	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  result,
	}).Build(c)
}

func (s *Server) Contracts(c echo.Context) error {
	ctx := context.Background()
	pagination, page, limit := getPagingOption(c)
	// default filter
	filterCrit := &types.ContractsFilter{
		Type:       c.QueryParam("type"),
		Pagination: pagination,
		Status:     c.QueryParam("status"),
	}

	result, total, err := s.dbClient.Contracts(ctx, filterCrit)
	if err != nil {
		return api.Invalid.Build(c)
	}

	finalResult := make([]*SimpleKRCTokenInfo, len(result))
	for i := range result {
		finalResult[i] = &SimpleKRCTokenInfo{
			Name:       result[i].Name,
			Address:    result[i].Address,
			Info:       result[i].Info,
			Type:       result[i].Type,
			Logo:       result[i].Logo,
			IsVerified: result[i].IsVerified,
		}
		tokenInfo, err := s.getKRCTokenInfo(ctx, result[i].Address)
		if err != nil {
			continue
		}
		finalResult[i].TokenSymbol = tokenInfo.TokenSymbol
		finalResult[i].TotalSupply = tokenInfo.TotalSupply
		finalResult[i].Decimal = tokenInfo.Decimals
	}
	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  finalResult,
	}).Build(c)
}

func (s *Server) Contract(c echo.Context) error {
	ctx := context.Background()
	smc, addrInfo, err := s.dbClient.Contract(ctx, c.Param("contractAddress"))
	if err != nil {
		return api.Invalid.Build(c)
	}
	if smc.ABI == "" && smc.Type != "" {
		abiStr, err := s.cacheClient.SMCAbi(ctx, cfg.SMCTypePrefix+smc.Type)
		if err != nil {
			abiStr, err = s.dbClient.SMCABIByType(ctx, smc.Type)
			if err != nil {
				s.logger.Warn("Cannot get ABI by type from db", zap.String("type", smc.Type), zap.Error(err))
			}
		}
		// update KRC token total supply from RPC
		if abiStr != "" {
			smc.ABI = abiStr
			krcTokenInfoRPC, err := s.getKRCTokenInfo(ctx, smc.Type)
			if err != nil {
				s.logger.Warn("New contract is not a KRC token", zap.Error(err), zap.Any("tokenInfo", krcTokenInfoRPC))
			} else {
				addrInfo.TotalSupply = krcTokenInfoRPC.TotalSupply
			}
		}
	}
	var result *KRCTokenInfo
	// map smc info to result
	smcJSON, err := json.Marshal(smc)
	if err != nil {
		return api.Invalid.Build(c)
	}
	err = json.Unmarshal(smcJSON, &result)
	if err != nil {
		return api.Invalid.Build(c)
	}
	// map address info to result
	addrInfoJSON, err := json.Marshal(addrInfo)
	if err != nil {
		return api.Invalid.Build(c)
	}
	err = json.Unmarshal(addrInfoJSON, &result)
	if err != nil {
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(result).Build(c)
}

func (s *Server) InsertContract(c echo.Context) error {
	lgr := s.logger.With(zap.String("method", "InsertContract"))

	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		return api.Unauthorized.Build(c)
	}

	lgr.Debug("Start insert contract")
	var (
		contract     types.Contract
		addrInfo     types.Address
		bodyBytes, _ = ioutil.ReadAll(c.Request().Body)
	)
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	if err := c.Bind(&contract); err != nil {
		lgr.Error("cannot bind data", zap.Error(err))
		return api.Invalid.Build(c)
	}
	contract.IsVerified = true
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	if err := c.Bind(&addrInfo); err != nil {
		lgr.Error("cannot bind data", zap.Error(err))
		return api.Invalid.Build(c)
	}
	ctx := context.Background()
	krcTokenInfoFromRPC, err := s.getKRCTokenInfoFromRPC(ctx, addrInfo.Address, addrInfo.KrcTypes)
	if err != nil && strings.HasPrefix(addrInfo.KrcTypes, "KRC") {
		s.logger.Warn("Updating contract is not KRC type", zap.Any("smcInfo", addrInfo))
		return api.Invalid.Build(c)
	}
	if krcTokenInfoFromRPC != nil {
		// cache new token info
		krcTokenInfoFromRPC.Logo = addrInfo.Logo
		if utils.CheckBase64Logo(addrInfo.Logo) {
			fileName, err := s.fileStorage.UploadLogo(addrInfo.Logo, utils.HashString(contract.Address), s.ConfigUploader)
			if err != nil {
				log.Fatal("Error when upload the image: ", err)
			} else {
				addrInfo.Logo = fileName
				contract.Logo = fileName
			}
		}
		_ = s.cacheClient.UpdateKRCTokenInfo(ctx, krcTokenInfoFromRPC)

		addrInfo.TokenName = krcTokenInfoFromRPC.TokenName
		addrInfo.TokenSymbol = krcTokenInfoFromRPC.TokenSymbol
		addrInfo.TotalSupply = krcTokenInfoFromRPC.TokenName
		addrInfo.Decimals = krcTokenInfoFromRPC.Decimals
	}
	if err := s.dbClient.InsertContract(ctx, &contract, &addrInfo); err != nil {
		lgr.Error("cannot bind insert", zap.Error(err))
		return api.InternalServer.Build(c)
	}

	return api.OK.Build(c)
}

func (s *Server) VerifyContract(ctx context.Context) error {
	lgr := s.logger.With(zap.String("method", "VerifyContract"))
	lgr.Debug("Start verify contract data")
	type verifyRequest struct {
		SMCAddress string `json:"smc_address"`
		Code       string `json:"code"`
	}
	var req verifyRequest

	code, err := s.kaiClient.GetCode(ctx, req.SMCAddress)
	if err != nil {
		return err
	}

	if req.Code != string(code) {
		return err
	}

	return nil
}

func (s *Server) UpdateContract(c echo.Context) error {
	lgr := s.logger.With(zap.String("method", "UpdateContract"))

	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		return api.Unauthorized.Build(c)
	}

	lgr.Debug("Start insert contract")
	var (
		contract     types.Contract
		addrInfo     types.Address
		bodyBytes, _ = ioutil.ReadAll(c.Request().Body)
	)
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	if err := c.Bind(&contract); err != nil {
		lgr.Error("cannot bind contract data", zap.Error(err))
		return api.Invalid.Build(c)
	}
	contract.IsVerified = true
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	if err := c.Bind(&addrInfo); err != nil {
		lgr.Error("cannot bind address data", zap.Error(err))
		return api.Invalid.Build(c)
	}
	ctx := context.Background()
	krcTokenInfoFromRPC, err := s.getKRCTokenInfoFromRPC(ctx, addrInfo.Address, addrInfo.KrcTypes)
	if err != nil && strings.HasPrefix(addrInfo.KrcTypes, "KRC") {
		s.logger.Warn("Updating contract is not KRC type", zap.Any("smcInfo", addrInfo), zap.Error(err))
		return api.Invalid.Build(c)
	}
	if krcTokenInfoFromRPC != nil {
		// cache new token info
		krcTokenInfoFromRPC.Logo = addrInfo.Logo

		if utils.CheckBase64Logo(addrInfo.Logo) {
			fileName, err := s.fileStorage.UploadLogo(addrInfo.Logo, utils.HashString(contract.Address), s.ConfigUploader)
			if err != nil {
				log.Fatal("Error when upload the image: ", err)
			} else {
				addrInfo.Logo = fileName
				contract.Logo = fileName
			}
		}

		_ = s.cacheClient.UpdateKRCTokenInfo(ctx, krcTokenInfoFromRPC)

		addrInfo.TokenName = krcTokenInfoFromRPC.TokenName
		addrInfo.TokenSymbol = krcTokenInfoFromRPC.TokenSymbol
		addrInfo.TotalSupply = krcTokenInfoFromRPC.TotalSupply
		addrInfo.Decimals = krcTokenInfoFromRPC.Decimals
	}
	if err := s.dbClient.UpdateContract(ctx, &contract, &addrInfo); err != nil {
		lgr.Error("cannot bind insert", zap.Error(err))
		return api.InternalServer.Build(c)
	}

	return api.OK.SetData(addrInfo).Build(c)
}

func (s *Server) UpdateSMCABIByType(c echo.Context) error {
	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		return api.Unauthorized.Build(c)
	}
	ctx := context.Background()
	var smcABI *types.ContractABI
	if err := c.Bind(&smcABI); err != nil {
		return api.Invalid.Build(c)
	}
	err := s.dbClient.UpsertSMCABIByType(ctx, smcABI.Type, smcABI.ABI)
	if err != nil {
		return api.Invalid.Build(c)
	}
	return api.OK.Build(c)
}

func (s *Server) SearchAddressByName(c echo.Context) error {
	var (
		ctx     = context.Background()
		name    = c.QueryParam("name")
		addrMap = make(map[string]*SimpleKRCTokenInfo)
	)
	if name == "" {
		return api.Invalid.Build(c)
	}

	addresses, err := s.dbClient.AddressByName(ctx, name)
	if err == nil && len(addresses) > 0 {
		for _, addr := range addresses {
			addrMap[addr.Address] = &SimpleKRCTokenInfo{
				Name:        addr.Name,
				Address:     addr.Address,
				Info:        addr.Info,
				Type:        "Address",
				TokenSymbol: addr.TokenSymbol,
			}
		}
	}
	s.logger.Info("Addresses search results", zap.Any("addresses", addresses), zap.Error(err))

	contracts, err := s.dbClient.ContractByName(ctx, name)
	if err == nil && len(contracts) > 0 {
		for _, smc := range contracts {
			if smc.Type == "" {
				smc.Type = "SMC"
			}
			if addrMap[smc.Address] != nil {
				addrMap[smc.Address].Type = smc.Type
				addrMap[smc.Address].Name = smc.Name
				addrMap[smc.Address].Logo = smc.Logo
			} else {
				addrMap[smc.Address] = &SimpleKRCTokenInfo{
					Name:    smc.Name,
					Address: smc.Address,
					Info:    smc.Info,
					Logo:    smc.Logo,
					Type:    smc.Type,
				}
			}
		}
	}
	s.logger.Info("Contracts search results", zap.Any("contracts", contracts), zap.Error(err))

	if len(addrMap) == 0 {
		return api.Invalid.Build(c)
	}
	result := make([]*SimpleKRCTokenInfo, 0, len(addrMap))
	for _, addr := range addrMap {
		result = append(result, addr)
	}
	return api.OK.SetData(result).Build(c)
}

func (s *Server) GetHoldersListByToken(c echo.Context) error {
	ctx := context.Background()
	var (
		page, limit int
		err         error
	)
	pagination, page, limit := getPagingOption(c)
	filterCrit := &types.HolderFilter{
		Pagination:      pagination,
		ContractAddress: c.Param("contractAddress"),
	}
	holders, total, err := s.dbClient.GetListHolders(ctx, filterCrit)
	if err != nil {
		s.logger.Warn("Cannot get events from db", zap.Error(err))
	}
	krcTokenInfo, _ := s.getKRCTokenInfo(ctx, c.Param("contractAddress"))
	if krcTokenInfo != nil {
		for i := range holders {
			holders[i].Logo = krcTokenInfo.Logo
			// remove redundant field
			holders[i].TokenName = ""
			holders[i].TokenSymbol = ""
			holders[i].Logo = ""
			holders[i].ContractAddress = ""
			// add address names
			holderInfo, _ := s.getAddressInfo(ctx, holders[i].HolderAddress)
			if holderInfo != nil {
				holders[i].HolderName = holderInfo.Name
			}
		}
	}
	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  holders,
	}).Build(c)
}

func (s *Server) GetInternalTxs(c echo.Context) error {
	ctx := context.Background()
	var (
		page, limit int
		err         error
	)
	pagination, page, limit := getPagingOption(c)
	filterCrit := &types.InternalTxsFilter{
		Pagination:      pagination,
		Contract:        c.QueryParam("contractAddress"),
		Address:         c.QueryParam("address"),
		TransactionHash: c.QueryParam("txHash"),
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
		fromInfo, _ = s.getAddressInfo(ctx, iTxs[i].From)
		if fromInfo != nil {
			result[i].FromName = fromInfo.Name
		}
		toInfo, _ = s.getAddressInfo(ctx, iTxs[i].To)
		if toInfo != nil {
			result[i].ToName = toInfo.Name
		}
		krcTokenInfo, _ := s.getKRCTokenInfo(ctx, iTxs[i].Contract)
		if krcTokenInfo != nil {
			result[i].KRCTokenInfo = krcTokenInfo
		}
	}
	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  result,
	}).Build(c)
}

func (s *Server) UpdateInternalTxs(c echo.Context) error {
	var (
		ctx             = context.Background()
		crit            *types.TxsFilter
		internalTxsCrit *types.InternalTxsFilter
		lgr             = s.logger.With(zap.String("api", "UpdateInternalTxs"))
		bodyBytes, _    = ioutil.ReadAll(c.Request().Body)
	)
	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		lgr.Warn("Cannot authorization request")
		return api.Unauthorized.Build(c)
	}
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	if err := c.Bind(&crit); err != nil {
		lgr.Error("cannot bind txs filter", zap.Error(err))
		return api.Invalid.Build(c)
	}
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	if err := c.Bind(&internalTxsCrit); err != nil {
		lgr.Error("cannot bind internal txs filter", zap.Error(err))
		return api.Invalid.Build(c)
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
			return api.Invalid.Build(c)
		}
		lgr.Info("Filtering events", zap.Uint64("latestBlockHeight", latestBlockHeight), zap.Uint64("from", fromBlock), zap.Uint64("to", toBlock),
			zap.Any("criteria", criteria), zap.Int("number of logs", len(logs)))
	} else {
		// find the block height where this contract is deployed
		txs, _, err := s.dbClient.FilterTxs(ctx, crit)
		lgr.Info("UpdateInternalTxs", zap.Any("criteria", crit))
		if err != nil || len(txs) == 0 {
			lgr.Error("Cannot get the transaction where this contract was deployed", zap.Error(err))
			return api.Invalid.Build(c)
		}

		lgr.Info("Filtering events", zap.Uint64("from", txs[0].BlockNumber), zap.Uint64("to", latestBlockHeight))
		for i := txs[0].BlockNumber; i < latestBlockHeight; i += cfg.FilterLogsInterval {
			latestBlockHeight, err = s.kaiClient.LatestBlockNumber(ctx)
			if err != nil {
				lgr.Error("Cannot get latest block height from RPC", zap.Error(err), zap.Any("criteria", crit))
				return api.Invalid.Build(c)
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
		return api.Invalid.Build(c)
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
			return api.Invalid.Build(c)
		}
	}

	// batch inserting to InternalTransactions collection in db
	lgr.Info("internalTxs ready to be batch inserted", zap.Any("iTxs", len(internalTxs)), zap.Any("logs", len(logs)))
	if err = s.dbClient.UpdateInternalTxs(ctx, internalTxs); err != nil {
		lgr.Error("Cannot batch inserting new internal txs in db", zap.Error(err))
		return api.Invalid.Build(c)
	}
	return api.OK.Build(c)
}

func (s *Server) getAddressInfo(ctx context.Context, address string) (*types.Address, error) {
	addrInfo, err := s.cacheClient.AddressInfo(ctx, address)
	if err == nil {
		return addrInfo, nil
	}
	s.logger.Info("Cannot get address info in cache, getting from db instead", zap.String("address", address), zap.Error(err))
	addrInfo, err = s.dbClient.AddressByHash(ctx, address)
	if err != nil {
		s.logger.Warn("Cannot get address info from db", zap.String("address", address), zap.Error(err))
		if err != nil {
			// insert new address to db
			newAddr, err := s.newAddressInfo(ctx, address)
			if err != nil {
				s.logger.Warn("Cannot store address info to db", zap.Any("address", newAddr), zap.Error(err))
			}
		}
		return nil, err
	}
	err = s.cacheClient.UpdateAddressInfo(ctx, addrInfo)
	if err != nil {
		s.logger.Warn("Cannot store address info to cache", zap.String("address", address), zap.Error(err))
	}
	return addrInfo, nil
}
