// Package server
package server

import (
	"context"
	"errors"
	"math/big"
	"strconv"

	"github.com/labstack/echo"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/types"
)

func (s *Server) newAddressInfo(ctx context.Context, address string) (*types.Address, error) {
	balance, err := s.kaiClient.GetBalance(ctx, address)
	if err != nil {
		return nil, err
	}
	balanceInBigInt, _ := new(big.Int).SetString(balance, 10)
	balanceFloat, _ := new(big.Float).SetPrec(100).Quo(new(big.Float).SetInt(balanceInBigInt), new(big.Float).SetInt(cfg.Hydro)).Float64() //converting to KAI from HYDRO
	addrInfo := &types.Address{
		Address:       address,
		BalanceFloat:  balanceFloat,
		BalanceString: balance,
		IsContract:    false,
	}
	code, err := s.kaiClient.GetCode(ctx, address)
	if err == nil && len(code) > 0 {
		addrInfo.IsContract = true
	}
	// write this address to db if its balance is larger than 0 or it's a SMC or it holds KRC token
	tokens, _, _ := s.dbClient.KRC20Holders(ctx, &types.KRC20HolderFilter{
		HolderAddress: address,
	})
	if balance != "0" || addrInfo.IsContract || len(tokens) > 0 {
		_ = s.dbClient.InsertAddress(ctx, addrInfo) // insert this address to database
	}
	return &types.Address{
		Address:       addrInfo.Address,
		BalanceString: addrInfo.BalanceString,
		IsContract:    addrInfo.IsContract,
	}, nil
}

//getValidators
func (s *Server) getValidators(ctx context.Context) ([]*types.Validator, error) {
	//validators, err := s.cacheClient.Validators(ctx)
	//if err == nil && len(validators.Validators) != 0 {
	//	return validators, nil
	//}
	// Try from db
	dbValidators, err := s.dbClient.Validators(ctx, db.ValidatorsFilter{})
	if err == nil {
		//s.logger.Debug("get validators from storage", zap.Any("Validators", dbValidators))
		stats, err := s.CalculateValidatorStats(ctx, dbValidators)
		if err == nil && len(dbValidators) != 0 {
			s.logger.Debug("stats ", zap.Any("stats", stats))
		}
		return dbValidators, nil
		//return dbValidators, nil
	}

	s.logger.Debug("Load validator from network")
	validators, err := s.kaiClient.Validators(ctx)
	if err != nil {
		s.logger.Warn("cannot get validators list from RPC", zap.Error(err))
		return nil, err
	}
	//err = s.cacheClient.UpdateValidators(ctx, vasList)
	//if err != nil {
	//	s.logger.Warn("cannot store validators list to cache", zap.Error(err))
	//}
	return validators, nil
}

func (s *Server) CalculateValidatorStats(ctx context.Context, validators []*types.Validator) (*types.ValidatorStats, error) {
	var stats types.ValidatorStats
	var (
		ErrParsingBigIntFromString = errors.New("cannot parse big.Int from string")
		proposersStakedAmount      = big.NewInt(0)
		delegatorsMap              = make(map[string]bool)
		totalProposers             = 0
		totalValidators            = 0
		totalCandidates            = 0
		totalStakedAmount          = big.NewInt(0)
		totalDelegatorStakedAmount = big.NewInt(0)
		totalDelegators            = 0

		valStakedAmount *big.Int
		delStakedAmount *big.Int
		ok              bool
	)
	for _, val := range validators {
		// Calculate total staked amount
		valStakedAmount, ok = new(big.Int).SetString(val.StakedAmount, 10)
		if !ok {
			return nil, ErrParsingBigIntFromString
		}
		totalStakedAmount = new(big.Int).Add(totalStakedAmount, valStakedAmount)

		for _, d := range val.Delegators {
			if !delegatorsMap[d.Address] {
				delegatorsMap[d.Address] = true
				totalDelegators++
			}
			delStakedAmount, ok = new(big.Int).SetString(d.StakedAmount, 10)
			if !ok {
				return nil, ErrParsingBigIntFromString
			}
			if d.Address == val.Address {
				proposersStakedAmount = new(big.Int).Add(proposersStakedAmount, delStakedAmount)
			} else {

				totalDelegatorStakedAmount = new(big.Int).Add(totalDelegatorStakedAmount, delStakedAmount)
			}
		}
		//val.Role = ec.getValidatorRole(valsSet, val.Address, val.Status)
		// validator who started a node and not in validators set is a normal validator
		if val.Role == 2 {
			totalProposers++
			totalValidators++
		} else if val.Role == 1 {
			totalValidators++
		} else if val.Role == 0 {
			totalCandidates++
		}
	}
	stats.TotalStakedAmount = totalStakedAmount.String()
	stats.TotalDelegatorStakedAmount = totalDelegatorStakedAmount.String()
	stats.TotalValidatorStakedAmount = proposersStakedAmount.String()
	stats.TotalDelegators = totalDelegators
	stats.TotalCandidates = totalCandidates
	stats.TotalValidators = totalValidators
	stats.TotalProposers = totalProposers
	return &stats, nil
}

func getPagingOption(c echo.Context) (*types.Pagination, int, int) {
	pageParams := c.QueryParam("page")
	limitParams := c.QueryParam("limit")
	if pageParams == "" && limitParams == "" {
		return nil, 0, 0
	}
	page, err := strconv.Atoi(pageParams)
	if err != nil {
		page = 1
	}
	page = page - 1
	limit, err := strconv.Atoi(limitParams)
	if err != nil {
		limit = 25
	}
	pagination := &types.Pagination{
		Skip:  page * limit,
		Limit: limit,
	}
	pagination.Sanitize()
	return pagination, page + 1, limit
}

func (s *Server) getValidatorsAddressAndRole(ctx context.Context) map[string]*valInfoResponse {
	validators, err := s.getValidators(ctx)
	if err != nil {
		return make(map[string]*valInfoResponse)
	}

	smcAddress := map[string]*valInfoResponse{}
	for _, v := range validators {
		smcAddress[v.SmcAddress] = &valInfoResponse{
			Name: v.Name,
			Role: v.Role,
		}
	}
	return smcAddress
}

func (s *Server) getAddressInfo(ctx context.Context, address string) (*types.Address, error) {
	addrInfo, err := s.cacheClient.AddressInfo(ctx, address)
	if err == nil {
		return addrInfo, nil
	}
	s.logger.Info("Cannot get address info in cache, getting from db instead", zap.String("address", address), zap.Error(err))
	addrInfo, err = s.dbClient.AddressByHash(ctx, address)
	if err != nil {
		s.logger.Warn("Cannot get address info from db", zap.String("address", address), zap.Error(err))
		if err != nil {
			// insert new address to db
			newAddr, err := s.newAddressInfo(ctx, address)
			if err != nil {
				s.logger.Warn("Cannot store address info to db", zap.Any("address", newAddr), zap.Error(err))
			}
		}
		return nil, err
	}
	err = s.cacheClient.UpdateAddressInfo(ctx, addrInfo)
	if err != nil {
		s.logger.Warn("Cannot store address info to cache", zap.String("address", address), zap.Error(err))
	}
	return addrInfo, nil
}
