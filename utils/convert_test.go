// Package utils
package utils

import (
	"fmt"
	"testing"
)

func TestBalanceToFloat(t *testing.T) {
	type input struct {
		balance string
	}
	type output struct {
		balance float64
	}

	f := BalanceToFloat("999999999999427946")
	fmt.Println("Float", f)
}
