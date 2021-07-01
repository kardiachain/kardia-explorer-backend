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
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math/big"
	"strconv"
	"strings"

	"github.com/chai2010/webp"
)

var Hydro = big.NewInt(1000000000000000000)

func StrToUint64(data string) uint64 {
	i, _ := strconv.ParseUint(data, 10, 64)
	return i
}

func StrToInt64(data string) int64 {
	i, _ := strconv.ParseInt(data, 10, 64)
	return i
}

func BalanceToFloat(balance string) float64 {
	balanceBI, _ := new(big.Int).SetString(balance, 10)
	balanceF, _ := new(big.Float).SetPrec(1000000).Quo(new(big.Float).SetInt(balanceBI), new(big.Float).SetInt(Hydro)).Float64() //converting to KAI from HYDRO
	return balanceF
}

func BalanceToFloatWithDecimals(balance *big.Int, decimals int64) float64 {
	tenPoweredByDecimal := new(big.Int).Exp(big.NewInt(10), big.NewInt(decimals), nil)
	floatFromBalance, _ := new(big.Float).SetPrec(100).Quo(new(big.Float).SetInt(balance), new(big.Float).SetInt(tenPoweredByDecimal)).Float64()
	return floatFromBalance
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

func Base64ToImage(rawString string) (image.Image, error) {
	var unbased []byte
	var imageDecode image.Image
	var errImage error
	switch {
	case strings.Contains(rawString, "data:image/png;base64,"):
		rawString = strings.ReplaceAll(rawString, "data:image/png;base64,", "")
		unbased, _ = base64.StdEncoding.DecodeString(string(rawString))
		imageDecode, errImage = png.Decode(bytes.NewReader(unbased))
		if errImage != nil {
			return nil, errImage
		}
		break
	case strings.Contains(rawString, "data:image/jpeg;base64,"):
		rawString = strings.ReplaceAll(rawString, "data:image/jpeg;base64,", "")
		unbased, _ = base64.StdEncoding.DecodeString(string(rawString))
		imageDecode, errImage = jpeg.Decode(bytes.NewReader(unbased))
		if errImage != nil {
			return nil, errImage
		}
		break
	case strings.Contains(rawString, "data:image/webp;base64"):
		rawString = strings.ReplaceAll(rawString, "data:image/webp;base64,", "")
		unbased, _ = base64.StdEncoding.DecodeString(string(rawString))
		imageDecode, errImage = webp.Decode(bytes.NewReader(unbased))
		if errImage != nil {
			return nil, errImage
		}
		break
	default:
		break
	}

	return imageDecode, nil
}

func EncodeImage(image image.Image, rawString string, fileName string) ([]byte, string) {
	switch {
	case strings.Contains(rawString, "data:image/png;base64,"):
		buf := new(bytes.Buffer)
		errConverter := png.Encode(buf, image)
		if errConverter != nil {
			return nil, ""
		}
		sendS3 := buf.Bytes()
		spl := strings.Split(fileName+".png", ".")
		uploadedFileName := strings.Join(spl, ".")
		return sendS3, uploadedFileName
	case strings.Contains(rawString, "data:image/jpeg;base64,"):
		buf := new(bytes.Buffer)
		errConverter := jpeg.Encode(buf, image, nil)
		if errConverter != nil {
			return nil, ""
		}
		sendS3 := buf.Bytes()
		spl := strings.Split(fileName+".png", ".")
		uploadedFileName := strings.Join(spl, ".")
		return sendS3, uploadedFileName
	case strings.Contains(rawString, "data:image/webp;base64"):
		buf := new(bytes.Buffer)
		errConverter := webp.Encode(buf, image, nil)
		if errConverter != nil {
			return nil, ""
		}
		sendS3 := buf.Bytes()
		spl := strings.Split(fileName+".webp", ".")
		uploadedFileName := strings.Join(spl, ".")
		return sendS3, uploadedFileName
	}

	return nil, ""
}

func HashString(name string) string {
	h := sha1.New()
	h.Write([]byte(name))
	return hex.EncodeToString(h.Sum(nil))
}
