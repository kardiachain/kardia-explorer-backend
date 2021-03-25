package kardia

import (
	"context"
	"math/big"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"

	"go.uber.org/zap"

	"github.com/kardiachain/go-kardia/lib/abi"
	"github.com/kardiachain/go-kardia/lib/common"
)

// getKRC20TotalSupply returns total supply of a KRC token
func (ec *Client) getKRC20TotalSupply(ctx context.Context, a *abi.ABI, krcTokenAddr common.Address) (*big.Int, error) {
	payload, err := a.Pack("totalSupply")
	if err != nil {
		ec.lgr.Error("Error packing get total supply payload: ", zap.Error(err))
		return nil, err
	}

	res, err := ec.KardiaCall(ctx, contructCallArgs(krcTokenAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Warn("getKRC20TotalSupply KardiaCall error: ", zap.Error(err))
		return nil, err
	}
	if len(res) == 0 {
		return nil, ErrEmptyList
	}

	var totalSupply *big.Int
	// unpack result
	err = a.UnpackIntoInterface(&totalSupply, "totalSupply", res)
	if err != nil {
		ec.lgr.Error("Error unpacking total supply: ", zap.Error(err))
		return nil, err
	}
	return totalSupply, nil
}

// GetKRC20BalanceByAddress returns balance of a KRC holder
func (ec *Client) GetKRC20BalanceByAddress(ctx context.Context, a *abi.ABI, krcTokenAddr common.Address, holder common.Address) (*big.Int, error) {
	payload, err := a.Pack("balanceOf", holder)
	if err != nil {
		ec.lgr.Error("Error packing get balance payload: ", zap.Error(err))
		return nil, err
	}

	res, err := ec.KardiaCall(ctx, contructCallArgs(krcTokenAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Warn("GetKRC20BalanceByAddress KardiaCall error: ", zap.Error(err))
		return nil, err
	}

	var balance *big.Int
	// unpack result
	err = a.UnpackIntoInterface(&balance, "balanceOf", res)
	if err != nil {
		ec.lgr.Error("Error unpacking balance: ", zap.Error(err))
		return nil, err
	}
	return balance, nil
}

// getKRC20TokenDecimal
func (ec *Client) getKRC20TokenDecimal(ctx context.Context, a *abi.ABI, krcTokenAddr common.Address) (uint8, error) {
	payload, err := a.Pack("decimals")
	if err != nil {
		ec.lgr.Error("Error packing get decimals payload: ", zap.Error(err))
		return 0, err
	}

	res, err := ec.KardiaCall(ctx, contructCallArgs(krcTokenAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Warn("getKRC20TokenDecimal KardiaCall error: ", zap.Error(err))
		return 0, err
	}
	if len(res) == 0 {
		return 0, ErrEmptyList
	}

	var decimals uint8
	// unpack result
	err = a.UnpackIntoInterface(&decimals, "decimals", res)
	if err != nil {
		ec.lgr.Error("Error unpacking decimals: ", zap.Error(err))
		return 0, err
	}
	return decimals, nil
}

// getKRC20TokenName
func (ec *Client) getKRC20TokenName(ctx context.Context, a *abi.ABI, krcTokenAddr common.Address) (string, error) {
	payload, err := a.Pack("name")
	if err != nil {
		ec.lgr.Error("Error packing get token name payload: ", zap.Error(err))
		return "", err
	}

	res, err := ec.KardiaCall(ctx, contructCallArgs(krcTokenAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Warn("getKRC20TokenName KardiaCall error: ", zap.Error(err))
		return "", err
	}
	if len(res) == 0 {
		return "", ErrEmptyList
	}

	var tokenName string
	// unpack result
	err = a.UnpackIntoInterface(&tokenName, "name", res)
	if err != nil {
		ec.lgr.Error("Error unpacking token name: ", zap.Error(err))
		return "", err
	}
	return tokenName, nil
}

// getKRC20TokenSymbol
func (ec *Client) getKRC20TokenSymbol(ctx context.Context, a *abi.ABI, krcTokenAddr common.Address) (string, error) {
	payload, err := a.Pack("symbol")
	if err != nil {
		ec.lgr.Error("Error packing token symbol payload: ", zap.Error(err))
		return "", err
	}

	res, err := ec.KardiaCall(ctx, contructCallArgs(krcTokenAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Warn("getKRC20TokenSymbol KardiaCall error: ", zap.Error(err))
		return "", err
	}
	if len(res) == 0 {
		return "", ErrEmptyList
	}

	var tokenSymbol string
	// unpack result
	err = a.UnpackIntoInterface(&tokenSymbol, "symbol", res)
	if err != nil {
		ec.lgr.Error("Error unpacking token symbol: ", zap.Error(err))
		return "", err
	}
	return tokenSymbol, nil
}

func (ec *Client) GetKRC20TokenInfo(ctx context.Context, a *abi.ABI, krcTokenAddr common.Address) (*types.KRCTokenInfo, error) {
	name, err := ec.getKRC20TokenName(ctx, a, krcTokenAddr)
	if err != nil {
		ec.lgr.Error("Cannot get KRC20 token name", zap.String("smcAddress", krcTokenAddr.Hex()), zap.Error(err))
		return nil, err
	}
	symbol, err := ec.getKRC20TokenSymbol(ctx, a, krcTokenAddr)
	if err != nil {
		ec.lgr.Error("Cannot get KRC20 token symbol", zap.String("smcAddress", krcTokenAddr.Hex()), zap.Error(err))
		return nil, err
	}
	totalSupply, err := ec.getKRC20TotalSupply(ctx, a, krcTokenAddr)
	if err != nil {
		ec.lgr.Error("Cannot get KRC20 token total supply", zap.String("smcAddress", krcTokenAddr.Hex()), zap.Error(err))
		return nil, err
	}
	decimals, err := ec.getKRC20TokenDecimal(ctx, a, krcTokenAddr)
	if err != nil {
		ec.lgr.Error("Cannot get KRC20 token decimals", zap.String("smcAddress", krcTokenAddr.Hex()), zap.Error(err))
		return nil, err
	}
	return &types.KRCTokenInfo{
		Address:     krcTokenAddr.Hex(),
		TokenName:   name,
		TokenType:   cfg.SMCTypeKRC20,
		TokenSymbol: symbol,
		TotalSupply: totalSupply.String(),
		Decimals:    int64(decimals),
	}, nil
}
