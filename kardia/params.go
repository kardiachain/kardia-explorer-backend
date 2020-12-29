/*
 *  Copyright 2020 KardiaChain
 *  This file is part of the go-kardia library.
 *
 *  The go-kardia library is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Lesser General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  The go-kardia library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Lesser General Public License for more details.
 *
 *  You should have received a copy of the GNU Lesser General Public License
 *  along with the go-kardia library. If not, see <http://www.gnu.org/licenses/>.
 */

package kardia

import (
	"context"
	"math/big"

	"go.uber.org/zap"
)

// GetMaxProposers returns max number of proposers
func (ec *Client) GetMaxProposers(ctx context.Context) (int64, error) {
	payload, err := ec.paramsUtil.Abi.Pack("getMaxProposers")
	if err != nil {
		return 0, err
	}
	res, err := ec.KardiaCall(ctx, contructCallArgs(ec.paramsUtil.ContractAddress.Hex(), payload))
	if err != nil {
		return 0, err
	}

	var result struct {
		MaxProposers *big.Int
	}
	// unpack result
	err = ec.paramsUtil.Abi.UnpackIntoInterface(&result, "getMaxProposers", res)
	if err != nil {
		ec.lgr.Error("Error unpacking max proposers", zap.Error(err))
		return 0, err
	}
	return result.MaxProposers.Int64(), nil
}
