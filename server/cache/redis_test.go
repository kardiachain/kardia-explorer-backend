// Package cache
package cache

import (
	"context"
	"testing"

	"gotest.tools/assert"

	"github.com/kardiachain/explorer-backend/types"
)

func TestRedis_ImportBlock(t *testing.T) {
	type Case struct {
		Input   *types.Block
		Want    *types.Block
		WantErr error
	}
	cases := map[string]Case{
		"Success": {
			Input:   nil,
			WantErr: nil,
		},
		"Failed": {
			Input:   nil,
			WantErr: nil,
		},
	}
	cache := Redis{
		client: nil,
		logger: nil,
	}
	ctx := context.Background()
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, cache.ImportBlock(ctx, c.Input), c.WantErr)
		})
	}
}
