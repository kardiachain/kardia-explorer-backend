// Package kardia
package kardia

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewWrapper(t *testing.T) {
	strArr := []string{}
	data := strings.Join(strArr[:], ",")
	fmt.Println("Data", data)
}
