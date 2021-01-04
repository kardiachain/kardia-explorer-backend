// Package utils
package utils

import (
	"github.com/kardiachain/go-kardia/types/time"
)

// GetToday return ISO format: YYYY-MM-DD
func GetToday() string {
	return time.Now().Format("2006-01-02")
}
