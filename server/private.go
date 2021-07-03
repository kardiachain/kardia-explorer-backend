// Package server
package server

import (
	"context"
	"strings"

	kClient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/kardia-explorer-backend/api"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

func (s *Server) RefreshContractsInfo(c echo.Context) error {
	lgr := s.Logger
	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		return api.Unauthorized.Build(c)
	}

	contracts, err := s.dbClient.AllContracts(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}

	for _, c := range contracts {

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

	return api.OK.Build(c)
}

func (s *Server) RefreshKRC721Info(c echo.Context) error {
	lgr := s.Logger
	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		return api.Unauthorized.Build(c)
	}

	krc721Tokens, err := s.dbClient.ContractByType(ctx, cfg.SMCTypeKRC721)
	if err != nil {
		return api.Invalid.Build(c)
	}

	for _, krc721 := range krc721Tokens {
		krc721.Status = types.ContractStatusUnverified
		if krc721.IsVerified {
			krc721.Status = types.ContractStatusVerified
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

	return api.OK.Build(c)
}

func (s *Server) RefreshKRC20Info(c echo.Context) error {
	lgr := s.Logger
	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		return api.Unauthorized.Build(c)
	}

	krc20Tokens, err := s.dbClient.ContractByType(ctx, cfg.SMCTypeKRC20)
	if err != nil {
		return api.Invalid.Build(c)
	}

	for _, krc20 := range krc20Tokens {
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

	return api.OK.Build(c)
}

func (s *Server) SyncContractInfo(c echo.Context) error {
	lgr := s.Logger
	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		return api.Unauthorized.Build(c)
	}

	//  Select all txs which contractAddress != ''
	contractCreationTxs, err := s.dbClient.FindContractCreationTxs(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}

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
					contract.Name = addressInfo.TokenName
					contract.Symbol = addressInfo.TokenSymbol
					contract.Decimals = uint8(addressInfo.Decimals)
					contract.TotalSupply = addressInfo.TotalSupply
				}

			}
			if err := s.dbClient.UpdateContract(ctx, contract, nil); err != nil {
				lgr.Error("cannot update contract", zap.Error(err))
			}
		}
		if tx.Status == types.TransactionStatusFailed {
			if err := s.dbClient.RemoveContract(ctx, tx.ContractAddress); err != nil {
				lgr.Error("cannot delete contract")
			}
		}
	}

	return api.OK.Build(c)
}

func (s *Server) RemoveNilContracts(c echo.Context) error {
	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.infoServer.HttpRequestSecret {
		return api.Unauthorized.Build(c)
	}
	if err := s.dbClient.RemoveContracts(ctx); err != nil {
		return api.Invalid.Build(c)
	}

	return api.OK.Build(c)
}
