// Package api
package api

import (
	"bytes"
	"context"
	"io/ioutil"
	"strings"

	kClient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/go-kardia/lib/common"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/kardiachain/kardia-explorer-backend/utils"
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
			method:      echo.PUT,
			path:        "/contracts",
			fn:          srv.UpdateContract,
			middlewares: nil,
		},
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
			method:      echo.PUT,
			path:        "/contracts/abi",
			fn:          srv.UpdateSMCABIByType,
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

	result, total, err := s.dbClient.Contracts(ctx, filterCrit)
	if err != nil {
		return Invalid.Build(c)
	}

	finalResult := make([]*SimpleKRCTokenInfo, len(result))
	for i := range result {
		finalResult[i] = &SimpleKRCTokenInfo{
			Name:        result[i].Name,
			Address:     result[i].Address,
			Info:        result[i].Info,
			Type:        result[i].Type,
			Logo:        result[i].Logo,
			IsVerified:  result[i].IsVerified,
			Status:      int64(result[i].Status),
			TotalSupply: result[i].TotalSupply,
			TokenSymbol: result[i].Symbol,
			Decimal:     int64(result[i].Decimals),
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
		Info:          addrInfo.Info,
		Logo:          addrInfo.Logo,
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

func (s *Server) UpdateContract(c echo.Context) error {
	lgr := s.logger.With(zap.String("method", "UpdateContract"))
	if c.Request().Header.Get("Authorization") != s.authorizationSecret {
		return Unauthorized.Build(c)
	}

	var (
		contract     types.Contract
		addrInfo     types.Address
		bodyBytes, _ = ioutil.ReadAll(c.Request().Body)
	)
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	if err := c.Bind(&contract); err != nil {
		lgr.Error("cannot bind contract data", zap.Error(err))
		return Invalid.Build(c)
	}
	contract.IsVerified = true
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	if err := c.Bind(&addrInfo); err != nil {
		lgr.Error("cannot bind address data", zap.Error(err))
		return Invalid.Build(c)
	}
	ctx := context.Background()
	krcTokenInfoFromRPC, err := s.getKRCTokenInfoFromRPC(ctx, addrInfo.Address, addrInfo.KrcTypes)
	if err != nil && strings.HasPrefix(addrInfo.KrcTypes, "KRC") {
		s.logger.Warn("Updating contract is not KRC type", zap.Any("smcInfo", addrInfo), zap.Error(err))
		return Invalid.Build(c)
	}
	if krcTokenInfoFromRPC != nil {
		// cache new token info
		krcTokenInfoFromRPC.Logo = addrInfo.Logo

		if (strings.Contains(addrInfo.Logo, "https") ||
			strings.Contains(addrInfo.Logo, "http")) &&
			strings.Contains(addrInfo.Logo, "png") &&
			!strings.HasPrefix(addrInfo.Logo, s.ConfigUploader.PathAvatar) {
			addrInfo.Logo = utils.ConvertUrlPngToBase64(addrInfo.Logo)
		}

		if utils.CheckBase64Logo(addrInfo.Logo) {
			addressHash := contract.Address
			if strings.HasPrefix(addressHash, "0x") {
				addressHash = string(addressHash[2:])
			}
			fileName, err := s.fileStorage.UploadLogo(addrInfo.Logo, addressHash, s.ConfigUploader)
			if err != nil {
				lgr.Error("cannot upload image", zap.Error(err))
			} else {
				addrInfo.Logo = fileName
				contract.Logo = fileName
			}
		}

		_ = s.cacheClient.UpdateKRCTokenInfo(ctx, krcTokenInfoFromRPC)
		_ = s.cacheClient.UpdateSMCAbi(ctx, contract.Address, contract.ABI)

		addrInfo.TokenName = krcTokenInfoFromRPC.TokenName
		addrInfo.TokenSymbol = krcTokenInfoFromRPC.TokenSymbol
		addrInfo.TotalSupply = krcTokenInfoFromRPC.TotalSupply
		addrInfo.Decimals = krcTokenInfoFromRPC.Decimals
	}
	if err := s.dbClient.UpdateContract(ctx, &contract, &addrInfo); err != nil {
		lgr.Error("cannot bind insert", zap.Error(err))
		return InternalServer.Build(c)
	}

	return OK.SetData(addrInfo).Build(c)
}

func (s *Server) UpdateSMCABIByType(c echo.Context) error {
	if c.Request().Header.Get("Authorization") != s.authorizationSecret {
		return Unauthorized.Build(c)
	}
	ctx := context.Background()
	var smcABI *types.ContractABI
	if err := c.Bind(&smcABI); err != nil {
		return Invalid.Build(c)
	}
	err := s.dbClient.UpsertSMCABIByType(ctx, smcABI.Type, smcABI.ABI)
	if err != nil {
		return Invalid.Build(c)
	}
	return OK.Build(c)
}
