// Package utils
package utils

import (
	"strconv"
)

func StrToUint64(data string) uint64 {
	i, _ := strconv.ParseUint(data, 10, 64)
	return i
}
