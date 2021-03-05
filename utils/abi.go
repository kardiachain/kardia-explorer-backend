// Package utils
package utils

import (
	"encoding/base64"
)

func EncodeABI(abi string) string {
	return base64.StdEncoding.EncodeToString([]byte(abi))
}

func DecodeABI(encodedABI string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encodedABI)
}
