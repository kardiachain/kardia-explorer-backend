// Package utils
package utils

import (
	"math/rand"
)

func RandInRange(min, max int) int {
	return rand.Intn(max-min) + min
}
