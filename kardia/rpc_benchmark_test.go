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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kardiachain/go-kardiamain/lib/common"
)

func BenchmarkLatestBlockNumber(b *testing.B) {
	client, ctx, _, err := SetupKAIClient()
	assert.Nil(b, err)

	for i := 0; i < b.N; i++ {
		_, _ = client.LatestBlockNumber(ctx)
	}
}

func BenchmarkBlockByHash(b *testing.B) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		_, _ = client.BlockByHash(ctx, common.HexToHash(testSuite.blockHash))
	}
}

func BenchmarkBlockByNumber(b *testing.B) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		_, _ = client.BlockByNumber(ctx, testSuite.blockHeight)
	}
}

func BenchmarkBlockHeaderByNumber(b *testing.B) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		_, _ = client.BlockHeaderByNumber(ctx, testSuite.blockHeight)
	}
}

func BenchmarkBlockHeaderByHash(b *testing.B) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		_, _ = client.BlockHeaderByHash(ctx, common.HexToHash(testSuite.blockHash))
	}
}

func BenchmarkBalanceAt(b *testing.B) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		_, _ = client.BalanceAt(ctx, common.HexToAddress(testSuite.address), common.NewZeroHash(), 0)
	}
}

func BenchmarkNonceAt(b *testing.B) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		_, _ = client.NonceAt(ctx, common.HexToAddress(testSuite.address))
	}
}

func BenchmarkGetTransaction(b *testing.B) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		_, _, _ = client.GetTransaction(ctx, common.HexToHash(testSuite.txHash))
	}
}

func BenchmarkGetTransactionReceipt(b *testing.B) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		_, _ = client.GetTransactionReceipt(ctx, common.HexToHash(testSuite.txHash))
	}
}

func BenchmarkPeers(b *testing.B) {
	client, ctx, _, err := SetupKAIClient()
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		_, _ = client.Peers(ctx)
	}
}

func BenchmarkNodeInfo(b *testing.B) {
	client, ctx, _, err := SetupKAIClient()
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		_, _ = client.NodeInfo(ctx)
	}

}

func BenchmarkDataDir(b *testing.B) {
	client, ctx, _, err := SetupKAIClient()
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		_, _ = client.Datadir(ctx)
	}
}

func BenchmarkValidators(b *testing.B) {
	client, ctx, _, err := SetupKAIClient()
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		_ = client.Validators(ctx)
	}
}

func BenchmarkValidator(b *testing.B) {
	client, ctx, _, err := SetupKAIClient()
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		_ = client.Validator(ctx)
	}
}

// TODO(trinhdn): continue testing other implemented methods
