package kardia

import (
	"context"
	"math/big"

	"go.uber.org/zap"

	"github.com/kardiachain/go-kardia/lib/abi"
	"github.com/kardiachain/go-kardia/lib/common"
)

// GetKRCTotalSupply returns total supply of a KRC token
func (ec *Client) GetKRCTotalSupply(ctx context.Context, a *abi.ABI, krcTokenAddr common.Address) (*big.Int, error) {
	payload, err := a.Pack("totalSupply")
	if err != nil {
		ec.lgr.Error("Error packing get total supply payload: ", zap.Error(err))
		return nil, err
	}

	res, err := ec.KardiaCall(ctx, contructCallArgs(krcTokenAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Warn("GetKRCTotalSupply KardiaCall error: ", zap.Error(err))
		return nil, err
	}
	if len(res) == 0 {
		return nil, ErrEmptyList
	}

	var totalSupply *big.Int
	// unpack result
	err = ec.validatorUtil.Abi.UnpackIntoInterface(&totalSupply, "totalSupply", res)
	if err != nil {
		ec.lgr.Error("Error unpacking get total supply error: ", zap.Error(err))
		return nil, err
	}
	return totalSupply, nil
}

// GetKRCBalanceByAddress returns balance of a KRC holder
func (ec *Client) GetKRCBalanceByAddress(ctx context.Context, a *abi.ABI, krcTokenAddr common.Address, holder common.Address) (*big.Int, error) {
	payload, err := a.Pack("balanceOf")
	if err != nil {
		ec.lgr.Error("Error packing get balance payload: ", zap.Error(err))
		return nil, err
	}

	res, err := ec.KardiaCall(ctx, contructCallArgs(krcTokenAddr.Hex(), payload))
	if err != nil {
		ec.lgr.Warn("GetKRCBalanceByAddress KardiaCall error: ", zap.Error(err))
		return nil, err
	}
	if len(res) == 0 {
		return nil, ErrEmptyList
	}

	var balance *big.Int
	// unpack result
	err = ec.validatorUtil.Abi.UnpackIntoInterface(&balance, "balanceOf", res)
	if err != nil {
		ec.lgr.Error("Error unpacking get balance error: ", zap.Error(err))
		return nil, err
	}
	return balance, nil
}
