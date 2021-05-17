// Package types
package types

import (
	"errors"
	"fmt"
	"testing"
)

func TestError(t *testing.T) {
	wrapErr := fmt.Errorf("%w", ErrABINotFound)

	if errors.Is(wrapErr, ErrABINotFound) {
		fmt.Println("ABI not found")
		return
	}

}
