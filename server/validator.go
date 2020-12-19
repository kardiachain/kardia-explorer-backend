// Package server
package server

import (
	"context"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

type Validator interface {
	ValidatorStats(ctx context.Context, address string, pagination *types.Pagination) (*types.Validators, error)
	Validators(ctx context.Context) (*types.Validators, error)
}

func (s *infoServer) Validators(ctx context.Context) (*types.Validators, error) {
	return s.getValidatorsListFromCache(ctx)
}

func (s *infoServer) getValidatorsListFromCache(ctx context.Context) (*types.Validators, error) {
	valsList, err := s.cacheClient.Validators(ctx)
	if err == nil {
		s.logger.Debug("got validators list from cache", zap.Error(err))
		return valsList, nil
	}
	s.logger.Warn("cannot get validators list from cache", zap.Error(err))
	valsList, err = s.kaiClient.Validators(ctx)
	if err != nil {
		s.logger.Warn("cannot get validators list from RPC", zap.Error(err))
		return nil, err
	}
	s.logger.Debug("Got validators list from RPC")
	err = s.cacheClient.UpdateValidators(ctx, valsList)
	if err != nil {
		s.logger.Warn("cannot store validators list to cache", zap.Error(err))
	}
	return valsList, nil
}
