// Package receipts
package receipts

import (
	"context"
	"errors"
	"time"

	kClient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"go.uber.org/zap"
)

type Server struct {
	db db.Client

	node   kClient.Node
	cache  cache.Client
	logger *zap.Logger
}

var ErrRedisNil = errors.New("redis: nil")

func (s *Server) ProcessReceipts(ctx context.Context, interval time.Duration) {
	// Read receipt from cache and start processing flow
	lgr := s.logger.With(zap.String("Task", "ProcessReceipts"))
	lgr.Info("Run process receipts flow...")
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			receiptHash, err := s.cache.PopReceipt(ctx)
			if err != nil {
				if errors.Is(err, ErrRedisNil) {
					lgr.Info("No receipt left!")
					continue
				}
			}
			lgr.Info("Processing", zap.String("ReceiptHash", receiptHash))
			// Get receipt from network
			r, err := s.node.GetTransactionReceipt(ctx, receiptHash)
			if err != nil {
				// Push back receipt hash into list
				lgr.Error("cannot get receipt from network. Push back for retry later", zap.Error(err))
				if err := s.cache.PushReceipts(ctx, []string{receiptHash}); err != nil {
					lgr.Error("cannot push back receipt hash into list", zap.Error(err))
					// todo: Implement notify
					continue
				}
			}

			// If failed
			if r.Status == 0 {
				continue
			}

			// Start processing
			if err := s.processReceipt(ctx, r); err != nil {
				lgr.Error("cannot process receipt", zap.Error(err))
				if err := s.cache.PushReceipts(ctx, []string{receiptHash}); err != nil {
					lgr.Error("cannot push back receipt into list", zap.Error(err))
					// todo: Implement notify
					continue
				}
			}
		}
	}
}

func (s *Server) processReceipt(ctx context.Context, r *kClient.Receipt) error {
	for _, l := range r.Logs {
		// Process if transfer event
		if l.Topics[0] == cfg.KRCTransferTopic {
			s.processTransferLog(ctx, l)
		}

		// Process if mint/burn event

	}

	return nil
}
