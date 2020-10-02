// Package server
package server

import (
	"time"
)

const DefaultExpiredTime = 24 * time.Hour

const (
	KeyLatestStats  = "#stats#latest"
	KeyLatestBlocks = "#blocks#latest"
)
