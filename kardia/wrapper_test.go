package kardia

import (
	"context"
	"fmt"
	"github.com/kardiachain/go-kaiclient/kardia"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestWrapper_validatorWithNode(t *testing.T) {
	lgr, _ := zap.NewDevelopment()
	node, err := kardia.NewNode("https://rpc.kardiachain.io", lgr)
	assert.Nil(t, err)
	w := Wrapper{
		trustedNodes: []kardia.Node{node},
		publicNodes:  []kardia.Node{node},
		wsNodes:      []kardia.Node{node},
		logger:       lgr,
	}
	info, err := w.validatorWithNode(context.Background(), "0xdC4A94805f449A64B27B589233C49d87eE99fBBc", node)
	assert.Nil(t, err)
	fmt.Printf("Data: %+v \n", info)
}
