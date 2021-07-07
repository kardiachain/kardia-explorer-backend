// Package api
package api

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	kClient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/kardiachain/kardia-explorer-backend/utils"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

type IPrivate interface {
	ReloadAddressesBalance(c echo.Context) error
	UpdateAddressName(c echo.Context) error
	UpsertNetworkNodes(c echo.Context) error
	RemoveNetworkNodes(c echo.Context) error
	UpdateSupplyAmounts(c echo.Context) error
	RemoveDuplicateEvents(c echo.Context) error

	//todo: Rework or remove
	ReloadValidators(c echo.Context) error

	//
	RemoveNilContracts(c echo.Context) error
	SyncContractInfo(c echo.Context) error
	RefreshKRC20Info(c echo.Context) error
	RefreshKRC721Info(c echo.Context) error
	RefreshContractsInfo(c echo.Context) error
	RefreshHolders(c echo.Context) error
}

func bindPrivateAPIs(gr *echo.Group, srv RestServer) {
	apis := []restDefinition{
		{
			method:      echo.PUT,
			path:        "/contracts/sync",
			fn:          srv.SyncContractInfo,
			middlewares: nil,
		},
		{
			method:      echo.PUT,
			path:        "/contracts/kcr20/refresh",
			fn:          srv.RefreshKRC20Info,
			middlewares: nil,
		},

		{
			method:      echo.PUT,
			path:        "/contracts/kcr721/refresh",
			fn:          srv.RefreshKRC721Info,
			middlewares: nil,
		},
		{
			method:      echo.PUT,
			path:        "/contracts/refresh",
			fn:          srv.RefreshContractsInfo,
			middlewares: nil,
		},
		{
			method:      echo.DELETE,
			path:        "/contracts/nil",
			fn:          srv.RemoveNilContracts,
			middlewares: nil,
		},
		{
			method:      echo.DELETE,
			path:        "/holders/refresh",
			fn:          srv.RefreshHolders,
			middlewares: nil,
		},
		{
			method:      echo.PUT,
			path:        "/contracts",
			fn:          srv.UpdateContract,
			middlewares: nil,
		},
		{
			method:      echo.PUT,
			path:        "/contracts/abi",
			fn:          srv.UpdateSMCABIByType,
			middlewares: nil,
		},
	}
	for _, api := range apis {
		gr.Add(api.method, api.path, api.fn, api.middlewares...)
	}
}

func (s *Server) RefreshContractsInfo(c echo.Context) error {
	lgr := s.logger
	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.authorizationSecret {
		return Unauthorized.Build(c)
	}

	contracts, err := s.dbClient.AllContracts(ctx)
	if err != nil {
		return Invalid.Build(c)
	}

	for _, c := range contracts {
		if c.Type == "" {
			c.Type = cfg.SMCTypeNormal
		}
		c.Status = types.ContractStatusUnverified
		if c.IsVerified {
			c.Status = types.ContractStatusVerified
		}

		if c.Type == cfg.SMCTypeValidator ||
			c.Type == cfg.SMCTypeStaking ||
			c.Type == cfg.SMCTypeParams ||
			c.Type == cfg.SMCTypeTreasury {
			// Set verified for special contracts
			c.Status = types.ContractStatusVerified
		}
		if err := s.dbClient.UpdateContract(ctx, c, nil); err != nil {
			lgr.Error("cannot update contract", zap.Error(err))
			continue
		}
	}

	return OK.SetData(nil).Build(c)
}

func (s *Server) RefreshKRC721Info(c echo.Context) error {

	lgr := s.logger
	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.authorizationSecret {
		return Unauthorized.Build(c)
	}

	krc721Tokens, err := s.dbClient.ContractByType(ctx, cfg.SMCTypeKRC721)
	if err != nil {
		return Invalid.Build(c)
	}

	for _, krc721 := range krc721Tokens {
		krc721.Status = types.ContractStatusUnverified
		if krc721.IsVerified {
			krc721.Status = types.ContractStatusVerified
		}
		fmt.Println("Address", krc721.Address)
		token, err := kClient.NewToken(s.node, krc721.Address)
		if err != nil {
			lgr.Error("cannot create token object", zap.Error(err))
			continue
		}
		krc721Info, err := token.KRC721Info(ctx)
		if err != nil {
			lgr.Error("cannot get KRC721 info of token", zap.Error(err))
			continue
		}
		if krc721Info.Name != "" {
			krc721.Name = krc721Info.Name
		}

		if krc721Info.Symbol != "" {
			krc721.Symbol = krc721Info.Symbol
		}

		if krc721Info.TotalSupply != nil {
			krc721.TotalSupply = krc721Info.TotalSupply.String()
		}

		// Change base64 image to default token
		if strings.HasPrefix(krc721.Logo, "data:image") {
			krc721.Logo = cfg.DefaultKRCTokenLogo
		}

		if err := s.dbClient.UpdateContract(ctx, krc721, nil); err != nil {
			lgr.Error("cannot update contract", zap.Error(err))
			continue
		}
	}

	return OK.SetData(nil).Build(c)
}

func (s *Server) RefreshKRC20Info(c echo.Context) error {

	lgr := s.logger
	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.authorizationSecret {
		return Unauthorized.Build(c)
	}

	krc20Tokens, err := s.dbClient.ContractByType(ctx, cfg.SMCTypeKRC20)
	if err != nil {
		return Invalid.Build(c)
	}

	for _, krc20 := range krc20Tokens {
		lgr.Info("Process KRC20", zap.String("Address", krc20.Address))
		krc20.Status = types.ContractStatusUnverified
		if krc20.IsVerified {
			krc20.Status = types.ContractStatusVerified
		}

		token, err := kClient.NewToken(s.node, krc20.Address)
		if err != nil {
			lgr.Error("cannot create token object", zap.Error(err))
			continue
		}
		krc20Info, err := token.KRC20Info(ctx)
		if err != nil {
			lgr.Error("cannot get KRC20 info of token", zap.Error(err))
			continue
		}
		if krc20Info.Name != "" {
			krc20.Name = krc20Info.Name
		}

		krc20.Symbol = krc20Info.Symbol
		krc20.Decimals = krc20Info.Decimals

		if krc20Info.TotalSupply != nil {
			krc20.TotalSupply = krc20Info.TotalSupply.String()
		}
		// Change base64 image to default token
		if strings.HasPrefix(krc20.Logo, "data:image") {
			krc20.Logo = cfg.DefaultKRCTokenLogo
		}

		if err := s.dbClient.UpdateContract(ctx, krc20, nil); err != nil {
			lgr.Error("cannot update contract", zap.Error(err))
			continue
		}
	}

	return OK.SetData(nil).Build(c)
}

func (s *Server) SyncContractInfo(c echo.Context) error {

	lgr := s.logger
	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.authorizationSecret {
		return Unauthorized.Build(c)
	}

	//  Select all txs which contractAddress != ''
	contractCreationTxs, err := s.dbClient.FindContractCreationTxs(ctx)
	if err != nil {
		return Invalid.Build(c)
	}

	lgr.Info("Total contract to sync", zap.Int("Size", len(contractCreationTxs)))

	// Find contract info in `Address` collection and upsert with addition information into `Contracts` collection
	for _, tx := range contractCreationTxs {
		if tx.Status == types.TransactionStatusSuccess {
			contract := &types.Contract{
				Address:      tx.ContractAddress,
				OwnerAddress: tx.From,
				TxHash:       tx.Hash,
				Type:         cfg.SMCTypeNormal,
				CreatedAt:    tx.Time.Unix(),
				UpdatedAt:    tx.Time.Unix(),
			}

			addressInfo, err := s.dbClient.AddressByHash(ctx, tx.ContractAddress)
			if err == nil {
				contract.Name = addressInfo.Name
				if addressInfo.KrcTypes != "" {
					contract.Type = addressInfo.KrcTypes
				}
				contract.Info = addressInfo.Info

				if contract.Type == cfg.SMCTypeKRC20 { // Sync KRC20 information
					fmt.Println("KRC20 Contract", contract.Address)
					contract.Name = addressInfo.TokenName
					contract.Symbol = addressInfo.TokenSymbol
					contract.Decimals = uint8(addressInfo.Decimals)
					contract.Logo = addressInfo.Logo
					contract.TotalSupply = addressInfo.TotalSupply
				}

			}
			if err := s.dbClient.UpdateContract(ctx, contract, nil); err != nil {
				lgr.Error("cannot update contract", zap.Error(err))
			}
		}
		if tx.Status == types.TransactionStatusFailed {
			if err := s.dbClient.RemoveContract(ctx, tx.ContractAddress); err != nil {
				lgr.Error("cannot delete contract", zap.Error(err))
			}
		}
	}

	return OK.SetData(nil).Build(c)
}

func (s *Server) RemoveNilContracts(c echo.Context) error {

	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.authorizationSecret {
		return Unauthorized.Build(c)
	}
	if err := s.dbClient.RemoveContracts(ctx); err != nil {
		return Invalid.Build(c)
	}

	return OK.SetData(nil).Build(c)
}

func (s *Server) UpsertNetworkNodes(c echo.Context) error {
	//ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.authorizationSecret {
		return Unauthorized.Build(c)
	}
	var nodeInfo *types.NodeInfo
	if err := c.Bind(&nodeInfo); err != nil {
		return Invalid.Build(c)
	}
	if nodeInfo.ID == "" || nodeInfo.Moniker == "" {
		return Invalid.Build(c)
	}
	ctx := context.Background()
	if err := s.dbClient.UpsertNode(ctx, nodeInfo); err != nil {
		return InternalServer.Build(c)
	}

	return OK.Build(c)
}

func (s *Server) RemoveNetworkNodes(c echo.Context) error {
	//ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.authorizationSecret {
		return Unauthorized.Build(c)
	}
	nodesID := c.Param("nodeID")
	if nodesID == "" {
		return Invalid.Build(c)
	}

	ctx := context.Background()
	if err := s.dbClient.RemoveNode(ctx, nodesID); err != nil {
		return InternalServer.Build(c)
	}

	return OK.Build(c)
}

func (s *Server) ReloadAddressesBalance(c echo.Context) error {
	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.authorizationSecret {
		return Unauthorized.Build(c)
	}

	addresses, err := s.dbClient.Addresses(ctx)
	if err != nil {
		return Invalid.Build(c)
	}

	for id, a := range addresses {
		balance, err := s.kaiClient.GetBalance(ctx, a.Address)
		if err != nil {
			continue
		}
		addresses[id].BalanceString = balance
	}

	if err := s.dbClient.UpdateAddresses(ctx, addresses); err != nil {
		return Invalid.Build(c)
	}

	return OK.Build(c)
}

func (s *Server) UpdateAddressName(c echo.Context) error {
	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.authorizationSecret {
		return Unauthorized.Build(c)
	}
	var addressName types.UpdateAddress
	if err := c.Bind(&addressName); err != nil {
		fmt.Println("cannot bind ", err)
		return Invalid.Build(c)
	}
	addressInfo, err := s.dbClient.AddressByHash(ctx, addressName.Address)
	if err != nil {
		return Invalid.Build(c)
	}

	addressInfo.Name = addressName.Name

	if err := s.dbClient.UpdateAddresses(ctx, []*types.Address{addressInfo}); err != nil {
		fmt.Println("cannot update ", err)
		return Invalid.Build(c)
	}
	_ = s.cacheClient.UpdateAddressInfo(ctx, addressInfo)
	return OK.Build(c)
}

func (s *Server) ReloadValidators(c echo.Context) error {
	if c.Request().Header.Get("Authorization") != s.authorizationSecret {
		return Unauthorized.Build(c)
	}

	//todo longnd: rework reload validator API
	//validators, err := s.kaiClient.Validators(ctx)
	//if err != nil {
	//	return Invalid.Build(c)
	//}
	//
	//if err := s.dbClient.UpsertValidators(ctx, validators); err != nil {
	//	return Invalid.Build(c)
	//}

	return OK.Build(c)
}

func (s *Server) RemoveDuplicateEvents(c echo.Context) error {
	ctx := context.Background()
	data, err := s.dbClient.RemoveDuplicateEvents(ctx)
	if err != nil {
		return InternalServer.Build(c)
	}
	return OK.SetData(data).Build(c)
}

func (s *Server) RefreshHolders(c echo.Context) error {
	ctx := context.Background()
	if err := s.dbClient.RemoveKRC20Holders(ctx); err != nil {
		return Invalid.Build(c)
	}

	return OK.SetData(nil).Build(c)
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
	contract.Status = types.ContractStatusVerified
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
