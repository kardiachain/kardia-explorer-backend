// Package server
package server

import (
	"context"

	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/labstack/echo"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/api"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/types"
)

func (s *Server) StakingStats(c echo.Context) error {
	lgr := s.logger.With(zap.String("method", "StakingStats"))
	ctx := context.Background()
	stats, err := s.cacheClient.StakingStats(ctx)
	if err != nil {
		lgr.Debug("cannot get staking stats from cache", zap.Error(err))
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(stats).Build(c)
}

func (s *Server) Validators(c echo.Context) error {
	ctx := context.Background()

	validators, err := s.dbClient.Validators(ctx, db.ValidatorsFilter{})
	if err != nil {
		return api.Invalid.Build(c)
	}

	var resp []*types.Validator

	for _, v := range validators {
		if v.Role == 2 {
			resp = append(resp, v)
		}
	}
	return api.OK.SetData(resp).Build(c)
}

func (s *Server) ValidatorsByDelegator(c echo.Context) error {
	ctx := context.Background()
	delAddr := c.Param("address")
	valsList, err := s.kaiClient.GetValidatorsByDelegator(ctx, common.HexToAddress(delAddr))
	if err != nil {
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(valsList).Build(c)
}

func (s *Server) Candidates(c echo.Context) error {
	ctx := context.Background()
	candidates, err := s.dbClient.Validators(ctx, db.ValidatorsFilter{Role: cfg.RoleCandidate})
	if err != nil {
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(candidates).Build(c)
}

func (s *Server) Validator(c echo.Context) error {
	lgr := s.logger.With(zap.String("method", "Validator"))
	ctx := context.Background()
	var (
		err error
	)

	validatorSMCAddress := c.Param("address")
	_, page, limit := getPagingOption(c)

	// get validators list from cache
	validator, err := s.dbClient.Validator(ctx, validatorSMCAddress)
	if err != nil {
		lgr.Error("cannot load validator from db", zap.Error(err))
		return api.Invalid.Build(c)
	}

	// get delegation details
	filter := db.DelegatorFilter{
		ValidatorSMCAddress: validatorSMCAddress,
		Skip:                int64(page),
		Limit:               int64(limit),
	}
	delegators, err := s.dbClient.Delegators(ctx, filter)
	if err != nil {
		lgr.Error("cannot load validator from db", zap.Error(err))
		return api.Invalid.Build(c)
	}
	validator.Delegators = delegators
	total, err := s.dbClient.CountDelegators(ctx, filter)
	if err != nil {
		lgr.Error("cannot count delegator", zap.Error(err))
		return api.Invalid.Build(c)
	}

	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: uint64(total),
		Data:  validator,
	}).Build(c)
}
