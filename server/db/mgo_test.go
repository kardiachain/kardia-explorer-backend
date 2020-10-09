// Package db
package db

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

// MgoImportBlock seed size * 1000000 records into Blocks collection before run
// todo: Improve setup time for benchmark (should we ?)

func TestMgo_ImportBlock(t *testing.T) {
	type testCase struct {
	}
	cases := map[string]testCase{
		"Success":  {},
		"Failed":   {},
		"Failed 2": {},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			fmt.Printf("%#v", c)
		})
	}
}

func Test_mongoDB_AddressByHash(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	type args struct {
		ctx         context.Context
		addressHash string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.Address
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			got, err := m.AddressByHash(tt.args.ctx, tt.args.addressHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddressByHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddressByHash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mongoDB_BlockByHash(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
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
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			got, err := m.BlockByHash(tt.args.ctx, tt.args.blockHash)
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

func Test_mongoDB_BlockByHeight(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	type args struct {
		ctx         context.Context
		blockNumber uint64
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
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			got, err := m.BlockByHeight(tt.args.ctx, tt.args.blockNumber)
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

func Test_mongoDB_Blocks(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	type args struct {
		ctx        context.Context
		pagination *types.Pagination
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*types.Block
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			got, err := m.Blocks(tt.args.ctx, tt.args.pagination)
			if (err != nil) != tt.wantErr {
				t.Errorf("Blocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Blocks() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mongoDB_InsertBlock(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			if err := m.InsertBlock(tt.args.ctx, tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("InsertBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_mongoDB_InsertListTxByAddress(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	type args struct {
		ctx  context.Context
		list []*types.TransactionByAddress
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			if err := m.InsertListTxByAddress(tt.args.ctx, tt.args.list); (err != nil) != tt.wantErr {
				t.Errorf("InsertListTxByAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_mongoDB_InsertReceipts(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			if err := m.InsertReceipts(tt.args.ctx, tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("InsertReceipts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_mongoDB_InsertTxs(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	type args struct {
		ctx context.Context
		txs []*types.Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			if err := m.InsertTxs(tt.args.ctx, tt.args.txs); (err != nil) != tt.wantErr {
				t.Errorf("InsertTxs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_mongoDB_IsBlockExist(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	type args struct {
		ctx   context.Context
		block *types.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			got, err := m.IsBlockExist(tt.args.ctx, tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsBlockExist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsBlockExist() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mongoDB_OwnedTokensOfAddress(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	type args struct {
		ctx           context.Context
		walletAddress string
		pagination    *types.Pagination
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*types.TokenHolder
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			got, err := m.OwnedTokensOfAddress(tt.args.ctx, tt.args.walletAddress, tt.args.pagination)
			if (err != nil) != tt.wantErr {
				t.Errorf("OwnedTokensOfAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OwnedTokensOfAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mongoDB_TokenHolders(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	type args struct {
		ctx          context.Context
		tokenAddress string
		pagination   *types.Pagination
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*types.TokenHolder
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			got, err := m.TokenHolders(tt.args.ctx, tt.args.tokenAddress, tt.args.pagination)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenHolders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TokenHolders() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mongoDB_TxByHash(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			got, err := m.TxByHash(tt.args.ctx, tt.args.txHash)
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

func Test_mongoDB_TxByNonce(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	type args struct {
		ctx   context.Context
		nonce int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *types.Transaction
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			got, err := m.TxByNonce(tt.args.ctx, tt.args.nonce)
			if (err != nil) != tt.wantErr {
				t.Errorf("TxByNonce() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TxByNonce() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mongoDB_Txs(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	type args struct {
		ctx        context.Context
		pagination *types.Pagination
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			m.Txs(context.Background(), &types.Pagination{})
		})
	}
}

func Test_mongoDB_TxsByAddress(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	type args struct {
		ctx        context.Context
		address    string
		pagination *types.Pagination
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*types.Transaction
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			got, err := m.TxsByAddress(tt.args.ctx, tt.args.address, tt.args.pagination)
			if (err != nil) != tt.wantErr {
				t.Errorf("TxsByAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TxsByAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mongoDB_TxsByBlockHash(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	type args struct {
		ctx        context.Context
		blockHash  string
		pagination *types.Pagination
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*types.Transaction
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			got, err := m.TxsByBlockHash(tt.args.ctx, tt.args.blockHash, tt.args.pagination)
			if (err != nil) != tt.wantErr {
				t.Errorf("TxsByBlockHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TxsByBlockHash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mongoDB_TxsByBlockHeight(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	type args struct {
		ctx         context.Context
		blockHeight uint64
		pagination  *types.Pagination
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*types.Transaction
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			got, err := m.TxsByBlockHeight(tt.args.ctx, tt.args.blockHeight, tt.args.pagination)
			if (err != nil) != tt.wantErr {
				t.Errorf("TxsByBlockHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TxsByBlockHeight() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mongoDB_UpdateActiveAddresses(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	type args struct {
		ctx       context.Context
		addresses []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			if err := m.UpdateActiveAddresses(tt.args.ctx, tt.args.addresses); (err != nil) != tt.wantErr {
				t.Errorf("UpdateActiveAddresses() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_mongoDB_UpsertBlock(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			if err := m.UpsertBlock(tt.args.ctx, tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("UpsertBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_mongoDB_UpsertReceipts(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			if err := m.UpsertReceipts(tt.args.ctx, tt.args.block); (err != nil) != tt.wantErr {
				t.Errorf("UpsertReceipts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_mongoDB_UpsertTxs(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	type args struct {
		ctx context.Context
		txs []*types.Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			if err := m.UpsertTxs(tt.args.ctx, tt.args.txs); (err != nil) != tt.wantErr {
				t.Errorf("UpsertTxs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_mongoDB_dropCollection(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	type args struct {
		collectionName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			m.dropCollection(tt.name)
		})
	}
}

func Test_mongoDB_ping(t *testing.T) {
	type fields struct {
		logger  *zap.Logger
		wrapper *KaiMgo
		db      *mongo.Database
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mongoDB{
				logger:  tt.fields.logger,
				wrapper: tt.fields.wrapper,
				db:      tt.fields.db,
			}
			if err := m.ping(); (err != nil) != tt.wantErr {
				t.Errorf("ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_newMongoDB(t *testing.T) {
	type args struct {
		cfg ClientConfig
	}
	tests := []struct {
		name    string
		args    args
		want    *mongoDB
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newMongoDB(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("newMongoDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newMongoDB() got = %v, want %v", got, tt.want)
			}
		})
	}
}
