/*
 *  Copyright 2018 KardiaChain
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
// Package utils
package utils

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

var Hydro = big.NewInt(1000000000000000000)

func StrToUint64(data string) uint64 {
	i, _ := strconv.ParseUint(data, 10, 64)
	return i
}

func BalanceToFloat(balance string) float64 {
	balanceBI, _ := new(big.Int).SetString(balance, 10)
	balanceF, _ := new(big.Float).SetPrec(1000000).Quo(new(big.Float).SetInt(balanceBI), new(big.Float).SetInt(Hydro)).Float64() //converting to KAI from HYDRO
	return balanceF
}

func CalculateVotingPower(raw string, total *big.Int) (string, error) {
	var (
		tenPoweredBy5 = new(big.Int).Exp(big.NewInt(10), big.NewInt(5), nil)
	)
	valStakedAmount, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return "", fmt.Errorf("cannot convert from string to *big.Int")
	}
	tmp := new(big.Int).Mul(valStakedAmount, tenPoweredBy5)
	result := new(big.Int).Div(tmp, total).String()
	result = fmt.Sprintf("%020s", result)
	result = strings.TrimLeft(strings.TrimRight(strings.TrimRight(result[:len(result)-3]+"."+result[len(result)-3:], "0"), "."), "0")
	if strings.HasPrefix(result, ".") {
		result = "0" + result
	}
	return result, nil
}
