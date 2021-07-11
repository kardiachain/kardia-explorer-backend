// Package api
package api

import (
	"context"

	kClient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/go-kardia/lib/common"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

type IContract interface {
	Contracts(c echo.Context) error
	Contract(c echo.Context) error
	UpdateContract(c echo.Context) error
	UpdateSMCABIByType(c echo.Context) error
	ContractEvents(c echo.Context) error
}

func bindContractAPIs(gr *echo.Group, srv RestServer) {
	apis := []restDefinition{

		{
			method: echo.GET,
			// Query params
			// [?status=(Verified, Unverified)]
			path:        "/contracts",
			fn:          srv.Contracts,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/contracts/:contractAddress",
			fn:          srv.Contract,
			middlewares: nil,
		},
		{
			method: echo.GET,
			// Query params: ?page=0&limit=10&contractAddress=0x&methodName=0x&txHash=0x
			path:        "/contracts/events",
			fn:          srv.ContractEvents,
			middlewares: nil,
		},
	}
	for _, api := range apis {
		gr.Add(api.method, api.path, api.fn, api.middlewares...)
	}
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
			return Invalid.Build(c)
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
	return OK.SetData(PagingResponse{
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

	results, total, err := s.dbClient.Contracts(ctx, filterCrit)
	if err != nil {
		return Invalid.Build(c)
	}

	finalResult := make([]*SimpleKRCTokenInfo, len(results))
	for i := range results {
		if results[i].Type == cfg.SMCTypeKRC20 {
			abi, err := kClient.KRC20ABI()
			if err == nil {
				krcInfo, err := s.kaiClient.GetKRC20TokenInfo(ctx, abi, common.HexToAddress(results[i].Address))
				if err == nil && krcInfo != nil {
					results[i].TotalSupply = krcInfo.TotalSupply
				}
			}
		}
		finalResult[i] = &SimpleKRCTokenInfo{
			Name:        results[i].Name,
			Address:     results[i].Address,
			Info:        results[i].Info,
			Type:        results[i].Type,
			Logo:        results[i].Logo,
			IsVerified:  results[i].IsVerified,
			Status:      int64(results[i].Status),
			TokenSymbol: results[i].Symbol,
			Decimal:     int64(results[i].Decimals),
		}
	}
	return OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  finalResult,
	}).Build(c)
}

func (s *Server) Contract(c echo.Context) error {
	ctx := context.Background()
	contractAddress := c.Param("contractAddress")

	smc, addrInfo, err := s.dbClient.Contract(ctx, contractAddress)
	if err != nil {
		return Invalid.Build(c)
	}

	result := &KRCTokenInfo{
		Name:          smc.Name,
		Address:       smc.Address,
		OwnerAddress:  smc.OwnerAddress,
		TxHash:        smc.TxHash,
		Type:          smc.Type,
		BalanceString: addrInfo.BalanceString,
		Info:          smc.Info,
		Logo:          smc.Logo,
		IsContract:    addrInfo.IsContract,
		TokenName:     smc.Name,
		TokenSymbol:   smc.Symbol,
		Decimals:      int64(smc.Decimals),
		TotalSupply:   smc.TotalSupply,
		Status:        int64(smc.Status),
		CreatedAt:     smc.CreatedAt,
	}

	if smc.Type == cfg.SMCTypeKRC20 {
		// Get totalSupply from network
		abi, err := kClient.KRC20ABI()
		if err == nil {
			krcInfo, err := s.kaiClient.GetKRC20TokenInfo(ctx, abi, common.HexToAddress(smc.Address))
			if err != nil {
				return Invalid.Build(c)
			}
			result.TotalSupply = krcInfo.TotalSupply
		}
	}
	return OK.SetData(result).Build(c)
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
