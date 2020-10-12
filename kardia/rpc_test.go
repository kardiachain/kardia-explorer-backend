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
	"math/bits"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/blendle/zapdriver"
	"github.com/stretchr/testify/assert"

	"github.com/kardiachain/go-kardiamain/lib/p2p"
	coreTypes "github.com/kardiachain/go-kardiamain/types"

	"github.com/kardiachain/explorer-backend/metrics"
	"github.com/kardiachain/explorer-backend/types"
)

type testSuite struct {
	rpcURL []string

	minBlockNumber uint64

	m *metrics.Provider

	blockHeight uint64
	blockHash   string
	txHash      string
	address     string

	sampleBlock       *types.Block
	sampleBlockHeader *types.Header
	sampleTx          *types.Transaction
	sampleTxReceipt   *types.Receipt
	samplePeer        *p2p.PeerInfo
	sampleNodeInfo    *p2p.NodeInfo
	sampleDatadir     string
	sampleValidator   []*types.Validator
}

func setupTestSuite() *testSuite {
	blockHeight := uint64(395)
	blockHash := "0x634662e42bc71d2a7ca767ca19735c8f19694fd7dfbbc70bb28698e0e01be888"
	txHash := "0x02e90c26892a6d230b6964a78446e055b289c5ad53039468ea6a147c0ee31990"
	address := "0xc1fe56E3F58D3244F606306611a5d10c8333f1f6"
	sampleBlock := &types.Block{
		Hash:   blockHash,
		Height: blockHeight,
		Time:   1601908120,
		NumTxs: 0,
		// NumDualEvents:
		GasLimit:   1050000000,
		GasUsed:    0,
		LastBlock:  "0xf9fd47f388c3f41214d55c51fc3d59c8a5e550099a2aa3468d500539d15b0c7a",
		CommitHash: "0x45239add9675da72e0edb5c0ccc80d0f1c758a3cd398fee7f863d69d4863a759",
		DataHash:   "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
		// DualEventsHash:
		ReceiptsRoot:  "",
		LogsBloom:     coreTypes.Bloom{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Validator:     "0xc1fe56E3F58D3244F606306611a5d10c8333f1f6",
		ValidatorHash: "0x6231cec385931237749482972bf28d819fe9527c5ba618cd3620a1ba3be65bbd",
		ConsensusHash: "0x0000000000000000000000000000000000000000000000000000000000000000",
		AppHash:       "0x0d5d8f1a6fdffac4d9c93b1e230da611b65bad7b3c82fb28300247e3a40df76c",
		EvidenceHash:  "0x0000000000000000000000000000000000000000000000000000000000000000",

		Txs:      []*types.Transaction(nil),
		Receipts: []*types.Receipt(nil),
	}
	sampleBlockHeader := &types.Header{
		Hash:   blockHash,
		Height: blockHeight,
		Time:   1601908120,
		NumTxs: 0,
		// NumDualEvents:
		GasLimit:   1050000000,
		GasUsed:    0,
		LastBlock:  "0xf9fd47f388c3f41214d55c51fc3d59c8a5e550099a2aa3468d500539d15b0c7a",
		CommitHash: "0x45239add9675da72e0edb5c0ccc80d0f1c758a3cd398fee7f863d69d4863a759",
		DataHash:   "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
		// DualEventsHash:
		ReceiptsRoot:  "",
		LogsBloom:     coreTypes.Bloom{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Validator:     "0xc1fe56E3F58D3244F606306611a5d10c8333f1f6",
		ValidatorHash: "0x6231cec385931237749482972bf28d819fe9527c5ba618cd3620a1ba3be65bbd",
		ConsensusHash: "0x0000000000000000000000000000000000000000000000000000000000000000",
		AppHash:       "0x0d5d8f1a6fdffac4d9c93b1e230da611b65bad7b3c82fb28300247e3a40df76c",
		EvidenceHash:  "0x0000000000000000000000000000000000000000000000000000000000000000",
	}
	sampleTx := &types.Transaction{
		Hash: txHash,
		To:   "0x2500A193147c8B8FfB4808564a2DC0f476400B86",
		From: address,
		// Status:
		// ContractAddress:
		Value:    "2",
		GasPrice: 1,
		GasFee:   21000,
		// GasLimit:
		BlockNumber: 142,
		Nonce:       1304,
		BlockHash:   "0x484d030a20881754beea5b17485868df2e9cde3fea20adbe9ae48dbc73529605",
		Time:        1601908519,
		InputData:   "0x",
		// Logs:
		TransactionIndex: 1,
		// ReceiptReceived:
	}
	sampleTxReceipt := &types.Receipt{
		BlockHash:         "0x484d030a20881754beea5b17485868df2e9cde3fea20adbe9ae48dbc73529605",
		BlockHeight:       142,
		TransactionHash:   txHash,
		TransactionIndex:  1,
		From:              address,
		To:                "0x2500A193147c8B8FfB4808564a2DC0f476400B86",
		GasUsed:           21000,
		CumulativeGasUsed: 42000,
		ContractAddress:   "0x",
		Logs:              []types.Log{},
		LogsBloom:         coreTypes.Bloom{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		Root:              "0x",
		Status:            1,
	}
	samplePeer := &p2p.PeerInfo{}
	sampleNodeInfo := &p2p.NodeInfo{}
	sampleValidator := []*types.Validator{}
	m := metrics.New()
	return &testSuite{
		rpcURL:            []string{"http://10.10.0.251:8545", "http://10.10.0.251:8546", "http://10.10.0.251:8547", "http://10.10.0.251:8548", "http://10.10.0.251:8549", "http://10.10.0.251:8550", "http://10.10.0.251:8551"},
		minBlockNumber:    1<<bits.UintSize - 1,
		m:                 m,
		blockHeight:       blockHeight,
		blockHash:         blockHash,
		txHash:            txHash,
		address:           address,
		sampleBlock:       sampleBlock,
		sampleBlockHeader: sampleBlockHeader,
		sampleTx:          sampleTx,
		sampleTxReceipt:   sampleTxReceipt,
		samplePeer:        samplePeer,
		sampleNodeInfo:    sampleNodeInfo,
		sampleDatadir:     "/home/.kardia",
		sampleValidator:   sampleValidator,
	}
}

func SetupKAIClient() (ClientInterface, context.Context, *testSuite, error) {
	suite := setupTestSuite()
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
		return nil, nil, suite, fmt.Errorf("Failed to create logger: %v", err)
	}
	defer logger.Sync()
	client, err := NewKaiClient(suite.rpcURL, logger)
	if err != nil {
		return nil, nil, suite, fmt.Errorf("Failed to create new KaiClient: %v", err)
	}
	return client, ctx, suite, nil
}

func TestSanity(t *testing.T) {

}

func TestLatestBlockNumber(t *testing.T) {
	client, ctx, _, err := SetupKAIClient()
	assert.Nil(t, err)

	num, err := client.LatestBlockNumber(ctx)

	assert.Nil(t, err)
	t.Log("Latest block number: ", num)
	assert.IsTypef(t, uint64(0), num, "Block number must be an uint64")
}

func TestBlockByHash(t *testing.T) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(t, err)

	b, err := client.BlockByHash(ctx, testSuite.blockHash)

	assert.Nil(t, err)
	t.Log("Hash: ", testSuite.blockHash, "\nBlock: ", b)
	assert.EqualValuesf(t, testSuite.sampleBlock, b, "Received block must be equal to sampleBlock in testSuite")
}

func TestBlockByNumber(t *testing.T) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(t, err)

	b, err := client.BlockByHeight(ctx, testSuite.blockHeight)

	assert.Nil(t, err)
	t.Log("\nBlock number: ", testSuite.blockHeight, "\nBlock: ", b)
	assert.EqualValuesf(t, testSuite.sampleBlock, b, "Received block must be equal to sampleBlock in testSuite")
}

func TestBlockHeaderByNumber(t *testing.T) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(t, err)

	h, err := client.BlockHeaderByNumber(ctx, testSuite.blockHeight)

	assert.Nil(t, err)
	t.Log("Block number: ", testSuite.blockHeight, "\nBlock header: ", h)
	assert.EqualValuesf(t, testSuite.sampleBlockHeader, h, "Received block header must be equal to sampleBlockHeader in testSuite")
}

func TestBlockHeaderByHash(t *testing.T) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(t, err)

	h, err := client.BlockHeaderByHash(ctx, testSuite.blockHash)

	assert.Nil(t, err)
	t.Log("\nHash: ", testSuite.blockHash, "\nBlock header: ", h)
	assert.EqualValuesf(t, testSuite.sampleBlockHeader, h, "Received block header must be equal to sampleBlockHeader in testSuite")
}

func TestBalanceAt(t *testing.T) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(t, err)

	b, err := client.BalanceAt(ctx, testSuite.address, nil)

	assert.Nil(t, err)
	t.Log("Address: ", testSuite.address, " Balance: ", b)
	assert.IsTypef(t, "", b, "Balance must be a string")
	assert.NotEqualValuesf(t, b, "-1", "Balance must be larger than -1")
}

func TestNonceAt(t *testing.T) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(t, err)

	n, err := client.NonceAt(ctx, testSuite.address)

	assert.Nil(t, err)
	t.Log("Address: ", testSuite.address, " Nonce: ", n)
	assert.IsTypef(t, uint64(0), n, "Nonce must be an uint64")
}

func TestGetTransaction(t *testing.T) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(t, err)

	tx, isPending, err := client.GetTransaction(ctx, testSuite.txHash)

	assert.Nil(t, err)
	assert.EqualValuesf(t, false, isPending, "isPending must be true")
	assert.EqualValuesf(t, testSuite.sampleTx, tx, "Received tx must be equal to sampleTx in testSuite")
}

func TestGetTransactionReceipt(t *testing.T) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(t, err)

	receipt, err := client.GetTransactionReceipt(ctx, testSuite.txHash)

	assert.Nil(t, err)
	assert.EqualValuesf(t, testSuite.sampleTxReceipt, receipt, "Received receipt must be equal to sampleTxReceipt in testSuite")

	for i := 0; i < 1000; i++ {
		startTime := time.Now()
		_, _ = client.GetTransactionReceipt(ctx, testSuite.txHash)
		testSuite.m.RecordProcessingTime(time.Since(startTime))
	}

	t.Log("1000 req: ", testSuite.m.GetProcessingTime())
	testSuite.m.Reset()

	// for i := 0; i < 10000; i++ {
	// 	startTime := time.Now()
	// 	_, _ = client.GetTransactionReceipt(ctx, testSuite.txHash)
	// 	testSuite.m.RecordProcessingTime(time.Since(startTime))
	// }

	// t.Log("10000 req: ", testSuite.m.GetProcessingTime())
	// testSuite.m.Reset()

	// for i := 0; i < 100000; i++ {
	// 	startTime := time.Now()
	// 	_, _ = client.GetTransactionReceipt(ctx, testSuite.txHash)
	// 	testSuite.m.RecordProcessingTime(time.Since(startTime))
	// }

	// t.Log("100000 req: ", testSuite.m.GetProcessingTime())
	// testSuite.m.Reset()
}

func TestPeers(t *testing.T) {
	client, ctx, _, err := SetupKAIClient()
	assert.Nil(t, err)

	peers, err := client.Peers(ctx)

	assert.Nil(t, err)
	assert.IsTypef(t, []*p2p.PeerInfo{}, peers, "peers must be an array of *p2p.PeerInfo")
	// assert.EqualValuesf(t, testSuite.sampleTxReceipt, peers, "Received receipt must be equal to sampleTxReceipt in testSuite")
}

func TestNodeInfo(t *testing.T) {
	client, ctx, _, err := SetupKAIClient()
	assert.Nil(t, err)

	node, err := client.NodeInfo(ctx)

	assert.Nil(t, err)
	assert.IsTypef(t, &p2p.NodeInfo{}, node, "node must be a *p2p.NodeInfo")
	// assert.EqualValuesf(t, testSuite.sampleTxReceipt, node, "Received receipt must be equal to sampleTxReceipt in testSuite")
}

func TestDataDir(t *testing.T) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(t, err)

	dir, err := client.Datadir(ctx)

	assert.Nil(t, err)
	assert.EqualValuesf(t, testSuite.sampleDatadir, dir, "Receive data directory must be equal to sampleDatadir in testSuite")
}

func TestValidators(t *testing.T) {
	client, ctx, testSuite, err := SetupKAIClient()
	assert.Nil(t, err)

	validators := client.Validators(ctx)
	t.Log(validators[0])

	assert.IsTypef(t, testSuite.sampleValidator, validators, "Validators must be a []*Validator")
	// assert.EqualValuesf(t, testSuite.sampleDatadir, dir, "Receive data directory must be equal to sampleDatadir in testSuite")
}

func TestValidator(t *testing.T) {
	client, ctx, _, err := SetupKAIClient()
	assert.Nil(t, err)

	validator := client.Validator(ctx)
	t.Log(validator)

	assert.IsTypef(t, &types.Validator{}, validator, "Validator must be a *Validator")
	// assert.EqualValuesf(t, testSuite.sampleDatadir, dir, "Receive data directory must be equal to sampleDatadir in testSuite")
}

// TODO(trinhdn): continue testing other implemented methods
