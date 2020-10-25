// Package cache
package cache

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
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
			assert.Equal(t, cache.InsertBlock(ctx, c.Input), c.WantErr)
		})
	}
}

func TestRedis_BlockByHash(t *testing.T) {
	type fields struct {
		client *redis.Client
		logger *zap.Logger
	}
	type args struct {
		ctx       context.Context
		blockHash string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.Block
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Redis{
				client: tt.fields.client,
				logger: tt.fields.logger,
			}
			got, err := c.BlockByHash(tt.args.ctx, tt.args.blockHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockByHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockByHash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedis_BlockByHeight(t *testing.T) {
	type fields struct {
		client *redis.Client
		logger *zap.Logger
	}
	type args struct {
		ctx         context.Context
		blockHeight uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.Block
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Redis{
				client: tt.fields.client,
				logger: tt.fields.logger,
			}
			got, err := c.BlockByHeight(tt.args.ctx, tt.args.blockHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockByHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockByHeight() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedis_InsertBlock(t *testing.T) {
	type fields struct {
		client *redis.Client
		logger *zap.Logger
	}
	type args struct {
		ctx   context.Context
		block *types.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Redis{
				client: tt.fields.client,
				logger: tt.fields.logger,
			}
			if err := c.InsertBlock(tt.args.ctx, tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("InsertBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRedis_InsertTxs(t *testing.T) {
	type fields struct {
		client *redis.Client
		logger *zap.Logger
	}
	type args struct {
		ctx   context.Context
		block *types.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Redis{
				client: tt.fields.client,
				logger: tt.fields.logger,
			}
			if err := c.InsertTxsOfBlock(tt.args.ctx, tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("InsertTxsOfBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRedis_TxByHash(t *testing.T) {
	type fields struct {
		client *redis.Client
		logger *zap.Logger
	}
	type args struct {
		ctx    context.Context
		txHash string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.Transaction
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Redis{
				client: tt.fields.client,
				logger: tt.fields.logger,
			}
			got, err := c.TxByHash(tt.args.ctx, tt.args.txHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("TxByHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TxByHash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedis_getBlockIndex(t *testing.T) {
	type fields struct {
		client *redis.Client
		logger *zap.Logger
	}
	type args struct {
		ctx   context.Context
		index int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.Block
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Redis{
				client: tt.fields.client,
				logger: tt.fields.logger,
			}
			got, err := c.getBlockIndex(tt.args.ctx, tt.args.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBlockIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getBlockIndex() got = %v, want %v", got, tt.want)
			}
		})
	}
}
