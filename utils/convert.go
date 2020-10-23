// Package utils
package utils

import (
	"strconv"
)

func StrToInt64(data string) int64 {
	i, _ := strconv.ParseInt(data, 10, 64)
	return i
}

func StrToUint64(data string) uint64 {
	return uint64(StrToInt64(data))
}
