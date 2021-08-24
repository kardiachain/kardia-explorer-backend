// Package balance
package balance

import (
	"context"
	"errors"
	"math/big"
	"time"

	kClient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/kardiachain/kardia-explorer-backend/utils"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

var (
	ZERO_BI = new(big.Int).SetInt64(0)
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

//var ErrNotFoundReceipt = errors.New("not found")

func (s *Server) HandleTokenBalance(ctx context.Context, interval time.Duration) {
	//	badReceipts := make(map[string]int)
	// Read receipt from cache and start processing flow
	lgr := s.logger.With(zap.String("task", "handle_token_balance"))
	lgr.Info("Run process token flow...")
	//poolSize := 8
	//p, err := ants.NewPoolWithFunc(poolSize, func(i interface{}) {
	//	r := i.(*kClient.Receipt)
	//	if err := s.processBalance(ctx, r); err != nil {
	//		lgr.Error("cannot handle pair event", zap.Error(err))
	//	}
	//}, ants.WithPreAlloc(true))
	//if err != nil {
	//	return
	//}
	//defer p.Release()

	for {
		select {
		case <-time.After(1000 * time.Millisecond):
			tokenHash, err := s.cache.PopToken(ctx)
			if err != nil {
				if errors.Is(err, ErrRedisNil) {
					lgr.Info("No token left! Sleep ")
					time.Sleep(10 * time.Second)
					continue
				}
				continue
			}

			if tokenHash == "" {
				continue
			}

			// Get holders from db
			lgr.Info("Processing token", zap.String("TokenHash", tokenHash))
			holders, totalHolders, err := s.db.KRC20Holders(ctx, &types.KRC20HolderFilter{
				ContractAddress: tokenHash,
			})
			token, err := kClient.NewToken(s.node, tokenHash)
			if err != nil {
				lgr.Error("cannot create token instance", zap.Error(err))
				continue
			}

			lgr.Info("total holder", zap.Uint64("total holder", totalHolders))
			tokenInfo, err := token.KRC20Info(ctx)
			if err != nil {
				lgr.Error("cannot get token info", zap.Error(err))
				return
			}

			for i := uint64(0); i < totalHolders; i++ {
				holder := holders[i]
				balance, err := token.HolderBalance(ctx, holder.HolderAddress)
				if err != nil {
					lgr.Error("cannot balance of address", zap.String("address", holder.HolderAddress))
					continue
				}

				lgr.Info("Balance from network", zap.String("Balance", balance.String()))
				holder.BalanceString = balance.String()
				holder.BalanceFloat = utils.BalanceToFloatWithDecimals(balance, int64(tokenInfo.Decimals))
				holder.UpdatedAt = time.Now().Unix()

				if balance.Cmp(ZERO_BI) == 0 {
					lgr.Info("Remove holder since from balance = 0")
					if err := s.db.RemoveKRC20Holder(ctx, holder); err != nil {
						lgr.Error("cannot remove holder", zap.Error(err))
						continue
					}
				} else {
					lgr.Info("Update holder since from balance != 0")
					if err := s.db.UpsertKRC20Holders(ctx, []*types.KRC20Holder{holder}); err != nil {
						lgr.Error("cannot update holder", zap.Error(err))
						continue
					}
				}
			}

		}

	}
}
