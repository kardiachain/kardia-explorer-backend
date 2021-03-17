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

// getKRC721TotalSupply returns total supply of a KRC token
func (ec *Client) getKRC721TotalSupply(ctx context.Context, a *abi.ABI, krcTokenAddr common.Address) (*big.Int, error) {
	payload, err := a.Pack("totalSupply")
	if err != nil {
		ec.lgr.Error("Error packing get total supply payload: ", zap.Error(err))
		return nil, err
	}

	res, err := ec.KardiaCall(ctx, contructCallArgs(krcTokenAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Warn("getKRC721TotalSupply KardiaCall error: ", zap.Error(err))
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

// getKRC721TotalToken
func (ec *Client) getKRC721TotalToken(ctx context.Context, a *abi.ABI, krcTokenAddr common.Address) (*big.Int, error) {
	payload, err := a.Pack("numTokensTotal")
	if err != nil {
		ec.lgr.Error("Error packing get decimals payload: ", zap.Error(err))
		return nil, err
	}

	res, err := ec.KardiaCall(ctx, contructCallArgs(krcTokenAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Warn("getKRC721TotalToken KardiaCall error: ", zap.Error(err))
		return nil, err
	}
	if len(res) == 0 {
		return nil, ErrEmptyList
	}

	var numTokensTotal *big.Int
	// unpack result
	err = a.UnpackIntoInterface(&numTokensTotal, "numTokensTotal", res)
	if err != nil {
		ec.lgr.Error("Error unpacking decimals: ", zap.Error(err))
		return nil, err
	}
	return numTokensTotal, nil
}

// getKRC721TokenName
func (ec *Client) getKRC721TokenName(ctx context.Context, a *abi.ABI, krcTokenAddr common.Address) (string, error) {
	payload, err := a.Pack("name")
	if err != nil {
		ec.lgr.Error("Error packing get token name payload: ", zap.Error(err))
		return "", err
	}

	res, err := ec.KardiaCall(ctx, contructCallArgs(krcTokenAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Warn("getKRC721TokenName KardiaCall error: ", zap.Error(err))
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

// getKRC721TokenSymbol
func (ec *Client) getKRC721TokenSymbol(ctx context.Context, a *abi.ABI, krcTokenAddr common.Address) (string, error) {
	payload, err := a.Pack("symbol")
	if err != nil {
		ec.lgr.Error("Error packing token symbol payload: ", zap.Error(err))
		return "", err
	}

	res, err := ec.KardiaCall(ctx, contructCallArgs(krcTokenAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Warn("getKRC721TokenSymbol KardiaCall error: ", zap.Error(err))
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

func (ec *Client) GetKRC721TokenInfo(ctx context.Context, a *abi.ABI, krcTokenAddr common.Address) (*types.KRCTokenInfo, error) {
	name, err := ec.getKRC721TokenName(ctx, a, krcTokenAddr)
	if err != nil {
		return nil, err
	}
	symbol, err := ec.getKRC721TokenSymbol(ctx, a, krcTokenAddr)
	if err != nil {
		return nil, err
	}
	totalSupply, err := ec.getKRC721TotalSupply(ctx, a, krcTokenAddr)
	if err != nil {
		return nil, err
	}
	totalTokens, err := ec.getKRC721TotalToken(ctx, a, krcTokenAddr)
	if err != nil {
		return nil, err
	}
	return &types.KRCTokenInfo{
		Address:        krcTokenAddr.Hex(),
		TokenName:      name,
		TokenType:      cfg.SMCTypeKRC721,
		TokenSymbol:    symbol,
		TotalSupply:    totalSupply.String(),
		NumTokensTotal: totalTokens.Int64(),
	}, nil
}
