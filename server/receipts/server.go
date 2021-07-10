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
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

type Server struct {
	db db.Client

	node   kClient.Node
	cache  cache.Client
	logger *zap.Logger
	p      ants.PoolWithFunc
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
	badReceipts := make(map[string]int)
	// Read receipt from cache and start processing flow
	lgr := s.logger.With(zap.String("task", "handle_receipts"))
	lgr.Info("Run process receipts flow...")
	poolSize := 64
	p, err := ants.NewPoolWithFunc(poolSize, func(i interface{}) {
		r := i.(*kClient.Receipt)
		if err := s.processReceipt(ctx, r); err != nil {
			lgr.Error("cannot handle pair event", zap.Error(err))
		}
	}, ants.WithPreAlloc(true))
	if err != nil {
		return
	}
	defer p.Release()

	for {
		receiptHash, err := s.cache.PopReceipt(ctx)
		if err != nil {
			if errors.Is(err, ErrRedisNil) {
				lgr.Info("No receipt left! Sleep ")
				time.Sleep(interval)
				continue
			}
		}

		if receiptHash == "" {
			continue
		}

		// Get receipt from network
		lgr.Info("Processing receipt", zap.String("ReceiptHash", receiptHash))
		r, err := s.node.GetTransactionReceipt(ctx, receiptHash)
		if err != nil {
			lgr.Error("cannot get receipt from network", zap.Error(err))
			// Push back receipt hash into list

			if badReceipts[receiptHash] < 3 {
				badReceipts[receiptHash]++
				lgr.Info("Push back to process list")
				if err := s.cache.PushReceipts(ctx, []string{receiptHash}); err != nil {
					lgr.Error("cannot push back receipt hash into list", zap.Error(err))
					// todo: Implement notify
					continue
				}
			}

			if badReceipts[receiptHash] >= 3 {
				lgr.Info("Skip, insert into bad list")
				if err := s.cache.PushBadReceipts(ctx, []string{receiptHash}); err != nil {
					lgr.Error("cannot push back receipt hash into list", zap.Error(err))
					// todo: Implement notify
					continue
				}
				delete(badReceipts, receiptHash)
			}
			continue
		}

		// If failed
		if r.Status == 0 {
			continue
		}

		if err := p.Invoke(r); err != nil {
			lgr.Error("invoke process error", zap.Error(err))
			if badReceipts[receiptHash] < 3 {
				badReceipts[receiptHash]++
				lgr.Info("Push back to process list")
				if err := s.cache.PushReceipts(ctx, []string{receiptHash}); err != nil {
					lgr.Error("cannot push back receipt hash into list", zap.Error(err))
					// todo: Implement notify
					continue
				}
			}

			if badReceipts[receiptHash] >= 3 {
				lgr.Info("Skip, insert into bad list")
				if err := s.cache.PushBadReceipts(ctx, []string{receiptHash}); err != nil {
					lgr.Error("cannot push back receipt hash into list", zap.Error(err))
					// todo: Implement notify
					continue
				}
				delete(badReceipts, receiptHash)
			}
		}

		//// Start processing
		//if err := s.processReceipt(ctx, r); err != nil {
		//	lgr.Error("cannot process receipt", zap.Error(err))
		//	if err := s.cache.PushReceipts(ctx, []string{receiptHash}); err != nil {
		//		lgr.Error("cannot push back receipt into list", zap.Error(err))
		//		// todo: Implement notify
		//		continue
		//	}
		//}
	}
}

func (s *Server) processReceipt(ctx context.Context, r *kClient.Receipt) error {
	//lgr := s.logger
	for _, l := range r.Logs {
		// Process if transfer event
		if l.Topics[0] == cfg.KRCTransferTopic {
			s.processTransferLog(ctx, l)
		}

		// Process if mint/burn event

	}

	return nil
}
