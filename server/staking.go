// Package server
package server

import (
	"context"
	"strings"

	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/labstack/echo"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/api"
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

func (s *Server) GetValidatorsByDelegator(c echo.Context) error {
	ctx := context.Background()
	delAddr := c.Param("address")
	valsList, err := s.kaiClient.GetValidatorsByDelegator(ctx, common.HexToAddress(delAddr))
	if err != nil {
		return api.Invalid.Build(c)
	}
	return api.OK.SetData(valsList).Build(c)
}

func (s *Server) GetCandidatesList(c echo.Context) error {
	ctx := context.Background()
	validators, err := s.getValidators(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}
	var (
		result    []*types.Validator
		valsCount = 0
	)
	for _, val := range validators {
		if val.Role == 0 {
			result = append(result, val)
		} else {
			valsCount++
		}
	}
	return api.OK.SetData(result).Build(c)
}

func (s *Server) Validator(c echo.Context) error {
	ctx := context.Background()
	var (
		page, limit int
		err         error
	)
	pagination, page, limit := getPagingOption(c)

	// get validators list from cache
	validators, err := s.getValidators(ctx)
	if err != nil {
		return api.Invalid.Build(c)
	}

	// get delegation details
	validator, err := s.kaiClient.Validator(ctx, c.Param("address"))
	if err != nil {
		s.logger.Warn("cannot get validator info from RPC, use cached validator info instead", zap.Error(err))
	}
	// get validator additional info such as commission rate
	for _, val := range validators {
		if strings.ToLower(val.Address) == strings.ToLower(c.Param("address")) {
			if validator == nil {
				validator = val
				break
			}
			// update validator VotingPowerPercentage
			validator.VotingPowerPercentage = val.VotingPowerPercentage
			break
		}
	}
	if validator == nil {
		// address in param is not a validator
		return api.Invalid.Build(c)
	}
	var delegators []*types.Delegator
	if pagination.Skip > len(validator.Delegators) {
		delegators = []*types.Delegator(nil)
	} else if pagination.Skip+pagination.Limit > len(validator.Delegators) {
		delegators = validator.Delegators[pagination.Skip:len(validator.Delegators)]
	} else {
		delegators = validator.Delegators[pagination.Skip : pagination.Skip+pagination.Limit]
	}

	total := uint64(len(validator.Delegators))
	validator.Delegators = delegators

	return api.OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  validator,
	}).Build(c)
}
