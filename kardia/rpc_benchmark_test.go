/*
 *  Copyright 2018 KardiaChain
 *  This file is part of the go-kardia library.
 *
 *  The go-kardia library is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Lesser General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  The go-kardia library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Lesser General Public License for more details.
 *
 *  You should have received a copy of the GNU Lesser General Public License
 *  along with the go-kardia library. If not, see <http://www.gnu.org/licenses/>.
 */
// Package kardia
package kardia

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/blendle/zapdriver"
	"github.com/stretchr/testify/assert"

	"github.com/kardiachain/go-kardiamain/lib/common"
)

func SetupKAIClientForBenchmark() (*Client, context.Context, error) {
	ctx, cancelFn := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range sigCh {
			cancelFn()
		}
	}()
	cfg := zapdriver.NewProductionConfig()
	logger, err := cfg.Build()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create logger: %v", err)
	}
	defer logger.Sync()
	client, err := NewKaiClient("http://10.10.0.251:8551", logger)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create new KaiClient: %v", err)
	}
	return client, ctx, nil
}

func BenchmarkLatestBlockNumber(b *testing.B) {
	client, ctx, err := SetupKAIClientForBenchmark()
	assert.Nil(b, err)

	for i := 0; i < b.N; i++ {
		_, _ = client.LatestBlockNumber(ctx)
	}
}

func BenchmarkBlockByHash(b *testing.B) {
	client, ctx, err := SetupKAIClientForBenchmark()
	assert.Nil(b, err)
	hash := "0x63e5862cd056fc0807beb5d47a39b9eac5900c33673df78c1c216b0a3a3f4100"
	for i := 0; i < b.N; i++ {
		_, _ = client.BlockByHash(ctx, common.HexToHash(hash))
	}
}

func BenchmarkBlockByNumber(b *testing.B) {
	client, ctx, err := SetupKAIClientForBenchmark()
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		_, _ = client.BlockByNumber(ctx, 50)
	}
}

func BenchmarkBlockHeaderByNumber(b *testing.B) {
	client, ctx, err := SetupKAIClientForBenchmark()
	assert.Nil(b, err)

	for i := 0; i < b.N; i++ {
		_, _ = client.BlockHeaderByNumber(ctx, 50)
	}
}

func BenchmarkBlockHeaderByHash(b *testing.B) {
	client, ctx, err := SetupKAIClientForBenchmark()
	assert.Nil(b, err)
	hash := "0x63e5862cd056fc0807beb5d47a39b9eac5900c33673df78c1c216b0a3a3f4100"

	for i := 0; i < b.N; i++ {
		_, _ = client.BlockHeaderByHash(ctx, common.HexToHash(hash))
	}
}

func BenchmarkBalanceAt(b *testing.B) {
	client, ctx, err := SetupKAIClientForBenchmark()
	assert.Nil(b, err)
	addr := "0xfF3dac4f04dDbD24dE5D6039F90596F0a8bb08fd"

	for i := 0; i < b.N; i++ {
		_, _ = client.BalanceAt(ctx, common.HexToAddress(addr), common.NewZeroHash(), 0)
	}
}

func BenchmarkNonceAt(b *testing.B) {
	client, ctx, err := SetupKAIClientForBenchmark()
	assert.Nil(b, err)
	addr := "0xfF3dac4f04dDbD24dE5D6039F90596F0a8bb08fd"

	for i := 0; i < b.N; i++ {
		_, _ = client.NonceAt(ctx, common.HexToAddress(addr))
	}
}

// TODO(trinhdn): continue testing other implemented methods
