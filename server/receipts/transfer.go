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
	contract, _, err := s.db.Contract(ctx, l.Address)
	if err != nil {
		return nil
	}

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
			return err
		}

		krc20Info, err := token.KRC20Info(ctx)
		if err != nil {
			return err
		}
		if krc20Info.Name != "" {
			c.Name = krc20Info.Name
		}

		if krc20Info.Symbol != "" {
			c.Name = krc20Info.Name
		}

		if krc20Info.Name != "" {
			c.Name = krc20Info.Name
		}

		if krc20Info.Name != "" {
			c.Name = krc20Info.Name
		}

		c.Type = cfg.SMCTypeKRC721
		// Update into db
		if err := s.db.UpdateContract(ctx, c, nil); err != nil {
			return err
		}
		// Insert new transfer and holder
		if err := s.insertTokenTransfer(ctx, unpackedLog); err != nil {
			return err
		}

		if err := s.upsertTokenHolder(ctx, unpackedLog); err != nil {
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
			return err
		}

		krc721Info, err := token.KRC721Info(ctx)
		if err != nil {
			return err
		}
		c.Name = krc721Info.Name
		c.Type = cfg.SMCTypeKRC721
		// Update into db
		if err := s.db.UpdateContract(ctx, c, nil); err != nil {
			return err
		}

		// Insert new transfer and holder
		if err := s.insertTokenTransfer(ctx, unpackedLog); err != nil {
			return err
		}

		if err := s.upsertTokenHolder(ctx, unpackedLog); err != nil {
			return err
		}

		lgr.Info("Contract is KRC20", zap.String("Address", c.Address))
		return nil
	}

	return nil
}

func (s *Server) onKRC20Transfer(ctx context.Context, c *types.Contract, l *kClient.Log) error {
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
	s.logger.Info("UnpackLog", zap.Any("UnpackedLog", unpackedLog))

	// Insert new transfer and holder
	if err := s.insertTokenTransfer(ctx, unpackedLog); err != nil {
		return err
	}

	if err := s.upsertTokenHolder(ctx, unpackedLog); err != nil {
		return err
	}

	return nil
}

func (s *Server) onKRC721Transfer(ctx context.Context, c *types.Contract, l *kClient.Log) error {
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
	s.logger.Info("UnpackLog", zap.Any("UnpackedLog", unpackedLog))
	// Insert new transfer and holder
	if err := s.insertTokenTransfer(ctx, unpackedLog); err != nil {
		return err
	}

	if err := s.upsertTokenHolder(ctx, unpackedLog); err != nil {
		return err
	}

	return nil
}
