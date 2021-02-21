// Package server
package server

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/kardiachain/go-kardia/lib/common"

	"github.com/bxcodec/faker/v3"
	"github.com/labstack/echo"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/api"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/types"
)

func (s *Server) Ping(c echo.Context) error {
	type pingStat struct {
		Version string `json:"version"`
	}
	stats := &pingStat{Version: cfg.ServerVersion}
	return api.OK.SetData(stats).Build(c)
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

func (s *Server) ValidatorStats(c echo.Context) error {
	ctx := context.Background()
	var (
		page, limit int
		err         error
	)
	pagination, page, limit := getPagingOption(c)

	// get validators list from cache
	validators, err := s.getValidators(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}

	// get delegation details
	validator, err := s.kaiClient.Validator(ctx, c.Param("address"))
	if err != nil {
		s.logger.Warn("cannot get validator info from RPC, use cached validator info instead", zap.Error(err))
	}
	// get validator additional info such as commission rate
	for _, val := range validators {
		if strings.ToLower(val.Address.Hex()) == strings.ToLower(c.Param("address")) {
			if validator == nil {
				validator = val
				break
			}
			// update validator VotingPowerPercentage
			validator.VotingPowerPercentage = val.VotingPowerPercentage
			break
		}
	}
	if validator == nil {
		// address in param is not a validator
		return api.Invalid.Build(c)
	}
	var delegators []*types.Delegator
	if pagination.Skip > len(validator.Delegators) {
		delegators = []*types.Delegator(nil)
	} else if pagination.Skip+pagination.Limit > len(validator.Delegators) {
		delegators = validator.Delegators[pagination.Skip:len(validator.Delegators)]
	} else {
		delegators = validator.Delegators[pagination.Skip : pagination.Skip+pagination.Limit]
	}

	total := uint64(len(validator.Delegators))
	validator.Delegators = delegators

	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  validator,
	}).Build(c)
}

func (s *Server) Validators(c echo.Context) error {
	ctx := context.Background()
	validators, err := s.getValidators(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}

	stats, err := s.CalculateValidatorStats(ctx, validators)
	if err != nil {
		return api.Invalid.Build(c)
	}

	var resp struct {
		*types.ValidatorStats
		Validators []*types.Validator `json:"validators"`
	}

	resp.ValidatorStats = stats

	for _, v := range validators {
		if v.Role != 0 {
			resp.Validators = append(resp.Validators, v)
		}
	}
	return api.OK.SetData(resp).Build(c)
}

func (s *Server) GetValidatorsByDelegator(c echo.Context) error {
	ctx := context.Background()
	delAddr := c.Param("address")
	valsList, err := s.kaiClient.GetValidatorsByDelegator(ctx, common.HexToAddress(delAddr))
	if err != nil {
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(valsList).Build(c)
}

func (s *Server) GetCandidatesList(c echo.Context) error {
	ctx := context.Background()
	validators, err := s.getValidators(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}
	var (
		result    []*types.Validator
		valsCount = 0
	)
	for _, val := range validators {
		if val.Role == 0 {
			result = append(result, val)
		} else {
			valsCount++
		}
	}
	return api.OK.SetData(result).Build(c)
}

func (s *Server) GetSlashEvents(c echo.Context) error {
	ctx := context.Background()
	slashEvents, err := s.kaiClient.GetSlashEvents(ctx, common.HexToAddress(c.Param("address")))
	if err != nil {
		s.logger.Warn("Cannot GetSlashEvents", zap.Error(err))
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(slashEvents).Build(c)
}

func (s *Server) GetSlashedTokens(c echo.Context) error {
	ctx := context.Background()
	result, err := s.kaiClient.GetTotalSlashedToken(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(result).Build(c)
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
		smcAddress[v.Address.String()] = &valInfoResponse{
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
		smcAddress[v.Address.String()] = &valInfoResponse{
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
			t.ToName = smcAddress[tx.To].Name
			t.Role = smcAddress[tx.To].Role
			t.IsInValidatorsList = true
		}
		if tx.To == cfg.StakingContractAddr {
			t.ToName = cfg.StakingContractName
		}
		if tx.To == cfg.TreasuryContractAddr {
			t.ToName = cfg.TreasuryContractName
		}
		if tx.To == cfg.KardiaDeployerAddr {
			t.ToName = cfg.KardiaDeployerName
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
		smcAddress[v.Address.String()] = &valInfoResponse{
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
			t.ToName = smcAddress[tx.To].Name
			t.Role = smcAddress[tx.To].Role
			t.IsInValidatorsList = true
		}

		if tx.To == cfg.StakingContractAddr {
			t.ToName = cfg.StakingContractName
		}
		if tx.To == cfg.TreasuryContractAddr {
			t.ToName = cfg.TreasuryContractName
		}
		if tx.To == cfg.KardiaDeployerAddr {
			t.ToName = cfg.KardiaDeployerName
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
			addrInfo.Name = smcAddress[addr.Address].Name
		}
		if addr.Address == cfg.TreasuryContractAddr {
			addrInfo.Name = cfg.TreasuryContractName
		}
		if addr.Address == cfg.StakingContractAddr {
			addrInfo.Name = cfg.StakingContractName
		}
		if addr.Address == cfg.KardiaDeployerAddr {
			addrInfo.Name = cfg.KardiaDeployerName
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
		result := SimpleAddress{
			Address:       addrInfo.Address,
			BalanceString: addrInfo.BalanceString,
			IsContract:    addrInfo.IsContract,
			Name:          addrInfo.Name,
		}
		if smcAddress[result.Address] != nil {
			result.IsInValidatorsList = true
			result.Role = smcAddress[result.Address].Role
			result.Name = smcAddress[result.Address].Name
		}
		if result.Address == cfg.TreasuryContractAddr {
			result.Name = cfg.TreasuryContractName
		}
		if result.Address == cfg.StakingContractAddr {
			result.Name = cfg.StakingContractName
		}
		if result.Address == cfg.KardiaDeployerAddr {
			result.Name = cfg.KardiaDeployerName
		}
		balance, err := s.kaiClient.GetBalance(ctx, address)
		if err != nil {
			return err
		}
		if balance != addrInfo.BalanceString {
			addrInfo.BalanceString = balance
			_ = s.dbClient.UpdateAddresses(ctx, []*types.Address{addrInfo})
		}
		return api.OK.SetData(result).Build(c)
	}
	s.logger.Warn("address not found in db, getting from RPC instead...", zap.Error(err))
	// try to get balance and code at this address to determine whether we should write this address info to database or not
	balance, err := s.kaiClient.GetBalance(ctx, address)
	if err != nil {
		return err
	}
	balanceInBigInt, _ := new(big.Int).SetString(balance, 10)
	balanceFloat, _ := new(big.Float).SetPrec(100).Quo(new(big.Float).SetInt(balanceInBigInt), new(big.Float).SetInt(cfg.Hydro)).Float64() //converting to KAI from HYDRO
	addrInfo = &types.Address{
		Address:       address,
		BalanceFloat:  balanceFloat,
		BalanceString: balance,
		IsContract:    false,
	}
	code, err := s.kaiClient.GetCode(ctx, address)
	if err == nil && len(code) > 0 {
		addrInfo.IsContract = true
	}
	// write this address to db if its balance is larger than 0 or it's a SMC
	if balance != "0" || addrInfo.IsContract {
		_ = s.dbClient.InsertAddress(ctx, addrInfo) // insert this address to database
	}
	result := &SimpleAddress{
		Address:       addrInfo.Address,
		BalanceString: addrInfo.BalanceString,
		IsContract:    addrInfo.IsContract,
	}
	if smcAddress[result.Address] != nil {
		result.IsInValidatorsList = true
		result.Role = smcAddress[result.Address].Role
		result.Name = smcAddress[result.Address].Name
	}
	return api.OK.SetData(result).Build(c)
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
			t.ToName = smcAddress[tx.To].Name
			t.Role = smcAddress[tx.To].Role
			t.IsInValidatorsList = true
		}
		if tx.To == cfg.StakingContractAddr {
			t.ToName = cfg.StakingContractName
		}
		if tx.To == cfg.TreasuryContractAddr {
			t.ToName = cfg.TreasuryContractName
		}
		if tx.To == cfg.KardiaDeployerAddr {
			t.ToName = cfg.KardiaDeployerName
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
	var (
		page, limit int
		err         error
	)
	//blockHash := c.Param("blockHash")
	pageParams := c.QueryParam("page")
	limitParams := c.QueryParam("limit")
	page, err = strconv.Atoi(pageParams)
	if err != nil {
		page = 0
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

func (s *Server) TxByHash(c echo.Context) error {
	ctx := context.Background()
	txHash := c.Param("txHash")
	if txHash == "" {
		return api.Invalid.Build(c)
	}

	var tx *types.Transaction
	tx, err := s.dbClient.TxByHash(ctx, txHash)
	if err != nil {
		// try to get tx by hash through RPC
		s.Logger.Info("cannot get tx by hash from db:", zap.String("txHash", txHash))
		tx, err = s.kaiClient.GetTransaction(ctx, txHash)
		if err != nil {
			s.Logger.Warn("cannot get tx by hash from RPC:", zap.String("txHash", txHash))
			return api.Invalid.Build(c)
		}
		receipt, err := s.kaiClient.GetTransactionReceipt(ctx, txHash)
		if err != nil {
			s.Logger.Warn("cannot get receipt by hash from RPC:", zap.String("txHash", txHash))
		}
		if receipt != nil {
			tx.Logs = receipt.Logs
			tx.Root = receipt.Root
			tx.Status = receipt.Status
			tx.GasUsed = receipt.GasUsed
			tx.ContractAddress = receipt.ContractAddress
		}
	}

	// add name of recipient, if any
	if decoded, err := s.kaiClient.DecodeInputData(tx.To, tx.InputData); err == nil {
		tx.DecodedInputData = decoded
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
		Logs:             tx.Logs,
		TransactionIndex: tx.TransactionIndex,
		LogsBloom:        tx.LogsBloom,
		Root:             tx.Root,
	}
	if result.To == cfg.StakingContractAddr {
		result.ToName = cfg.StakingContractName
		return api.OK.SetData(result).Build(c)
	}
	if result.To == cfg.TreasuryContractAddr {
		result.ToName = cfg.TreasuryContractName
		return api.OK.SetData(result).Build(c)
	}
	if tx.To == cfg.KardiaDeployerAddr {
		result.ToName = cfg.KardiaDeployerName
		return api.OK.SetData(result).Build(c)
	}
	smcAddress := s.getValidatorsAddressAndRole(ctx)
	if smcAddress[result.To] != nil {
		result.ToName = smcAddress[result.To].Name
		result.Role = smcAddress[result.To].Role
		result.IsInValidatorsList = true
		return api.OK.SetData(result).Build(c)
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
			if !delegatorsMap[d.Address.String()] {
				delegatorsMap[d.Address.String()] = true
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
	page, err := strconv.Atoi(pageParams)
	if err != nil {
		page = 0
	}
	limit, err := strconv.Atoi(limitParams)
	if err != nil {
		limit = 10
	}
	pagination := &types.Pagination{
		Skip:  page * limit,
		Limit: limit,
	}
	pagination.Sanitize()
	return pagination, page, limit
}

func (s *Server) getValidatorsAddressAndRole(ctx context.Context) map[string]*valInfoResponse {

	validators, err := s.getValidators(ctx)
	if err != nil {
		return make(map[string]*valInfoResponse)
	}
	//vals, err := s.cacheClient.Validators(ctx)
	//if err != nil {
	//
	//}

	smcAddress := map[string]*valInfoResponse{}
	for _, v := range validators {
		smcAddress[v.SmcAddress.String()] = &valInfoResponse{
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

	return api.OK.Build(c)
}

func (s *Server) ReloadValidators(c echo.Context) error {
	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		return api.Unauthorized.Build(c)
	}

	validators, err := s.kaiClient.Validators(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}

	if err := s.dbClient.UpsertValidators(ctx, validators); err != nil {
		return api.Invalid.Build(c)
	}

	return api.OK.Build(c)
}
