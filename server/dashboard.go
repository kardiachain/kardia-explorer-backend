// Package server
package server

import (
	"context"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/external"
	"github.com/kardiachain/explorer-backend/types"
)

type Dashboard interface {
	Stats(ctx context.Context) ([]*types.TxStats, error)
	TokenHolders(ctx context.Context) (types.HolderStats, error)

	Nodes(ctx context.Context) ([]*types.NodeInfo, error)
	TokenInfo(ctx context.Context) (*types.TokenInfo, error)
	UpdateCirculatingSupply(ctx context.Context, circulatingAmount int64) error
}

func (s *infoServer) Stats(ctx context.Context) ([]*types.TxStats, error) {
	blocks, err := s.dbClient.Blocks(ctx, &types.Pagination{
		Skip:  0,
		Limit: 11,
	})
	if err != nil {
		return nil, err
	}

	var stats []*types.TxStats
	for _, b := range blocks {
		stat := &types.TxStats{
			NumTxs: b.NumTxs,
			Time:   uint64(b.Time.Unix()),
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

func (s *infoServer) TokenHolders(ctx context.Context) (types.HolderStats, error) {
	holders, contracts := s.cacheClient.TotalHolders(ctx)
	holdersStats := types.HolderStats{TotalHolders: holders,
		TotalContracts: contracts}
	return holdersStats, nil
}

func (s *infoServer) Nodes(ctx context.Context) ([]*types.NodeInfo, error) {
	nodes, err := s.cacheClient.NodesInfo(ctx)
	if err == nil && nodes != nil {
		return nodes, nil
	}

	nodes, err = s.kaiClient.NodesInfo(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.cacheClient.UpdateNodesInfo(ctx, nodes); err != nil {
		s.logger.Warn("cannot set nodes info to cache", zap.Error(err))
	}

	return nodes, nil
}

func (s *infoServer) TokenInfo(ctx context.Context) (*types.TokenInfo, error) {
	if !s.cacheClient.IsRequestToCoinMarket(ctx) {
		tokenInfo, err := s.cacheClient.TokenInfo(ctx)
		if err == nil {
			return tokenInfo, nil
		}
	}

	tokenInfo, err := external.TokenInfo(ctx)
	if err != nil {
		return nil, err
	}

	tokenInfo.MarketCap = tokenInfo.Price * float64(tokenInfo.CirculatingSupply)

	if err := s.cacheClient.UpdateTokenInfo(ctx, tokenInfo); err != nil {
		s.logger.Warn("cannot update token info ", zap.Error(err))
	}

	return tokenInfo, nil
}

func (s *infoServer) UpdateCirculatingSupply(ctx context.Context, circulatingAmount int64) error {
	if err := s.cacheClient.UpdateCirculatingSupply(ctx, circulatingAmount); err != nil {
		return err
	}
	return nil
}
