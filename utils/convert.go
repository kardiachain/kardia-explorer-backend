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
	"strconv"

	"github.com/kardiachain/explorer-backend/types"
)

func StrToUint64(data string) uint64 {
	i, _ := strconv.ParseUint(data, 10, 64)
	return i
}

func ToAddressMap(addrs []*types.Address) map[string]*types.Address {
	addrMap := make(map[string]*types.Address)
	for _, a := range addrs {
		addrMap[a.Address] = a
	}
	return addrMap
}

func ToValidatorMap(validators []*types.Validator) map[string]*types.Validator {
	validatorMap := make(map[string]*types.Validator)
	for _, v := range validators {
		validatorMap[v.Address.Hex()] = v
	}
	return validatorMap
}
