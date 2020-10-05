// Package db
package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/bxcodec/faker/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

// MgoImportBlock seed size * 1000000 records into Blocks collection before run
// todo: Improve setup time for benchmark (should we ?)
func MgoImportBlock(size int, b *testing.B) {
	size = size * 1000000
	host := "127.0.0.1:27017"
	dbName := "explorer_benchmark"
	logger, err := zap.NewDevelopment()
	if err != nil {
		return
	}
	db := &MongoDB{
		logger:  logger,
		wrapper: &KaiMgo{},
	}
	mgoURI := fmt.Sprintf("mongodb://%s", host)
	mgoClient, err := mongo.NewClient(options.Client().ApplyURI(mgoURI), options.Client().SetMinPoolSize(32), options.Client().SetMaxPoolSize(64))
	if err != nil {
		return
	}

	if err := mgoClient.Connect(context.Background()); err != nil {
		return
	}
	db.wrapper.Database(mgoClient.Database(dbName))

	for i := 0; i < size; i++ {
		// seeding record into database
		block := &types.Block{}
		_ = faker.FakeData(&block)
		_ = db.importBlock(context.Background(), block)
	}

	block := &types.Block{}
	_ = faker.FakeData(&block)
	for i := 0; i < b.N; i++ {
		_ = db.importBlock(context.Background(), block)
	}

	// Drop
	db.dropCollection("Blocks")
}

func BenchmarkMgo_ImportBlock10(b *testing.B) { MgoImportBlock(10, b) }

func BenchmarkMgo_ImportBlock100(b *testing.B) { MgoImportBlock(100, b) }

func BenchmarkMgo_ImportBlock1000(b *testing.B) { MgoImportBlock(1000, b) }

func BenchmarkMgo_ImportBlock10000(b *testing.B) { MgoImportBlock(10000, b) }

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