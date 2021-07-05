// Package api
package api

import (
	"context"
	"strconv"

	kClient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

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
		return Invalid.Build(c)
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
	return OK.SetData(PagingResponse{
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
			return Invalid.Build(c)
		}
		code, err := s.kaiClient.GetCode(ctx, address)
		if err != nil {
			s.logger.Warn("Cannot get address code from RPC", zap.String("address", address), zap.Error(err))
			return Invalid.Build(c)
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
		return OK.SetData(result).Build(c)
	}
	s.logger.Warn("address not found in db, getting from RPC instead...", zap.Error(err))
	// try to get balance and code at this address to determine whether we should write this address info to database or not
	newAddr, err := s.newAddressInfo(ctx, address)
	if err != nil {
		return Invalid.Build(c)
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
	return OK.SetData(result).Build(c)
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

	return OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  result,
	}).Build(c)
}

func (s *Server) AddressHolders(c echo.Context) error {
	lgr := s.logger
	ctx := context.Background()
	var (
		page, limit int
		err         error
	)
	address := c.Param("address")
	pagination, page, limit := getPagingOption(c)
	filterCrit := &types.HolderFilter{
		Pagination:    pagination,
		HolderAddress: address,
	}
	holders, total, err := s.dbClient.GetListHolders(ctx, filterCrit)
	if err != nil {
		lgr.Error("cannot get holders from db", zap.Error(err))
	}
	krc20ABI, err := kClient.KRC20ABI()
	if err != nil {
		return Invalid.Build(c)
	}
	for i := range holders {
		tokenInfo, _ := s.getKRCTokenInfo(ctx, holders[i].ContractAddress)
		if tokenInfo != nil {
			holders[i].Logo = tokenInfo.Logo
			balance, err := s.kaiClient.GetKRC20BalanceByAddress(ctx, krc20ABI, common.HexToAddress(holders[i].ContractAddress), common.HexToAddress(holders[i].HolderAddress))
			if err != nil {
				s.logger.Error("cannot holder balance of token", zap.Error(err), zap.String("holderAddress", holders[i].HolderAddress),
					zap.String("tokenAddress", holders[i].ContractAddress))
				continue
			}
			holders[i].BalanceString = balance.String()
			holders[i].BalanceFloat = s.calculateKRC20BalanceFloat(balance, tokenInfo.Decimals)
		}
	}
	return OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  holders,
	}).Build(c)
}

func (s *Server) SearchAddressByName(c echo.Context) error {
	var (
		ctx     = context.Background()
		name    = c.QueryParam("name")
		addrMap = make(map[string]*SimpleKRCTokenInfo)
	)
	if name == "" {
		return Invalid.Build(c)
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
		return Invalid.Build(c)
	}
	result := make([]*SimpleKRCTokenInfo, 0, len(addrMap))
	for _, addr := range addrMap {
		result = append(result, addr)
	}
	return OK.SetData(result).Build(c)
}
