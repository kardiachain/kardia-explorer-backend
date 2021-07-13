// Package audit
package audit

import (
	"context"

	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/kardia"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"go.uber.org/zap"
)

type Server struct {
	dbClient    db.Client
	cacheClient cache.Client
	kaiClient   kardia.ClientInterface

	logger *zap.Logger
}

func (s *Server) doAuditBlocks(ctx context.Context, start, end uint64) {
	for i := start; i <= end; i++ {

		nBlock, err := s.kaiClient.BlockByHeight(ctx, i)
		if err != nil {
			continue
		}

		dbBlock, err := s.dbClient.BlockByHeight(ctx, i)
		if err != nil {
			continue
		}
		if err := s.UpsertBlock(ctx, block); err != nil {
			lgr.Error("failed to upsert block", zap.Error(err))
			_ = srv.InsertErrorBlocks(ctx, blockHeight-1, blockHeight+1)
			continue
		}

		if nBlock.NumTxs > 0 {
			// Process txs and receipts
			s.ProcessTxs(ctx, nBlock)

		}

	}

}
func (s *Server) ImportBlock(ctx context.Context, block *types.Block, writeToCache bool) error {
	lgr := s.logger.With(zap.String("method", "ImportBlock"))
	lgr.Info("Importing block:", zap.Uint64("Height", block.Height),
		zap.Int("Txs length", len(block.Txs)), zap.Int("Receipts length", len(block.Receipts)))
	if isExist, err := s.dbClient.IsBlockExist(ctx, block.Height); err != nil || isExist {
		return types.ErrRecordExist
	}

	if writeToCache {
		if err := s.cacheClient.InsertBlock(ctx, block); err != nil {
			s.logger.Debug("cannot import block to cache", zap.Error(err))
		}
	}

	// update number of block proposed by this proposer
	numOfBlocks, err := s.cacheClient.CountBlocksOfProposer(ctx, block.ProposerAddress)
	if err != nil || numOfBlocks == 0 {
		numOfBlocks, err = s.dbClient.CountBlocksOfProposer(ctx, block.ProposerAddress)
		if err != nil {
			s.logger.Error("cannot get number of blocks by proposer from db", zap.Error(err), zap.String("proposer", block.ProposerAddress))
		}
	}
	if numOfBlocks > 0 {
		if err = s.cacheClient.UpdateNumOfBlocksByProposer(ctx, block.ProposerAddress, numOfBlocks+1); err != nil {
			s.logger.Warn("cannot set number of blocks by proposer to cache", zap.Error(err), zap.Any("block", block))
		}
	}

	if err := s.dbClient.InsertBlock(ctx, block); err != nil {
		return err
	}
	return nil
}

func (s *Server) ProcessTxs(ctx context.Context, block *types.Block) error {
	lgr := s.logger
	// merge receipts into corresponding transactions
	// because getBlockByHash/Height API returns 2 array contains txs and receipts separately
	block.Txs = s.mergeAdditionalInfoToTxs(ctx, block.Txs, block.Receipts)
	if err := s.dbClient.InsertTxs(ctx, block.Txs); err != nil {
		return err
	}
	if _, err := s.cacheClient.UpdateTotalTxs(ctx, block.NumTxs); err != nil {
		return err
	}

	var receiptHashes []string
	for _, r := range block.Receipts {
		receiptHashes = append(receiptHashes, r.TransactionHash)
	}
	if len(receiptHashes) > 0 {
		lgr.Info("Push receipts", zap.Any("Receipts", receiptHashes))
		if err := s.cacheClient.PushReceipts(ctx, receiptHashes); err != nil {
			lgr.Error("cannot push receipts", zap.Error(err))
			return err
		}
	}

	return nil
}

func (s *Server) ProcessActiveAddress(ctx context.Context, txs []*types.Transaction) error {
	lgr := s.logger.With(zap.String("method", "ProcessActiveAddress"))
	// update active addresses
	s.filterContracts(ctx, txs)

	addrsMap := filterAddrSet(txs)
	addrsList := s.getAddressBalances(ctx, addrsMap)

	if err := s.dbClient.UpdateAddresses(ctx, addrsList); err != nil {
		return err
	}
	totalAddresses, err := s.dbClient.CountAddresses(ctx)
	if err == nil {
		if err := s.cacheClient.UpdateTotalAddresses(ctx, totalAddresses); err != nil {
			lgr.Error("cannot update total addresses", zap.Error(err))
		}
	}
	totalContracts, err := s.dbClient.CountContracts(ctx)
	if err == nil {
		if err := s.cacheClient.UpdateTotalContracts(ctx, totalContracts); err != nil {
			lgr.Error("cannot update total contracts", zap.Error(err))
		}
	}

	return nil
}
