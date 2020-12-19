// Package server
package server

import (
	"context"

	"github.com/kardiachain/go-kardia/lib/common"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

type Validator interface {
	ValidatorStats(ctx context.Context, address string, pagination *types.Pagination) (*types.Validators, error)
	Validators(ctx context.Context) (*types.Validators, error)
	ValidatorsOfDelegator(ctx context.Context, address string) ([]*types.ValidatorsByDelegator, error)
	CandidatesList(ctx context.Context) (*types.Validators, error)
	SlashEvents(ctx context.Context, address string) ([]*types.SlashEvents, error)
	BlocksByProposer(ctx context.Context, address string, pagination *types.Pagination) ([]*types.Block, uint64, error)
}

func (s *infoServer) Validators(ctx context.Context) (*types.Validators, error) {
	valsList, err := s.getValidatorsList(ctx)
	if err != nil {
		return nil, err
	}
	var (
		result []*types.Validator
	)
	for _, val := range valsList.Validators {
		if val.Status > 0 {
			result = append(result, val)
		}
	}
	valsList.Validators = result
	return valsList, nil
}

func (s *infoServer) ValidatorsOfDelegator(ctx context.Context, address string) ([]*types.ValidatorsByDelegator, error) {
	return s.kaiClient.GetValidatorsByDelegator(ctx, common.HexToAddress(address))
}

func (s *infoServer) CandidatesList(ctx context.Context) (*types.Validators, error) {
	valsList, err := s.getValidatorsList(ctx)
	if err != nil {
		return nil, err
	}
	var (
		result []*types.Validator
	)
	for _, val := range valsList.Validators {
		if val.Status == 0 {
			result = append(result, val)
		}
	}
	valsList.Validators = result
	return valsList, nil
}

func (s *infoServer) SlashEvents(ctx context.Context, address string) ([]*types.SlashEvents, error) {
	return s.kaiClient.GetSlashEvents(ctx, common.HexToAddress(address))
}

func (s *infoServer) BlocksByProposer(ctx context.Context, address string, pagination *types.Pagination) ([]*types.Block, uint64, error) {
	return s.dbClient.BlocksByProposer(ctx, address, pagination)
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

func (s *infoServer) getValidatorsList(ctx context.Context) (*types.Validators, error) {
	valsList, err := s.kaiClient.Validators(ctx)
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
