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

func (s *Server) SetLogger(logger *zap.Logger) *Server {
	s.logger = logger
	return s
}

func (s *Server) SetStorage(db db.Client) *Server {
	s.db = db
	return s
}

func (s *Server) SetCache(cache cache.Client) *Server {
	s.cache = cache
	return s
}

func (s *Server) SetNode(node kClient.Node) *Server {
	s.node = node
	return s
}

var ErrRedisNil = errors.New("redis: nil")

func (s *Server) HandleReceipts(ctx context.Context, interval time.Duration) {
	// Read receipt from cache and start processing flow
	lgr := s.logger.With(zap.String("task", "handle_receipts"))
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

			if receiptHash == "" {
				continue
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
				continue
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
	lgr := s.logger
	for _, l := range r.Logs {
		// Process if transfer event
		lgr.Debug("process log", zap.Any("log", l))
		if l.Topics[0] == cfg.KRCTransferTopic {
			s.processTransferLog(ctx, l)
		}

		// Process if mint/burn event

	}

	return nil
}
