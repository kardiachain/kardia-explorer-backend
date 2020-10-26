// Package db
package db

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"gotest.tools/assert"

	"github.com/kardiachain/explorer-backend/metrics"
	"github.com/kardiachain/explorer-backend/types"
)

func TestMgo_ImportBlock(t *testing.T) {
	block := &types.Block{}
	assert.NilError(t, faker.FakeData(&block))
	type testCase struct {
		block *types.Block
		err   error
	}
	cases := map[string]testCase{
		"Success": {
			block: block,
			err:   nil,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			fmt.Println(c)
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
				t.Errorf("InsertTxsOfBlock() error = %v, wantErr %v", err, tt.wantErr)
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
			got, _, err := m.OwnedTokensOfAddress(tt.args.ctx, tt.args.walletAddress, tt.args.pagination)
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
			got, _, err := m.TokenHolders(tt.args.ctx, tt.args.tokenAddress, tt.args.pagination)
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
			got, _, err := m.TxsByAddress(tt.args.ctx, tt.args.address, tt.args.pagination)
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
			got, _, err := m.TxsByBlockHash(tt.args.ctx, tt.args.blockHash, tt.args.pagination)
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
			got, _, err := m.TxsByBlockHeight(tt.args.ctx, tt.args.blockHeight, tt.args.pagination)
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
		cfg Config
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

// =================================================================================================

const (
	host   string = "mongodb://127.0.0.1:27017"
	dbName string = "explorer_benchmark"
)

func SetupMongoClient() (Client, *metrics.Provider, error) {
	logCfg := zap.NewProductionConfig()
	logger, err := logCfg.Build()
	if err != nil {
		return nil, nil, err
	}
	dbConfig := Config{
		DbAdapter: MGO,
		DbName:    dbName,
		URL:       host,
		Logger:    logger,
		MinConn:   8,
		MaxConn:   32,
		FlushDB:   true,
	}
	c, err := NewClient(dbConfig)
	if err != nil {
		return nil, nil, err
	}
	metrics := metrics.New()
	return c, metrics, nil
}

func Test_mongoDB_InsertWithMassRecord(t *testing.T) {
	numberOfBlock := []int{1000, 5000, 10000, 15000, 20000, 25000, 30000, 35000}
	numberOfTxs := []int{1000, 3000, 5000, 10000}
	for _, blockSize := range numberOfBlock {
		for _, txSize := range numberOfTxs {
			generateRecordSet(blockSize, txSize, t)
		}
	}
}

func generateRecordSet(blockSize int, txSize int, t *testing.T) {
	db, m, err := SetupMongoClient()
	if err != nil {
		t.Error("error SetupMongoClient: ", err)
		return
	}

	type BlockPrototype struct {
		Height uint64
		Hash   string
	}
	blockList := []*BlockPrototype{}
	ctx := context.Background()
	pagination := &types.Pagination{
		Skip:  0,
		Limit: 20,
	}
	// measure insert time
	startTime := time.Now()
	for i := 0; i < blockSize; i++ {
		var block types.Block
		_ = faker.FakeData(&block)
		var txs []*types.Transaction
		for j := 0; j < txSize; j++ {
			var tx types.Transaction
			_ = faker.FakeData(&tx)
			txs = append(txs, &tx)
		}
		blockList = append(blockList, &BlockPrototype{
			Height: block.Height,
			Hash:   block.Hash,
		})
		insertBlockTime := time.Now()
		_ = db.InsertBlock(ctx, &block)
		_ = db.InsertTxs(ctx, txs)
		m.RecordProcessingTime(time.Since(insertBlockTime))
	}
	t.Log("\nblockSize: ", blockSize, "\ntxSize: ", txSize, "\nElasped time: ", time.Since(startTime).String(), "\nAverage time: ", m.GetProcessingTime())

	// measure query TxsByBlockHash time
	m.Reset()
	startTime = time.Now()
	for _, bInfo := range blockList {
		queryTxsTime := time.Now()
		_, _, _ = db.TxsByBlockHash(ctx, bInfo.Hash, pagination)
		m.RecordProcessingTime(time.Since(queryTxsTime))
	}
	t.Log("\nblockSize: ", blockSize, "\nTxsByBlockHash time: ", time.Since(startTime).String(), "\nAverage time: ", m.GetProcessingTime())

	// measure query TxsByBlockHeight time
	m.Reset()
	startTime = time.Now()
	for _, bInfo := range blockList {
		queryTxsTime := time.Now()
		_, _, _ = db.TxsByBlockHeight(ctx, bInfo.Height, pagination)
		m.RecordProcessingTime(time.Since(queryTxsTime))
	}
	t.Log("\nblockSize: ", blockSize, "\nTxsByBlockHeight time: ", time.Since(startTime).String(), "\nAverage time: ", m.GetProcessingTime())

	// measure query LatestTxs with count time
	m.Reset()
	startTime = time.Now()
	for i := 0; i < blockSize; i++ {
		queryTxsTime := time.Now()
		_, _, _ = db.LatestTxs(ctx, pagination, true)
		m.RecordProcessingTime(time.Since(queryTxsTime))
	}
	t.Log("\nblockSize: ", blockSize, "\nLatestTxs with count time: ", time.Since(startTime).String(), "\nAverage time: ", m.GetProcessingTime())

	// measure query LatestTxs without count time
	m.Reset()
	startTime = time.Now()
	for i := 0; i < blockSize; i++ {
		queryTxsTime := time.Now()
		_, _, _ = db.LatestTxs(ctx, pagination, false)
		m.RecordProcessingTime(time.Since(queryTxsTime))
	}
	t.Log("\nblockSize: ", blockSize, "\nLatestTxs without count time: ", time.Since(startTime).String(), "\nAverage time: ", m.GetProcessingTime())
}
