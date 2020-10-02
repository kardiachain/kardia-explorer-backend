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
	"errors"
	"strings"
)

func CleanUpHex(s string) string {
	s = strings.Replace(strings.TrimPrefix(s, "0x"), " ", "", -1)

	return strings.ToLower(s)
}

func ValidateAccount(accountAddress string) (string, error) {
	accountAddress = CleanUpHex(accountAddress)
	// check account length
	if len(accountAddress) != 40 {
		return "", errors.New("invalid account address")
	}
	return accountAddress, nil
}

func AppendNotEmpty(slice []string, str string) []string {
	if str != "" {
		return append(slice, str)
	}

	return slice
}
