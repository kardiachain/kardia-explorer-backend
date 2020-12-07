// Package server
package server

import (
	"context"
	"reflect"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/kardiachain/explorer-backend/kardia"
	"github.com/kardiachain/explorer-backend/metrics"
	"github.com/kardiachain/explorer-backend/server/cache"
	"github.com/kardiachain/explorer-backend/server/db"
	"github.com/kardiachain/explorer-backend/types"
)

func TestDB_InsertWithMassRecord(t *testing.T) {
	numberOfBlock := []int{1000, 5000, 10000, 15000, 20000, 25000, 30000, 35000}
	numberOfTxs := []int{1000, 3000, 5000, 10000}
	for _, blockSize := range numberOfBlock {
		for _, txSize := range numberOfTxs {
			generateRecordSet(blockSize, txSize)
		}
	}
}

func generateRecordSet(blockSize int, txSize int) {

}

func TestDB_UsingPG(t *testing.T) {

}

func TestDB_UsingMgo(t *testing.T) {

}

var (
	block1 = &types.Block{
		Height: 1,
		Hash:   "0xhash1",
		NumTxs: 5,
		Txs: []*types.Transaction{
			{BlockNumber: 1, BlockHash: "0xhash1", Hash: "0", TransactionIndex: 0},
			{BlockNumber: 1, BlockHash: "0xhash1", Hash: "1", TransactionIndex: 1},
			{BlockNumber: 1, BlockHash: "0xhash1", Hash: "2", TransactionIndex: 2},
			{BlockNumber: 1, BlockHash: "0xhash1", Hash: "3", TransactionIndex: 3},
			{BlockNumber: 1, BlockHash: "0xhash1", Hash: "4", TransactionIndex: 4},
		},
	}
	block2 = &types.Block{
		Height: 2,
		Hash:   "0xhash2",
		NumTxs: 5,
		Txs: []*types.Transaction{
			{BlockNumber: 2, BlockHash: "0xhash2", Hash: "5", TransactionIndex: 5},
			{BlockNumber: 2, BlockHash: "0xhash2", Hash: "6", TransactionIndex: 6},
			{BlockNumber: 2, BlockHash: "0xhash2", Hash: "7", TransactionIndex: 7},
			{BlockNumber: 2, BlockHash: "0xhash2", Hash: "8", TransactionIndex: 8},
			{BlockNumber: 2, BlockHash: "0xhash2", Hash: "9", TransactionIndex: 9},
		},
	}
)

func setup() (db.Client, cache.Client, kardia.ClientInterface, *metrics.Provider, string, *zap.Logger, error) {
	logCfg := zap.NewDevelopmentConfig()
	logCfg.Level.SetLevel(zapcore.DebugLevel)
	logger, err := logCfg.Build()
	if err != nil {
		return nil, nil, nil, nil, "", nil, err
	}

	dbConfig := db.Config{
		DbAdapter: db.MGO,
		DbName:    "explorerDB_test",
		URL:       "mongodb://localhost:27017",
		Logger:    logger,
		MinConn:   1,
		MaxConn:   50,

		FlushDB: true,
	}
	dbClient, err := db.NewClient(dbConfig)

	cacheCfg := cache.Config{
		Adapter:     cache.RedisAdapter,
		URL:         "localhost:6379",
		DB:          0,
		IsFlush:     true,
		BlockBuffer: 20,
		Logger:      logger,
	}
	cacheClient, err := cache.New(cacheCfg)
	if err != nil {
		return nil, nil, nil, nil, "", nil, err
	}
	avgMetrics := metrics.New()

	return dbClient, cacheClient, nil, avgMetrics, "secret", logger, nil
}

func Test_infoServer_VerifyBlock(t *testing.T) {
	type fields struct {
		dbClient          db.Client
		cacheClient       cache.Client
		kaiClient         kardia.ClientInterface
		metrics           *metrics.Provider
		HttpRequestSecret string
		logger            *zap.Logger
	}
	type args struct {
		ctx               context.Context
		blockHeight       uint64
		verifier          VerifyBlockStrategy
		needToInsertBlock *types.Block
		compareBlock      *types.Block
	}
	dbClient, cacheClient, _, avgMetrics, secret, logger, err := setup()
	if err != nil {
		t.Fatalf("cannot init fields for testing")
	}
	f := fields{
		dbClient:          dbClient,
		cacheClient:       cacheClient,
		kaiClient:         nil,
		metrics:           avgMetrics,
		HttpRequestSecret: secret,
		logger:            logger,
	}
	ctx := context.Background()
	verifyBlockParam := types.VerifyBlockParam{
		VerifyBlockHash: false,
		VerifyTxCount:   true,
	}
	verifier := func(db, network *types.Block) bool {
		if verifyBlockParam.VerifyTxCount {
			if db.NumTxs != network.NumTxs {
				return false
			}
		}
		if verifyBlockParam.VerifyBlockHash {
			return true
		}
		return true
	}
	var (
		corruptedBlock1 = &types.Block{
			Height: 1,
			Hash:   "0xhash1",
			NumTxs: 5,
			Txs: []*types.Transaction{
				{BlockNumber: 1, BlockHash: "0xhash1", Hash: "0", TransactionIndex: 0},
			{BlockNumber: 1, BlockHash: "0xhash1", Hash: "1", TransactionIndex: 1},
				{BlockNumber: 1, BlockHash: "0xhash1", Hash: "3", TransactionIndex: 3},
				{BlockNumber: 1, BlockHash: "0xhash1", Hash: "4", TransactionIndex: 4},
			},
		}
		corruptedBlock2 = &types.Block{
			Height: 2,
			Hash:   "0xhash2",
			NumTxs: 5,
			Txs: []*types.Transaction{
				{BlockNumber: 2, BlockHash: "0xhash2", Hash: "5", TransactionIndex: 5},
				{BlockNumber: 2, BlockHash: "0xhash2", Hash: "6", TransactionIndex: 6},
				{BlockNumber: 2, BlockHash: "0xhash2", Hash: "7", TransactionIndex: 7},
				{BlockNumber: 2, BlockHash: "0xhash2", Hash: "8", TransactionIndex: 8},
				{BlockNumber: 2, BlockHash: "0xhash2", Hash: "9", TransactionIndex: 9},
				{BlockNumber: 2, BlockHash: "0xhash2", Hash: "10", TransactionIndex: 10},
			},
		}
	)
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantInsertErr bool
		wantErr       bool
		result        bool
	}{
		{
			name:   "Test_VerifyBlock_ProperlyInserted_1",
			fields: f,
			args: args{
				ctx:               ctx,
				blockHeight:       1,
				verifier:          verifier,
				needToInsertBlock: block1,
				compareBlock:      block1,
			},
			wantInsertErr: false,
			wantErr:       false,
			result:        false,
		},
		{
			name:   "Test_VerifyBlock_ProperlyInserted_2",
			fields: f,
			args: args{
				ctx:               ctx,
				blockHeight:       2,
				verifier:          verifier,
				needToInsertBlock: block2,
				compareBlock:      block2,
			},
			wantInsertErr: false,
			wantErr:       false,
			result:        false,
		},
		{
			name:   "Test_VerifyBlock_ImproperlyInserted_1",
			fields: f,
			args: args{
				ctx:               ctx,
				blockHeight:       1,
				verifier:          verifier,
				needToInsertBlock: corruptedBlock1,
				compareBlock:      block1,
			},
			wantInsertErr: false,
			wantErr:       false,
			result:        true,
		},
		{
			name:   "Test_VerifyBlock_ImproperlyInserted_2",
			fields: f,
			args: args{
				ctx:               ctx,
				blockHeight:       2,
				verifier:          verifier,
				needToInsertBlock: corruptedBlock2,
				compareBlock:      block2,
			},
			wantInsertErr: false,
			wantErr:       false,
			result:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &infoServer{
				dbClient:          tt.fields.dbClient,
				cacheClient:       tt.fields.cacheClient,
				kaiClient:         tt.fields.kaiClient,
				metrics:           tt.fields.metrics,
				HttpRequestSecret: tt.fields.HttpRequestSecret,
				logger:            tt.fields.logger,
			}
			err := s.ImportBlock(tt.args.ctx, tt.args.needToInsertBlock, false)
			if (err != nil) != tt.wantInsertErr {
				t.Errorf("ImportBlock() error = %v, wantErr %v", err, tt.wantInsertErr)
			}
			got, err := s.VerifyBlock(tt.args.ctx, tt.args.blockHeight, tt.args.compareBlock, tt.args.verifier)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.result) {
				t.Errorf("VerifyBlock() result = %v, wantResult %v", got, tt.result)
			}
			if err := s.dbClient.DeleteBlockByHeight(tt.args.ctx, tt.args.blockHeight); err != nil {
				t.Fatalf("cannot flush database after testcase")
			}
		})
	}
}
