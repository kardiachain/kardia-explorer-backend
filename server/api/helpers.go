// Package api
package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kardiachain/go-kardia/lib/abi"
	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

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

func (s *Server) getSMCAbi(ctx context.Context, log *types.Log) (*abi.ABI, error) {
	smcABIStr, err := s.cacheClient.SMCAbi(ctx, log.Address)
	if err != nil {
		smc, _, err := s.dbClient.Contract(ctx, log.Address)
		if err != nil {
			s.logger.Debug("Cannot get smc info from db", zap.Error(err), zap.String("smcAddr", log.Address))
			return nil, err
		}
		if smc.Type != "" {
			err = s.cacheClient.UpdateSMCAbi(ctx, log.Address, cfg.SMCTypePrefix+smc.Type)
			if err != nil {
				s.logger.Warn("Cannot store smc abi to cache", zap.Error(err))
				return nil, err
			}
			smcABIStr, err = s.cacheClient.SMCAbi(ctx, cfg.SMCTypePrefix+smc.Type)
			if err != nil {
				// query then reinsert abi of this SMC type to cache
				smcABIBase64, err := s.dbClient.SMCABIByType(ctx, smc.Type)
				if err != nil {
					s.logger.Warn("Cannot get smc abi by type from DB", zap.Error(err))
					return nil, err
				}
				err = s.cacheClient.UpdateSMCAbi(ctx, cfg.SMCTypePrefix+smc.Type, smcABIBase64)
				if err != nil {
					s.logger.Warn("Cannot store smc abi by type to cache", zap.Error(err))
					return nil, err
				}
				smcABIStr, err = s.cacheClient.SMCAbi(ctx, cfg.SMCTypePrefix+smc.Type)
				if err != nil {
					s.logger.Warn("Cannot get smc abi from cache", zap.Error(err))
					return nil, err
				}
			}
		} else if smc.ABI != "" {
			smcABIStr = smc.ABI
		}
	}
	return s.decodeSMCABIFromBase64(ctx, smcABIStr, log.Address)
}

func (s *Server) decodeSMCABIFromBase64(ctx context.Context, abiStr, smcAddr string) (*abi.ABI, error) {
	abiData, err := base64.StdEncoding.DecodeString(abiStr)
	if err != nil {
		s.logger.Warn("Cannot decode smc abi", zap.Error(err))
		return nil, err
	}
	jsonABI, err := abi.JSON(bytes.NewReader(abiData))
	if err != nil {
		s.logger.Warn("Cannot convert decoded smc abi to JSON abi", zap.Error(err))
		return nil, err
	}
	// store this abi to cache
	err = s.cacheClient.UpdateSMCAbi(ctx, smcAddr, abiStr)
	if err != nil {
		s.logger.Warn("Cannot store smc abi to cache", zap.Error(err))
		return nil, err
	}
	return &jsonABI, nil
}

func (s *Server) getTokenInfo(ctx context.Context, address string) (*types.KRCTokenInfo, error) {
	contractInfo, _, err := s.dbClient.Contract(ctx, address)
	if err != nil {
		return nil, err
	}
	result := &types.KRCTokenInfo{
		Address:     contractInfo.Address,
		TokenName:   contractInfo.Name,
		TokenType:   contractInfo.Type,
		TokenSymbol: contractInfo.Symbol,
		TotalSupply: contractInfo.TotalSupply,
		Decimals:    int64(contractInfo.Decimals),
		Logo:        contractInfo.Logo,
	}
	return result, nil
}

func (s *Server) getKRCTokenInfo(ctx context.Context, krcTokenAddr string) (*types.KRCTokenInfo, error) {
	krcTokenInfo, err := s.cacheClient.KRCTokenInfo(ctx, krcTokenAddr)
	if err == nil {
		return krcTokenInfo, nil
	}
	s.logger.Warn("Cannot get KRC token info from cache, getting from database instead")
	addrInfo, err := s.dbClient.AddressByHash(ctx, krcTokenAddr)
	if err != nil {
		s.logger.Warn("Cannot get KRC token info from db", zap.Error(err))
		return nil, err
	}
	result := &types.KRCTokenInfo{
		Address:     addrInfo.Address,
		TokenName:   addrInfo.TokenName,
		TokenType:   addrInfo.KrcTypes,
		TokenSymbol: addrInfo.TokenSymbol,
		TotalSupply: addrInfo.TotalSupply,
		Decimals:    addrInfo.Decimals,
		Logo:        addrInfo.Logo,
	}
	err = s.cacheClient.UpdateKRCTokenInfo(ctx, result)
	if err != nil {
		s.logger.Warn("Cannot store KRC token info to cache", zap.Error(err))
		return nil, err
	}
	return result, nil
}

func (s *Server) getKRCTokenInfoFromRPC(ctx context.Context, krcTokenAddress, krcType string) (*types.KRCTokenInfo, error) {
	var tokenInfo *types.KRCTokenInfo
	if strings.EqualFold(krcType, cfg.SMCTypeKRC20) {
		// get KRC20 token info from RPC
		smcABIStr, err := s.dbClient.SMCABIByType(ctx, krcType)
		if err != nil {
			s.logger.Warn("Cannot get smc abi from db", zap.Error(err))
			return nil, err
		}
		abiData, err := base64.StdEncoding.DecodeString(smcABIStr)
		if err != nil {
			s.logger.Warn("Cannot decode smc abi", zap.Error(err))
			return nil, err
		}
		jsonABI, err := abi.JSON(bytes.NewReader(abiData))
		if err != nil {
			s.logger.Warn("Cannot convert decoded smc abi to JSON abi", zap.Error(err))
			return nil, err
		}
		tokenInfo, err = s.kaiClient.GetKRC20TokenInfo(ctx, &jsonABI, common.HexToAddress(krcTokenAddress))
		s.logger.Info("Update KRC20 token info", zap.Any("krc20TokenInfo", tokenInfo), zap.Error(err))
		if err != nil {
			return nil, err
		}
	} else if strings.EqualFold(krcType, cfg.SMCTypeKRC721) {
		// get KRC721 token info from RPC
		smcABIStr, err := s.dbClient.SMCABIByType(ctx, krcType)
		if err != nil {
			s.logger.Warn("Cannot get smc abi from db", zap.Error(err))
			return nil, err
		}
		abiData, err := base64.StdEncoding.DecodeString(smcABIStr)
		if err != nil {
			s.logger.Warn("Cannot decode smc abi", zap.Error(err))
			return nil, err
		}
		jsonABI, err := abi.JSON(bytes.NewReader(abiData))
		if err != nil {
			s.logger.Warn("Cannot convert decoded smc abi to JSON abi", zap.Error(err))
			return nil, err
		}
		tokenInfo, err = s.kaiClient.GetKRC721TokenInfo(ctx, &jsonABI, common.HexToAddress(krcTokenAddress))
		s.logger.Info("Update KRC721 token info", zap.Any("krc721TokenInfo", tokenInfo), zap.Error(err))
		if err != nil {
			return nil, err
		}
	}
	return tokenInfo, nil
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
	tokens, _, _ := s.dbClient.GetListHolders(ctx, &types.HolderFilter{
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

func (s *Server) calculateKRC20BalanceFloat(balance *big.Int, decimals int64) float64 {
	tenPoweredByDecimal := new(big.Int).Exp(big.NewInt(10), big.NewInt(decimals), nil)
	floatFromBalance, _ := new(big.Float).SetPrec(100).Quo(new(big.Float).SetInt(balance), new(big.Float).SetInt(tenPoweredByDecimal)).Float64()
	return floatFromBalance
}

func (s *Server) getInternalTxs(ctx context.Context, log *types.Log) *types.TokenTransfer {
	var (
		from, to, value string
		ok              bool
	)
	from, ok = log.Arguments["from"].(string)
	if !ok {
		return nil
	}
	to, ok = log.Arguments["to"].(string)
	if !ok {
		return nil
	}
	value, ok = log.Arguments["value"].(string)
	if !ok {
		return nil
	}
	// update time of internal transaction
	block, err := s.dbClient.BlockByHeight(ctx, log.BlockHeight)
	if err != nil {
		s.logger.Debug("Cannot get block from db", zap.Uint64("height", log.BlockHeight), zap.Error(err))
		block, err = s.kaiClient.BlockByHeight(ctx, log.BlockHeight)
		if err != nil {
			block = &types.Block{
				Time: time.Now(),
			}
		}
	}
	return &types.TokenTransfer{
		TransactionHash: log.TxHash,
		BlockHeight:     log.BlockHeight,
		Contract:        log.Address,
		From:            from,
		To:              to,
		Value:           value,
		Time:            block.Time,
		LogIndex:        log.Index,
	}
}

func (s *Server) fetchTokenInfo(ctx context.Context) (*types.TokenInfo, error) {
	type CMQuote struct {
		Price            float64 `json:"price"`
		Volume24h        float64 `json:"volume_24h"`
		PercentChange1h  float64 `json:"percent_change_1h"`
		PercentChange24h float64 `json:"percent_change_24h"`
		PercentChange7d  float64 `json:"percent_change_7d"`
		MarketCap        float64 `json:"market_cap"`
		LastUpdated      string  `json:"last_updated"`
	}
	type CMTokenInfo struct {
		ID                int                `json:"id"`
		Name              string             `json:"name"`
		Symbol            string             `json:"symbol"`
		Slug              string             `json:"slug"`
		NumMarketPairs    int                `json:"num_market_pairs"`
		DateAdded         string             `json:"date_added"`
		Tags              []string           `json:"tags"`
		MaxSupply         int64              `json:"max_supply"`
		CirculatingSupply int64              `json:"circulating_supply"`
		TotalSupply       int64              `json:"total_supply"`
		IsActive          int                `json:"is_active"`
		Platform          interface{}        `json:"platform"`
		CmcRank           int                `json:"cmc_rank"`
		IsFiat            int                `json:"is_fiat"`
		LastUpdated       string             `json:"last_updated"`
		Quote             map[string]CMQuote `json:"quote"`
	}

	type CMResponseStatus struct {
		Timestamp    string `json:"timestamp"`
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
		Elapsed      int    `json:"elapsed"`
		CreditCount  int    `json:"credit_count"`
		Notice       string `json:"notice"`
	}

	type CMResponse struct {
		Status CMResponseStatus       `json:"status"`
		Data   map[string]CMTokenInfo `json:"data"`
	}

	coinMarketUrl := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?id=5453"
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}

	req, _ := http.NewRequest(http.MethodGet, coinMarketUrl, nil)
	req.Header.Set("X-CMC_PRO_API_KEY", "a9aaf65c-1d6f-4daf-8e2e-df30bd405b66")

	response, _ := netClient.Do(req)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var cmResponse CMResponse

	if err := json.Unmarshal(body, &cmResponse); err != nil {
		return nil, err
	}

	if cmResponse.Status.ErrorCode != 0 {
		return nil, errors.New("api failed")
	}
	cmData := cmResponse.Data["5453"]
	cmQuote := cmData.Quote["USD"]

	// Cast to internal
	tokenInfo := &types.TokenInfo{
		Name:                     cmData.Name,
		Symbol:                   cmData.Symbol,
		Decimal:                  18,
		TotalSupply:              cmData.TotalSupply,
		ERC20CirculatingSupply:   cmData.CirculatingSupply,
		MainnetCirculatingSupply: 0,
		Price:                    cmQuote.Price,
		Volume24h:                cmQuote.Volume24h,
		Change1h:                 cmQuote.PercentChange1h,
		Change24h:                cmQuote.PercentChange24h,
		Change7d:                 cmQuote.PercentChange7d,
		MarketCap:                cmQuote.MarketCap,
	}
	if err := s.cacheClient.UpdateTokenInfo(ctx, tokenInfo); err != nil {
		return nil, err
	}

	return tokenInfo, nil
}
