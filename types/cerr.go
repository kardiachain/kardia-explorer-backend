// Package contracts
package types

import (
	"errors"
)

var ErrRecordExist = errors.New("block exist")
var ErrABINotFound = errors.New("abi not found")
var ErrSMCTypeNormal = errors.New("abi type normal")
