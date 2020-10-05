package kardia

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/blendle/zapdriver"
	"github.com/stretchr/testify/assert"

	"github.com/kardiachain/explorer-backend/metrics"
	"github.com/kardiachain/explorer-backend/types"
	"github.com/kardiachain/go-kardiamain/lib/common"
)

const (
	stressTestAmount uint64 = 0
	// minBlockNumber uint64 = 1<<bits.UintSize - 1

)

func SetupKAIClient() (*Client, context.Context, *metrics.Provider, error) {
	ctx, _ := context.WithCancel(context.Background())
	cfg := zapdriver.NewProductionConfig()
	logger, err := cfg.Build()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to create logger: %v", err)
	}
	// defer logger.Sync()
	client, err := NewKaiClient("http://10.10.0.251:8551", logger)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to create new KaiClient: %v", err)
	}
	return client, ctx, metrics.New(), nil
}

func TestLatestBlockNumber(t *testing.T) {
	client, ctx, metrics, err := SetupKAIClient()
	assert.Nil(t, err)

	startTime := time.Now()
	num, err := client.LatestBlockNumber(ctx)
	metrics.RecordProcessingTime(time.Since(startTime))

	assert.Nil(t, err)
	t.Log("Latest block number: ", num, " Elasped time: ", metrics.GetProcessingTime())
	assert.IsTypef(t, uint64(0), num, "Block number must be an uint64")
	assert.NotNil(t, num)

	for i := uint64(0); i < stressTestAmount; i++ {
		startTime = time.Now()
		num, err = client.LatestBlockNumber(ctx)
		metrics.RecordProcessingTime(time.Since(startTime))

		assert.Nil(t, err)
		assert.IsTypef(t, uint64(0), num, "Block number must be an uint64")
		assert.NotNil(t, num)
	}
	t.Log("Latest block number: ", num, " Last operation executed time: ", time.Since(startTime))
	t.Log("Stress test average time: ", metrics.GetProcessingTime())
}

func TestBlockByHash(t *testing.T) {
	client, ctx, metrics, err := SetupKAIClient()
	assert.Nil(t, err)
	emptyBlock := types.Block{}

	hash := "0xd533fc1f9d6836394c6fd43fa3c6d86524fa5d45795a021ed7cccbe7164f7200"
	startTime := time.Now()
	b, err := client.BlockByHash(ctx, common.HexToHash(hash))
	metrics.RecordProcessingTime(time.Since(startTime))

	assert.Nil(t, err)
	t.Log("\nHash: ", hash, "\nBlock: ", b, "\nElasped time: ", metrics.GetProcessingTime())
	assert.IsTypef(t, &emptyBlock, b, "Block must be a types.Block object")
	assert.NotNil(t, b)

	for i := uint64(0); i < stressTestAmount; i++ {
		startTime = time.Now()
		b, err = client.BlockByHash(ctx, common.HexToHash(hash))
		metrics.RecordProcessingTime(time.Since(startTime))

		assert.Nil(t, err)
		assert.IsTypef(t, &emptyBlock, b, "Block must be a types.Block object")
		assert.NotNil(t, b)
	}
	t.Log("Block: ", b, "\nLast operation executed time: ", time.Since(startTime))
	t.Log("Stress test average time: ", metrics.GetProcessingTime())
}

func TestBlockByNumber(t *testing.T) {
	client, ctx, metrics, err := SetupKAIClient()
	assert.Nil(t, err)
	emptyBlock := types.Block{}

	num, err := client.LatestBlockNumber(ctx)
	startTime := time.Now()
	b, err := client.BlockByNumber(ctx, num)
	metrics.RecordProcessingTime(time.Since(startTime))

	assert.Nil(t, err)
	t.Log("\nBlock number: ", num, "\nBlock: ", b, "\nElasped time: ", metrics.GetProcessingTime())
	assert.IsTypef(t, &emptyBlock, b, "Block must be a types.Block object")
	assert.NotNil(t, b)

	for i := uint64(num); i > uint64(num)-stressTestAmount; i-- {
		startTime = time.Now()
		b, err = client.BlockByNumber(ctx, i)
		metrics.RecordProcessingTime(time.Since(startTime))

		assert.Nil(t, err)
		assert.IsTypef(t, &emptyBlock, b, "Block must be a types.Block object")
		assert.NotNil(t, b)
	}
	t.Log("Block: ", b, "\nLast operation executed time: ", time.Since(startTime))
	t.Log("Stress test average time: ", metrics.GetProcessingTime())
}

func TestBlockHeaderByNumber(t *testing.T) {
	client, ctx, metrics, err := SetupKAIClient()
	assert.Nil(t, err)
	emptyBlockHeader := types.Header{}

	num, err := client.LatestBlockNumber(ctx)
	startTime := time.Now()
	h, err := client.BlockHeaderByNumber(ctx, num)
	metrics.RecordProcessingTime(time.Since(startTime))

	assert.Nil(t, err)
	t.Log("\nBlock number: ", num, "\nBlock header: ", h, "\nElasped time: ", metrics.GetProcessingTime())
	assert.IsTypef(t, &emptyBlockHeader, h, "Block header must be a types.Header object")
	assert.NotNil(t, h)

	for i := uint64(num); i > uint64(num)-stressTestAmount; i-- {
		startTime = time.Now()
		h, err = client.BlockHeaderByNumber(ctx, i)
		metrics.RecordProcessingTime(time.Since(startTime))

		assert.Nil(t, err)
		assert.IsTypef(t, &emptyBlockHeader, h, "Block header must be a types.Header object")
		assert.NotNil(t, h)
	}
	t.Log("Block header: ", h, "\nLast operation executed time: ", time.Since(startTime))
	t.Log("Stress test average time: ", metrics.GetProcessingTime())
}

func TestBlockHeaderByHash(t *testing.T) {
	client, ctx, metrics, err := SetupKAIClient()
	assert.Nil(t, err)
	emptyBlockHeader := types.Header{}

	hash := "0xd533fc1f9d6836394c6fd43fa3c6d86524fa5d45795a021ed7cccbe7164f7200"
	startTime := time.Now()
	h, err := client.BlockHeaderByHash(ctx, common.HexToHash(hash))
	metrics.RecordProcessingTime(time.Since(startTime))

	assert.Nil(t, err)
	t.Log("\nHash: ", hash, "\nBlock header: ", h, "\nElasped time: ", metrics.GetProcessingTime())
	assert.IsTypef(t, &emptyBlockHeader, h, "Block header must be a types.Header object")
	assert.NotNil(t, h)

	for i := uint64(0); i < stressTestAmount; i++ {
		startTime = time.Now()
		h, err := client.BlockHeaderByHash(ctx, common.HexToHash(hash))
		metrics.RecordProcessingTime(time.Since(startTime))

		assert.Nil(t, err)
		assert.IsTypef(t, &emptyBlockHeader, h, "Block header must be a types.Header object")
		assert.NotNil(t, h)
	}
	t.Log("Block header: ", h, "\nLast operation executed time: ", time.Since(startTime))
	t.Log("Stress test average time: ", metrics.GetProcessingTime())
}

func TestBalanceAt(t *testing.T) {
	client, ctx, metrics, err := SetupKAIClient()
	assert.Nil(t, err)

	// num, err := client.LatestBlockNumber(ctx)
	addr := "0xfF3dac4f04dDbD24dE5D6039F90596F0a8bb08fd"
	startTime := time.Now()
	b, err := client.BalanceAt(ctx, common.HexToAddress(addr), common.NewZeroHash(), 0)
	metrics.RecordProcessingTime(time.Since(startTime))

	assert.Nil(t, err)
	t.Log("Address: ", addr, " Balance: ", b, " Elasped time: ", metrics.GetProcessingTime())
	assert.IsTypef(t, "", b, "Balance must be a string")
	assert.NotEqualValuesf(t, b, "-1", "Balance must be larger than -1")

	for i := uint64(0); i < stressTestAmount; i++ {
		startTime = time.Now()
		b, err = client.BalanceAt(ctx, common.HexToAddress(addr), common.NewZeroHash(), 0)
		metrics.RecordProcessingTime(time.Since(startTime))

		assert.Nil(t, err)
		assert.IsTypef(t, "", b, "Balance must be a string")
		assert.NotEqualValuesf(t, b, "-1", "Balance must be larger than -1")
	}
	t.Log("Address: ", addr, " Balance: ", b, " Last operation executed time: ", time.Since(startTime))
	t.Log("Stress test average time: ", metrics.GetProcessingTime())
}

func TestNonceAt(t *testing.T) {
	client, ctx, metrics, err := SetupKAIClient()
	assert.Nil(t, err)

	// num, err := client.LatestBlockNumber(ctx)
	addr := "0xfF3dac4f04dDbD24dE5D6039F90596F0a8bb08fd"
	startTime := time.Now()
	n, err := client.NonceAt(ctx, common.HexToAddress(addr))
	metrics.RecordProcessingTime(time.Since(startTime))

	assert.Nil(t, err)
	t.Log("Address: ", addr, " Nonce: ", n, " Elasped time: ", metrics.GetProcessingTime())
	assert.IsTypef(t, uint64(0), n, "Nonce must be an uint64")
	assert.NotNil(t, n)

	for i := uint64(0); i < stressTestAmount; i++ {
		startTime = time.Now()
		n, err = client.NonceAt(ctx, common.HexToAddress(addr))
		metrics.RecordProcessingTime(time.Since(startTime))

		assert.Nil(t, err)
		assert.IsTypef(t, uint64(0), n, "Nonce must be an uint64")
		assert.NotNil(t, n)
	}
	t.Log("Address: ", addr, " Nonce: ", n, " Last operation executed time: ", time.Since(startTime))
	t.Log("Stress test average time: ", metrics.GetProcessingTime())
}
