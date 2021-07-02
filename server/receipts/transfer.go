// Package receipts
package receipts

import (
	"context"

	kClient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/go-kardia/lib/abi"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"go.uber.org/zap"
)

func (s *Server) processTransferLog(ctx context.Context, l *kClient.Log) error {
	lgr := s.logger
	contract, _, err := s.db.Contract(ctx, l.Address)
	if err != nil {
		lgr.Error("cannot get contract from db", zap.Error(err))
		lgr.Info("Try get token info from network")
		// todo: Maybe insert new contract here since it may create in SMC
		// Try to get basic information about its
		token, err := kClient.NewToken(s.node, l.Address)
		if err != nil {
			lgr.Error("cannot create token object", zap.Error(err))
			return nil
		}

		newKRC20Info, err := token.KRC20Info(ctx)
		if err != nil {
			lgr.Error("cannot get KRC20 info", zap.Error(err))
			return nil
		}
		// Fetch tx info for fill into contract
		tx, err := s.node.GetTransaction(ctx, l.TxHash)
		if err != nil {
			lgr.Error("cannot get tx", zap.Error(err))
			return nil
		}
		newKRC20 := &types.Contract{
			Name:         newKRC20Info.Name,
			Address:      l.Address,
			OwnerAddress: tx.From,
			TxHash:       tx.Hash,
			Type:         cfg.SMCTypeKRC20,
			Symbol:       newKRC20Info.Symbol,
			TotalSupply:  newKRC20Info.TotalSupply.String(),
			Decimals:     newKRC20Info.Decimals,
			Logo:         cfg.DefaultKRCTokenLogo,
			IsVerified:   false,
			CreatedAt:    tx.Time.Unix(),
			UpdatedAt:    tx.Time.Unix(),
		}
		if err := s.db.InsertContract(ctx, newKRC20, nil); err != nil {
			lgr.Error("cannot insert contract", zap.Error(err))
			return nil
		}
		contract = newKRC20
	}
	lgr.Info("process transfer logs")
	switch contract.Type {
	case cfg.SMCTypeKRC20:
		return s.onKRC20Transfer(ctx, contract, l)
	case cfg.SMCTypeKRC721:
		return s.onKRC721Transfer(ctx, contract, l)
	default:
		return s.onUndetectedContractTransfer(ctx, contract, l)
	}
}

func (s *Server) onUndetectedContractTransfer(ctx context.Context, c *types.Contract, l *kClient.Log) error {
	lgr := s.logger
	lgr.Debug("handle undetected token transfer")
	// Try with KRC721 first
	var (
		unpackedLog *kClient.Log
		err         error
	)
	unpackedLog, err = tryKRC721(l)
	if err == nil {
		// Try get basic information about this token and update into storage
		token, err := kClient.NewToken(s.node, c.Address)
		if err != nil {
			lgr.Error("cannot create token instance", zap.Error(err))
		}

		krc721Info, err := token.KRC721Info(ctx)
		if err == nil {
			lgr.Error("cannot get KRC721 info", zap.Error(err))
			if krc721Info.Name != "" {
				c.Name = krc721Info.Name
			}

			if krc721Info.Symbol != "" {
				c.Symbol = krc721Info.Symbol
			}

			if krc721Info.TotalSupply != nil {
				c.TotalSupply = krc721Info.TotalSupply.String()
			}
		}
		// Take default logo
		c.Logo = cfg.DefaultKRCTokenLogo
		c.Type = cfg.SMCTypeKRC721
		// Update into db
		if err := s.db.UpdateContract(ctx, c, nil); err != nil {
			lgr.Error("cannot update contract", zap.Error(err))
			return err
		}
		// Insert new transfer and holder
		if err := s.insertKRC721Transfer(ctx, unpackedLog); err != nil {
			return err
		}

		if err := s.upsertKRC721Holder(ctx, unpackedLog); err != nil {
			return err
		}

		lgr.Info("Contract is KRC721", zap.String("Address", c.Address))
		return nil
	}

	unpackedLog, err = tryKRC20(l)
	if err == nil {
		// Try get basic information about this token and update into storage
		token, err := kClient.NewToken(s.node, c.Address)
		if err != nil {
			lgr.Error("cannot create token instance", zap.Error(err))
		}

		krc20Info, err := token.KRC20Info(ctx)
		if err == nil {
			if krc20Info.Name != "" {
				c.Name = krc20Info.Name
			}

			if krc20Info.Symbol != "" {
				c.Symbol = krc20Info.Symbol
			}

			if krc20Info.Decimals != 0 {
				c.Decimals = krc20Info.Decimals
			}

			if krc20Info.TotalSupply != nil {
				c.TotalSupply = krc20Info.TotalSupply.String()
			}
		}
		c.Logo = cfg.DefaultKRCTokenLogo
		c.Type = cfg.SMCTypeKRC20
		// Update into db
		if err := s.db.UpdateContract(ctx, c, nil); err != nil {
			lgr.Error("cannot update contract", zap.Error(err))
			return err
		}

		// Insert new transfer and holder
		if err := s.insertKRC20Transfer(ctx, unpackedLog); err != nil {
			return err
		}

		if err := s.upsertKRC20Holder(ctx, unpackedLog); err != nil {
			return err
		}

		lgr.Info("Contract is KRC20", zap.String("Address", c.Address))
		return nil
	}
	return nil
}

func (s *Server) onKRC20Transfer(ctx context.Context, c *types.Contract, l *kClient.Log) error {
	lgr := s.logger
	var krcABI *abi.ABI
	krcABI, err := kClient.KRC20ABI()
	if err != nil {
		return err
	}
	if c.ABI != "" {
		// Decode and use contract ABI instead

	}

	unpackedLog, err := kClient.UnpackLog(l, krcABI)
	if err != nil {
		return err
	}
	lgr.Info("UnpackLog", zap.Any("UnpackedLog", unpackedLog))

	// Insert new transfer and holder
	if err := s.insertKRC20Transfer(ctx, unpackedLog); err != nil {
		lgr.Error("cannot insert token transfer", zap.Error(err))
		return err
	}

	if err := s.upsertKRC20Holder(ctx, unpackedLog); err != nil {
		lgr.Error("cannot upsert token holder", zap.Error(err))
		return err
	}

	return nil
}

func (s *Server) onKRC721Transfer(ctx context.Context, c *types.Contract, l *kClient.Log) error {
	lgr := s.logger
	var krcABI *abi.ABI
	krcABI, err := kClient.KRC721ABI()
	if err != nil {
		return err
	}
	if c.ABI != "" {
		// Decode and use contract ABI instead
	}

	unpackedLog, err := kClient.UnpackLog(l, krcABI)
	if err != nil {
		return err
	}
	lgr.Info("UnpackLog", zap.Any("UnpackedLog", unpackedLog))
	// Insert new transfer and holder
	if err := s.insertKRC721Transfer(ctx, unpackedLog); err != nil {
		lgr.Error("cannot insert token transfer", zap.Error(err))
		return err
	}

	if err := s.upsertKRC721Holder(ctx, unpackedLog); err != nil {
		lgr.Error("cannot insert token holder", zap.Error(err))
		return err
	}

	return nil
}
