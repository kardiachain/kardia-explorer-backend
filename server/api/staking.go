// Package api
package api

import (
	"context"
	"math/big"
	"sort"

	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

func bindStakingAPIs(gr *echo.Group, srv RestServer) {
	apis := []restDefinition{
		//Validator
		{
			method:      echo.GET,
			path:        "/staking/stats",
			fn:          srv.StakingStats,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/validators/:address",
			fn:          srv.Validator,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/delegators/:address/validators",
			fn:          srv.ValidatorsByDelegator,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/validators/candidates",
			fn:          srv.MobileCandidates,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/validators",
			fn:          srv.MobileValidators,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/staking/candidates",
			fn:          srv.Candidates,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/staking/validators",
			fn:          srv.Validators,
			middlewares: nil,
		},
	}
	for _, api := range apis {
		gr.Add(api.method, api.path, api.fn, api.middlewares...)
	}
}

func (s *Server) StakingStats(c echo.Context) error {
	lgr := s.logger.With(zap.String("method", "StakingStats"))
	ctx := context.Background()
	stats, err := s.cacheClient.StakingStats(ctx)
	if err != nil {
		lgr.Debug("cannot get staking stats from cache", zap.Error(err))
		return Invalid.Build(c)
	}
	return OK.SetData(stats).Build(c)
}

func (s *Server) Validators(c echo.Context) error {
	ctx := context.Background()

	validators, err := s.dbClient.Validators(ctx, db.ValidatorsFilter{})
	if err != nil {
		return Invalid.Build(c)
	}

	var resp []*types.Validator
	sort.Slice(validators, func(i, j int) bool {
		iAmount, _ := new(big.Int).SetString(validators[i].StakedAmount, 10)
		jAmount, _ := new(big.Int).SetString(validators[j].StakedAmount, 10)
		return iAmount.Cmp(jAmount) == 1
	})
	for _, v := range validators {
		if v.Role == 2 || v.Role == 1 {
			resp = append(resp, v)
		}
	}

	return OK.SetData(resp).Build(c)
}

func (s *Server) ValidatorsByDelegator(c echo.Context) error {
	ctx := context.Background()
	delAddr := c.Param("address")
	valsList, err := s.kaiClient.GetValidatorsByDelegator(ctx, common.HexToAddress(delAddr))
	if err != nil {
		return Invalid.Build(c)
	}
	return OK.SetData(valsList).Build(c)
}

func (s *Server) Candidates(c echo.Context) error {
	ctx := context.Background()
	candidates, err := s.dbClient.Validators(ctx, db.ValidatorsFilter{Role: cfg.RoleCandidate})
	if err != nil {
		return Invalid.Build(c)
	}

	return OK.SetData(candidates).Build(c)
}

func SortAscByStakeAmount(address string, delegators []*types.Delegator) []*types.Delegator {
	var delegatorsOwner []*types.Delegator

	for idx, el := range delegators {
		if el.Address == address {
			delegatorsOwner = append(delegatorsOwner, el)
			delegators = RemoveIndexDelegator(delegators, idx)
		}
	}

	sort.Slice(delegatorsOwner, func(i, j int) bool {
		iAmount, _ := new(big.Int).SetString(delegatorsOwner[i].StakedAmount, 10)
		jAmount, _ := new(big.Int).SetString(delegatorsOwner[j].StakedAmount, 10)
		return iAmount.Cmp(jAmount) == 1
	})

	sort.Slice(delegators, func(i, j int) bool {
		iAmount, _ := new(big.Int).SetString(delegators[i].StakedAmount, 10)
		jAmount, _ := new(big.Int).SetString(delegators[j].StakedAmount, 10)
		return iAmount.Cmp(jAmount) == 1
	})

	delegators = append(delegatorsOwner, delegators...)
	return delegators
}

func RemoveIndexDelegator(s []*types.Delegator, index int) []*types.Delegator {
	return append(s[:index], s[index+1:]...)
}

func (s *Server) Validator(c echo.Context) error {
	lgr := s.logger.With(zap.String("method", "Validator"))
	ctx := context.Background()
	var (
		err error
	)

	validatorSMCAddress := c.Param("address")
	pagination, page, limit := getPagingOption(c)

	// get validators list from cache
	validator, err := s.dbClient.Validator(ctx, validatorSMCAddress)
	if err != nil {
		lgr.Error("cannot load validator from db", zap.Error(err))
		return Invalid.Build(c)
	}

	// get delegation details
	filter := db.DelegatorFilter{
		ValidatorSMCAddress: validator.SmcAddress,
		Skip:                int64(pagination.Skip),
		Limit:               int64(pagination.Limit),
	}
	delegators, err := s.dbClient.Delegators(ctx, filter)
	if err != nil {
		lgr.Error("cannot load validator from db", zap.Error(err))
		return Invalid.Build(c)
	}

	validator.Delegators = SortAscByStakeAmount(validatorSMCAddress, delegators)
	total, err := s.dbClient.CountDelegators(ctx, filter)
	if err != nil {
		lgr.Error("cannot count delegator", zap.Error(err))
		return Invalid.Build(c)
	}

	return OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: uint64(total),
		Data:  validator,
	}).Build(c)
}

func (s *Server) MobileValidators(c echo.Context) error {
	ctx := context.Background()

	validators, err := s.dbClient.Validators(ctx, db.ValidatorsFilter{})
	if err != nil {
		return Invalid.Build(c)
	}

	var resp []*types.Validator
	sort.Slice(validators, func(i, j int) bool {
		iAmount, _ := new(big.Int).SetString(validators[i].StakedAmount, 10)
		jAmount, _ := new(big.Int).SetString(validators[j].StakedAmount, 10)
		return iAmount.Cmp(jAmount) == 1
	})
	for _, v := range validators {
		if v.Role == 2 || v.Role == 1 {
			resp = append(resp, v)
		}
	}
	stats, err := s.cacheClient.StakingStats(ctx)
	if err != nil {
		stats = &types.StakingStats{}
	}
	type mobileResponse struct {
		*types.StakingStats
		Validators []*types.Validator `json:"validators"`
	}

	mobileResp := mobileResponse{stats, resp}
	return OK.SetData(mobileResp).Build(c)
}

func (s *Server) MobileCandidates(c echo.Context) error {
	ctx := context.Background()
	candidates, err := s.dbClient.Validators(ctx, db.ValidatorsFilter{Role: cfg.RoleCandidate})
	if err != nil {
		return Invalid.Build(c)
	}

	stats, err := s.cacheClient.StakingStats(ctx)
	if err != nil {
		stats = &types.StakingStats{}
	}
	type mobileResponse struct {
		*types.StakingStats
		Validators []*types.Validator `json:"validators"`
	}

	mobileResp := mobileResponse{stats, candidates}
	return OK.SetData(mobileResp).Build(c)
}
