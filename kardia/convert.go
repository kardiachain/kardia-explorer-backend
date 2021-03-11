// Package kardia
package kardia

import (
	"fmt"
	"math/big"
	"strings"
)

func validatorNameInString(data [32]byte) string {
	var name []byte
	for _, b := range data {
		if b != 0 {
			name = append(name, b)
		}
	}
	return string(name)
}

func convertBigIntToPercentage(raw string) (string, error) {
	input, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return "", ErrParsingBigIntFromString
	}
	tmp := new(big.Int).Mul(input, tenPoweredBy18)
	result := new(big.Int).Div(tmp, tenPoweredBy18).String()
	result = fmt.Sprintf("%020s", result)
	result = strings.TrimLeft(strings.TrimRight(strings.TrimRight(result[:len(result)-16]+"."+result[len(result)-16:], "0"), "."), "0")
	if strings.HasPrefix(result, ".") {
		result = "0" + result
	}
	return result, nil
}

func calculateVotingPower(raw string, total *big.Int) (string, error) {

	valStakedAmount, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return "", ErrParsingBigIntFromString
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
