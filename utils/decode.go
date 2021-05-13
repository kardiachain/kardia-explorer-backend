package utils

import (
	"bytes"
	"encoding/base64"

	"github.com/kardiachain/go-kardia/lib/abi"
)

func DecodeSMCABIFromBase64(abiStr string) (*abi.ABI, error) {
	abiData, err := base64.StdEncoding.DecodeString(abiStr)
	if err != nil {
		return nil, err
	}
	jsonABI, err := abi.JSON(bytes.NewReader(abiData))
	if err != nil {
		return nil, err
	}
	return &jsonABI, nil
}
