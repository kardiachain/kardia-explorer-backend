// Package cache
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gotest.tools/assert"

	"github.com/kardiachain/explorer-backend/cfg"
	"github.com/kardiachain/explorer-backend/types"
)

func setup() (*redis.Client, *zap.Logger, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		return nil, nil, err
	}
	msg, err := redisClient.FlushAll(context.Background()).Result()
	if err != nil || msg != "OK" {
		return nil, nil, err
	}

	loggerCfg := zap.NewDevelopmentConfig()
	logger, err := loggerCfg.Build()
	if err != nil {
		return nil, nil, err
	}
	return redisClient, logger, nil
}

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

func getBlockSetup(ctx context.Context, client *redis.Client) error {
	for i := uint64(1); i <= 10; i++ {
		block := types.Block{
			Height: i,
			Hash:   strconv.FormatUint(i, 10),
		}
		blockStr, err := json.Marshal(block)
		if err != nil {
			return err
		}
		if err = client.RPush(ctx, KeyBlocks, blockStr).Err(); err != nil {
			return err
		}
	}
	return nil
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
	client, logger, err := setup()
	if err != nil {
		t.Fatalf("cannot init fields for testing")
	}
	r := fields{
		client: client,
		logger: logger,
	}
	ctx := context.Background()

	// insert test data
	getBlockSetup(ctx, r.client)
	if err != nil {
		t.Fatalf("cannot store test data to redis")
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.Block
		wantErr bool
	}{
		{
			name:   "Test_BlockInCache_1",
			fields: r,
			args: args{
				ctx:       ctx,
				blockHash: "1",
			},
			want:    &types.Block{Hash: "1", Height: 1},
			wantErr: false,
		},
		{
			name:   "Test_BlockInCache_2",
			fields: r,
			args: args{
				ctx:       ctx,
				blockHash: "6",
			},
			want:    &types.Block{Hash: "6", Height: 6},
			wantErr: false,
		},
		{
			name:   "Test_BlockNotInCache_1",
			fields: r,
			args: args{
				ctx:       ctx,
				blockHash: "11",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Test_BlockNotInCache_2",
			fields: r,
			args: args{
				ctx:       ctx,
				blockHash: "0",
			},
			want:    nil,
			wantErr: true,
		},
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
			if !isBlockEqual(got, tt.want) {
				t.Errorf("BlockByHash() got = %v, want %v", got, tt.want)
			}
		})
	}
	_, _ = r.client.FlushAll(context.Background()).Result()
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
	client, logger, err := setup()
	if err != nil {
		t.Fatalf("cannot init fields for testing")
	}
	r := fields{
		client: client,
		logger: logger,
	}
	ctx := context.Background()

	// insert test data
	err = getBlockSetup(ctx, r.client)
	if err != nil {
		t.Fatalf("cannot store test data to redis")
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.Block
		wantErr bool
	}{
		{
			name:   "Test_BlockInCache_1",
			fields: r,
			args: args{
				ctx:         ctx,
				blockHeight: 1,
			},
			want:    &types.Block{Hash: "1", Height: 1},
			wantErr: false,
		},
		{
			name:   "Test_BlockInCache_2",
			fields: r,
			args: args{
				ctx:         ctx,
				blockHeight: 6,
			},
			want:    &types.Block{Hash: "6", Height: 6},
			wantErr: false,
		},
		{
			name:   "Test_BlockNotInCache_1",
			fields: r,
			args: args{
				ctx:         ctx,
				blockHeight: 11,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Test_BlockNotInCache_2",
			fields: r,
			args: args{
				ctx:         ctx,
				blockHeight: 0,
			},
			want:    nil,
			wantErr: true,
		},
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
			if !isBlockEqual(got, tt.want) {
				t.Errorf("BlockByHeight() got = %v, want %v", got, tt.want)
			}
		})
	}
	_, _ = r.client.FlushAll(context.Background()).Result()
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

func getTxsByBlockSetup(ctx context.Context, client *redis.Client) error {
	key := fmt.Sprintf(KeyTxsOfBlockHeight, 0)
	for i := 0; i < 10; i++ {
		tx := types.Transaction{
			BlockNumber: 0,
			BlockHash:   "0xhash0",
			Hash:        strconv.FormatInt(int64(i), 10),
		}
		txStr, err := json.Marshal(tx)
		if err != nil {
			return err
		}
		if err = client.RPush(ctx, key, txStr).Err(); err != nil {
			return err
		}
	}
	keyBlockHashByHeight := fmt.Sprintf(KeyBlockHashByHeight, "0xhash0")
	if err := client.Set(ctx, keyBlockHashByHeight, 0, cfg.BlockHeightByHashExpTime).Err(); err != nil {
		return err
	}
	return nil
}

func TestRedis_TxsByBlockHeight(t *testing.T) {
	type fields struct {
		client *redis.Client
		logger *zap.Logger
	}
	type args struct {
		ctx         context.Context
		blockHeight uint64
		pagination  *types.Pagination
	}
	client, logger, err := setup()
	if err != nil {
		t.Fatalf("cannot init fields for testing")
	}
	r := fields{
		client: client,
		logger: logger,
	}
	ctx := context.Background()

	// insert test data
	err = getTxsByBlockSetup(ctx, r.client)
	if err != nil {
		t.Fatalf("cannot store test data to redis")
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*types.Transaction
		wantErr bool
	}{
		{
			name:   "Test_BlockInCache_ProperPagination",
			fields: r,
			args: args{
				ctx:         ctx,
				blockHeight: 0,
				pagination: &types.Pagination{
					Skip:  0,
					Limit: 5,
				},
			},
			want: []*types.Transaction{
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "0"},
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "1"},
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "2"},
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "3"},
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "4"},
			},
			wantErr: false,
		},
		{
			name:   "Test_BlockInCache_ImproperPagination_1",
			fields: r,
			args: args{
				ctx:         ctx,
				blockHeight: 0,
				pagination: &types.Pagination{
					Skip:  10,
					Limit: 5,
				},
			},
			want:    []*types.Transaction(nil),
			wantErr: false,
		},
		{
			name:   "Test_BlockInCache_ImproperPagination_2",
			fields: r,
			args: args{
				ctx:         ctx,
				blockHeight: 0,
				pagination: &types.Pagination{
					Skip:  5,
					Limit: 10,
				},
			},
			want: []*types.Transaction{
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "5"},
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "6"},
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "7"},
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "8"},
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "9"},
			},
			wantErr: false,
		},
		{
			name:   "Test_BlockNotInCache",
			fields: r,
			args: args{
				ctx:         ctx,
				blockHeight: 1,
				pagination: &types.Pagination{
					Skip:  0,
					Limit: 5,
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Redis{
				client: tt.fields.client,
				logger: tt.fields.logger,
			}
			txsList, _, err := c.TxsByBlockHeight(tt.args.ctx, tt.args.blockHeight, tt.args.pagination)
			if (err != nil) != tt.wantErr {
				t.Errorf("TxsByBlockHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !isTxsListEqual(txsList, tt.want) {
				t.Errorf("TxsByBlockHeight() got = %v, want %v", txsList, tt.want)
			}
		})
	}
	_, _ = r.client.FlushAll(context.Background()).Result()
}

func TestRedis_TxsByBlockHash(t *testing.T) {
	type fields struct {
		client *redis.Client
		logger *zap.Logger
	}
	type args struct {
		ctx        context.Context
		blockHash  string
		pagination *types.Pagination
	}
	client, logger, err := setup()
	if err != nil {
		t.Fatalf("cannot init fields for testing")
	}
	r := fields{
		client: client,
		logger: logger,
	}
	ctx := context.Background()

	// insert test data
	err = getTxsByBlockSetup(ctx, r.client)
	if err != nil {
		t.Fatalf("cannot store test data to redis")
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*types.Transaction
		wantErr bool
	}{
		{
			name:   "Test_BlockInCache_ProperPagination",
			fields: r,
			args: args{
				ctx:       ctx,
				blockHash: "0xhash0",
				pagination: &types.Pagination{
					Skip:  0,
					Limit: 5,
				},
			},
			want: []*types.Transaction{
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "0"},
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "1"},
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "2"},
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "3"},
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "4"},
			},
			wantErr: false,
		},
		{
			name:   "Test_BlockInCache_ImproperPagination_1",
			fields: r,
			args: args{
				ctx:       ctx,
				blockHash: "0xhash0",
				pagination: &types.Pagination{
					Skip:  10,
					Limit: 5,
				},
			},
			want:    []*types.Transaction(nil),
			wantErr: false,
		},
		{
			name:   "Test_BlockInCache_ImproperPagination_2",
			fields: r,
			args: args{
				ctx:       ctx,
				blockHash: "0xhash0",
				pagination: &types.Pagination{
					Skip:  5,
					Limit: 10,
				},
			},
			want: []*types.Transaction{
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "5"},
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "6"},
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "7"},
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "8"},
				{BlockNumber: 0, BlockHash: "0xhash0", Hash: "9"},
			},
			wantErr: false,
		},
		{
			name:   "Test_BlockNotInCache",
			fields: r,
			args: args{
				ctx:       ctx,
				blockHash: "0xhash1",
				pagination: &types.Pagination{
					Skip:  0,
					Limit: 5,
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Redis{
				client: tt.fields.client,
				logger: tt.fields.logger,
			}
			txsList, _, err := c.TxsByBlockHash(tt.args.ctx, tt.args.blockHash, tt.args.pagination)
			if (err != nil) != tt.wantErr {
				t.Errorf("TxsByBlockHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !isTxsListEqual(txsList, tt.want) {
				t.Errorf("TxsByBlockHeight() got = %v, want %v", txsList, tt.want)
			}
		})
	}
	_, _ = r.client.FlushAll(context.Background()).Result()
}

func isTxsListEqual(src []*types.Transaction, dest []*types.Transaction) bool {
	if len(src) != len(dest) {
		return false
	}
	for i := range src {
		if (src[i].Hash != dest[i].Hash) || (src[i].BlockHash != dest[i].BlockHash) || (src[i].BlockNumber != dest[i].BlockNumber) {
			return false
		}
	}
	return true
}

func isBlockEqual(src *types.Block, dest *types.Block) bool {
	return (src == nil && dest == nil) || (src.Hash == dest.Hash && src.Height == dest.Height)
}
