// Package server
package server

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"math/big"
	"net"
	"net/http"
	"time"

	"github.com/kardiachain/explorer-backend/cfg"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/kardia"
	"github.com/kardiachain/explorer-backend/metrics"
	"github.com/kardiachain/explorer-backend/server/cache"
	"github.com/kardiachain/explorer-backend/server/db"
	"github.com/kardiachain/explorer-backend/types"
	"github.com/kardiachain/explorer-backend/utils"
	"github.com/kardiachain/go-kardia/lib/common"
)

type InfoServer interface {
	// API
	LatestBlockHeight(ctx context.Context) (uint64, error)
	UpdateCurrentStats(ctx context.Context) error

	// DB
	//LatestInsertBlockHeight(ctx context.Context) (uint64, error)

	// Share interface
	BlockByHeight(ctx context.Context, blockHeight uint64) (*types.Block, error)
	BlockByHash(ctx context.Context, blockHash string) (*types.Block, error)
	BlockByHeightFromRPC(ctx context.Context, blockHeight uint64) (*types.Block, error)

	ImportBlock(ctx context.Context, block *types.Block, writeToCache bool) error
	DeleteLatestBlock(ctx context.Context) (uint64, error)
	DeleteBlockByHeight(ctx context.Context, height uint64) error
	UpsertBlock(ctx context.Context, block *types.Block) error

	InsertErrorBlocks(ctx context.Context, start uint64, end uint64) error
	PopErrorBlockHeight(ctx context.Context) (uint64, error)
	InsertPersistentErrorBlocks(ctx context.Context, blockHeight uint64) error
	InsertUnverifiedBlocks(ctx context.Context, height uint64) error
	PopUnverifiedBlockHeight(ctx context.Context) (uint64, error)

	VerifyBlock(ctx context.Context, blockHeight uint64, networkBlock *types.Block) (bool, error)
}

// infoServer handle how data was retrieved, stored without interact with other network excluded dbClient
type infoServer struct {
	dbClient    db.Client
	cacheClient cache.Client
	kaiClient   kardia.ClientInterface

	metrics *metrics.Provider

	HttpRequestSecret string
	verifyBlockParam  *types.VerifyBlockParam

	logger *zap.Logger
}

func (s *infoServer) TokenInfo(ctx context.Context) (*types.TokenInfo, error) {

	type CMQuote struct {
		Price            float64 `json:"price"`
		Volume24h        float64 `json:"volume_24h"`
		PercentChange1h  float64 `json:"percent_change_1h"`
		PercentChange24h float64 `json:"percent_change_24h"`
		PercentChange7d  float64 `json:"percent_change_7d"`
		MarketCap        float64 `json:"market_cap"`
		LastUpdated      string  `json:"last_updated"`
	}
	type CMTokenInfo struct {
		ID                int                `json:"id"`
		Name              string             `json:"name"`
		Symbol            string             `json:"symbol"`
		Slug              string             `json:"slug"`
		NumMarketPairs    int                `json:"num_market_pairs"`
		DateAdded         string             `json:"date_added"`
		Tags              []string           `json:"tags"`
		MaxSupply         int64              `json:"max_supply"`
		CirculatingSupply int64              `json:"circulating_supply"`
		TotalSupply       int64              `json:"total_supply"`
		IsActive          int                `json:"is_active"`
		Platform          interface{}        `json:"platform"`
		CmcRank           int                `json:"cmc_rank"`
		IsFiat            int                `json:"is_fiat"`
		LastUpdated       string             `json:"last_updated"`
		Quote             map[string]CMQuote `json:"quote"`
	}

	type CMResponseStatus struct {
		Timestamp    string `json:"timestamp"`
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
		Elapsed      int    `json:"elapsed"`
		CreditCount  int    `json:"credit_count"`
		Notice       string `json:"notice"`
	}

	type CMResponse struct {
		Status CMResponseStatus       `json:"status"`
		Data   map[string]CMTokenInfo `json:"data"`
	}

	coinMarketUrl := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?id=5453"
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}

	req, _ := http.NewRequest(http.MethodGet, coinMarketUrl, nil)
	req.Header.Set("X-CMC_PRO_API_KEY", "a9aaf65c-1d6f-4daf-8e2e-df30bd405b66")

	response, _ := netClient.Do(req)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var cmResponse CMResponse

	if err := json.Unmarshal(body, &cmResponse); err != nil {
		return nil, err
	}

	if cmResponse.Status.ErrorCode != 0 {
		return nil, errors.New("api failed")
	}
	cmData := cmResponse.Data["5453"]
	cmQuote := cmData.Quote["USD"]

	// Cast to internal
	tokenInfo := &types.TokenInfo{
		Name:              cmData.Name,
		Symbol:            cmData.Symbol,
		Decimal:           18,
		TotalSupply:       cmData.TotalSupply,
		CirculatingSupply: cmData.CirculatingSupply,
		Price:             cmQuote.Price,
		Volume24h:         cmQuote.Volume24h,
		Change1h:          cmQuote.PercentChange1h,
		Change24h:         cmQuote.PercentChange24h,
		Change7d:          cmQuote.PercentChange7d,
		MarketCap:         cmQuote.MarketCap,
	}
	if err := s.cacheClient.UpdateTokenInfo(ctx, tokenInfo); err != nil {
		return nil, err
	}

	return tokenInfo, nil
}

func (s *infoServer) UpdateCurrentStats(ctx context.Context) error {
	totalAddr, totalContractAddr, err := s.dbClient.GetTotalAddresses(ctx)
	if err != nil {
		return err
	}
	txsCount, err := s.dbClient.TxsCount(ctx)
	if err != nil {
		return err
	}
	_, err = s.cacheClient.UpdateTotalTxs(ctx, txsCount)
	if err != nil {
		return err
	}
	err = s.cacheClient.UpdateTotalHolders(ctx, totalAddr, totalContractAddr)
	if err != nil {
		return err
	}
	vals, err := s.kaiClient.Validators(ctx)
	if err != nil {
		return err
	}
	err = s.cacheClient.UpdateValidators(ctx, vals)
	if err != nil {
		return err
	}
	return nil
}

// BlockByHash return block by its hash
func (s *infoServer) BlockByHash(ctx context.Context, hash string) (*types.Block, error) {
	lgr := s.logger.With(zap.String("Hash", hash))
	cacheBlock, err := s.cacheClient.BlockByHash(ctx, hash)
	if err == nil {
		return cacheBlock, nil
	}
	lgr.Debug("cannot find block in cache", zap.Error(err))

	dbBlock, err := s.dbClient.BlockByHash(ctx, hash)
	if err == nil {
		return dbBlock, nil
	}
	// Something wrong or we stay behind the network
	lgr.Debug("cannot find block in db", zap.Error(err))
	return s.kaiClient.BlockByHash(ctx, hash)
}

// BlockByHeight return a block by height number
// If our network got stuck atm, then make request to mainnet
func (s *infoServer) BlockByHeight(ctx context.Context, blockHeight uint64) (*types.Block, error) {
	lgr := s.logger.With(zap.Uint64("Height", blockHeight))
	cacheBlock, err := s.cacheClient.BlockByHeight(ctx, blockHeight)
	if err == nil {
		return cacheBlock, nil
	}
	lgr.Debug("cannot find block in cache")

	dbBlock, err := s.dbClient.BlockByHeight(ctx, blockHeight)
	if err == nil {
		return dbBlock, nil
	}
	// Something wrong or we stay behind the network
	lgr.Debug("cannot find block in db")

	return s.kaiClient.BlockByHeight(ctx, blockHeight)
}

// BlockByHeightFromRPC get block from RPC based on block number
func (s *infoServer) BlockByHeightFromRPC(ctx context.Context, blockHeight uint64) (*types.Block, error) {
	return s.kaiClient.BlockByHeight(ctx, blockHeight)
}

// ImportBlock handle workflow of import block into system
func (s *infoServer) ImportBlock(ctx context.Context, block *types.Block, writeToCache bool) error {
	s.logger.Info("Importing block:", zap.Uint64("Height", block.Height),
		zap.Int("Txs length", len(block.Txs)), zap.Int("Receipts length", len(block.Receipts)))
	if isExist, err := s.dbClient.IsBlockExist(ctx, block.Height); err != nil || isExist {
		return types.ErrRecordExist
	}

	if writeToCache {
		if err := s.cacheClient.InsertBlock(ctx, block); err != nil {
			s.logger.Debug("cannot import block to cache", zap.Error(err))
		}
	}

	// merge receipts into corresponding transactions
	// because getBlockByHash/Height API returns 2 array contains txs and receipts separately
	block.Txs = s.mergeAdditionalInfoToTxs(block.Txs, block.Receipts)

	// Start import block
	// consider new routine here
	// todo: add metrics
	// todo @longnd: Use redis or leveldb as mem-write buffer for N blocks
	startTime := time.Now()
	if err := s.dbClient.InsertBlock(ctx, block); err != nil {
		s.logger.Debug("cannot import block to db", zap.Error(err))
		return err
	}
	endTime := time.Since(startTime)
	s.metrics.RecordInsertBlockTime(endTime)
	s.logger.Debug("Total time for import block", zap.Duration("TimeConsumed", endTime), zap.String("Avg", s.metrics.GetInsertBlockTime()))

	if writeToCache {
		if err := s.cacheClient.InsertTxsOfBlock(ctx, block); err != nil {
			s.logger.Debug("cannot import txs to cache", zap.Error(err))
			return err
		}
	}

	startTime = time.Now()
	if err := s.dbClient.InsertTxs(ctx, block.Txs); err != nil {
		return err
	}
	endTime = time.Since(startTime)
	s.metrics.RecordInsertTxsTime(endTime)
	s.logger.Debug("Total time for import tx", zap.Duration("TimeConsumed", endTime), zap.String("Avg", s.metrics.GetInsertTxsTime()))

	// update active addresses
	startTime = time.Now()
	addrsList := filterAddrSet(block.Txs)
	addrsList = s.getAddressBalances(ctx, addrsList)
	if err := s.dbClient.UpdateAddresses(ctx, addrsList); err != nil {
		return err
	}
	endTime = time.Since(startTime)
	s.metrics.RecordInsertActiveAddressTime(endTime)
	s.logger.Debug("Total time for update addresses", zap.Duration("TimeConsumed", endTime), zap.String("Avg", s.metrics.GetInsertActiveAddressTime()))
	startTime = time.Now()
	totalAddr, totalContractAddr, err := s.dbClient.GetTotalAddresses(ctx)
	if err != nil {
		return err
	}
	err = s.cacheClient.UpdateTotalHolders(ctx, totalAddr, totalContractAddr)
	if err != nil {
		return err
	}
	s.logger.Debug("Total time for getting active addresses", zap.Duration("TimeConsumed", time.Since(startTime)))

	if _, err := s.cacheClient.UpdateTotalTxs(ctx, block.NumTxs); err != nil {
		return err
	}
	return nil
}

func (s *infoServer) DeleteLatestBlock(ctx context.Context) (uint64, error) {
	height, err := s.dbClient.DeleteLatestBlock(ctx)
	if err != nil {
		s.logger.Warn("cannot remove old latest block", zap.Error(err))
		return 0, err
	}
	return height, nil
}

func (s *infoServer) DeleteBlockByHeight(ctx context.Context, height uint64) error {
	err := s.dbClient.DeleteBlockByHeight(ctx, height)
	if err != nil {
		s.logger.Warn("cannot remove block in database by height", zap.Error(err))
		return err
	}
	return nil
}

func (s *infoServer) UpsertBlock(ctx context.Context, block *types.Block) error {
	s.logger.Info("Upserting block:", zap.Uint64("Height", block.Height), zap.Int("Txs length", len(block.Txs)), zap.Int("Receipts length", len(block.Receipts)))
	if err := s.dbClient.DeleteBlockByHeight(ctx, block.Height); err != nil {
		return err
	}
	return s.ImportBlock(ctx, block, false)
}

func (s *infoServer) InsertErrorBlocks(ctx context.Context, start uint64, end uint64) error {
	err := s.cacheClient.InsertErrorBlocks(ctx, start, end)
	if err != nil {
		s.logger.Warn("Cannot insert error block into retry list", zap.Uint64("Start", start), zap.Uint64("End", end))
		return err
	}
	return nil
}

func (s *infoServer) PopErrorBlockHeight(ctx context.Context) (uint64, error) {
	height, err := s.cacheClient.PopErrorBlockHeight(ctx)
	if err != nil {
		return 0, err
	}
	return height, nil
}

func (s *infoServer) InsertPersistentErrorBlocks(ctx context.Context, blockHeight uint64) error {
	err := s.cacheClient.InsertPersistentErrorBlocks(ctx, blockHeight)
	if err != nil {
		s.logger.Warn("Cannot insert persistent error block into list", zap.Uint64("blockHeight", blockHeight))
		return err
	}
	return nil
}

func (s *infoServer) InsertUnverifiedBlocks(ctx context.Context, height uint64) error {
	err := s.cacheClient.InsertUnverifiedBlocks(ctx, height)
	if err != nil {
		return err
	}
	return nil
}

func (s *infoServer) PopUnverifiedBlockHeight(ctx context.Context) (uint64, error) {
	height, err := s.cacheClient.PopUnverifiedBlockHeight(ctx)
	if err != nil {
		return 0, err
	}
	return height, nil
}

func (s *infoServer) ImportReceipts(ctx context.Context, block *types.Block) error {
	var listTxByFromAddress []*types.TransactionByAddress
	var listTxByToAddress []*types.TransactionByAddress
	jobs := make(chan types.Transaction, block.NumTxs)
	type response struct {
		err         error
		txByFromAdd *types.TransactionByAddress
		txByToAdd   *types.TransactionByAddress
	}
	results := make(chan response, block.NumTxs)
	addresses := make(map[string]*types.Address)

	//todo @longnd: Move this workers to config or dynamic settings
	for w := 0; w <= 10; w++ {
		go func(jobs <-chan types.Transaction, results chan<- response) {
			for tx := range jobs {
				//s.logger.Debug("Start worker", zap.Any("TX", tx))
				receipt, err := s.kaiClient.GetTransactionReceipt(ctx, tx.Hash)
				if err != nil {
					s.logger.Warn("get receipt err", zap.String("tx hash", tx.Hash), zap.Error(err))
					//todo: consider how we handle this err, just skip it now
					results <- response{
						err: err,
					}
					continue
				}
				toAddress := tx.To
				if tx.To == "" {
					if !utils.IsNilAddress(receipt.ContractAddress) {
						tx.ContractAddress = receipt.ContractAddress
					}
					tx.Status = receipt.Status
					toAddress = tx.ContractAddress
				}

				address, err := s.dbClient.AddressByHash(ctx, toAddress)
				if err != nil {
					//todo: consider how we handle this err, just skip it now
					s.logger.Warn("cannot get address by hash")
					results <- response{
						err: err,
					}
					continue
				}

				if address == nil || address.IsContract {
					for _, l := range receipt.Logs {
						addresses[l.Address] = nil
					}
					if err := s.dbClient.UpdateAddresses(ctx, addresses); err != nil {
						//todo: consider how we handle this err, just skip it now
						s.logger.Warn("cannot update active address")
						results <- response{
							err: err,
						}
						continue
					}
				}
				var res response
				res.txByFromAdd = &types.TransactionByAddress{
					Address: tx.From,
					TxHash:  tx.Hash,
					Time:    tx.Time,
				}

				if tx.From != toAddress {
					res.txByToAdd = &types.TransactionByAddress{
						Address: toAddress,
						TxHash:  tx.Hash,
						Time:    tx.Time,
					}
				}
				results <- res
			}
		}(jobs, results)
	}

	for _, tx := range block.Txs {
		jobs <- *tx
	}
	close(jobs)
	// todo @longnd: try to remove this loop
	size := int(block.NumTxs)
	for i := 0; i < size; i++ {
		r := <-results
		if r.err != nil {
			continue
		}
		if r.txByFromAdd != nil {
			listTxByFromAddress = append(listTxByFromAddress, r.txByFromAdd)
		}
		if r.txByToAdd != nil {
			listTxByToAddress = append(listTxByToAddress, r.txByToAdd)
		}
	}

	// todo @longnd: Handle insert failed
	if len(listTxByToAddress) > 0 {
		s.logger.Debug("ListTxFromAddress", zap.Int("Size", len(listTxByFromAddress)))
		if err := s.dbClient.InsertListTxByAddress(ctx, listTxByFromAddress); err != nil {
			return err
		}
	}

	// todo @longnd: Handle insert failed
	if len(listTxByToAddress) > 0 {
		s.logger.Debug("ListTxByToAddress", zap.Int("Size", len(listTxByFromAddress)))
		if err := s.dbClient.InsertListTxByAddress(ctx, listTxByToAddress); err != nil {
			return err
		}
	}

	return nil
}

func (s *infoServer) blockVerifier(db, network *types.Block) bool {
	if s.verifyBlockParam.VerifyTxCount {
		if db.NumTxs != network.NumTxs {
			return false
		}
	}
	if s.verifyBlockParam.VerifyBlockHash {
		return true
	}
	return true
}

// VerifyBlock called by verifier. It returns `true` if the block is upserted; otherwise it return `false`
func (s *infoServer) VerifyBlock(ctx context.Context, blockHeight uint64, networkBlock *types.Block) (bool, error) {
	isBlockImported, err := s.dbClient.IsBlockExist(ctx, blockHeight)
	if err != nil || !isBlockImported {
		startTime := time.Now()
		if err := s.dbClient.InsertBlock(ctx, networkBlock); err != nil {
			s.logger.Warn("Cannot import block", zap.Uint64("height", blockHeight))
			return false, err
		}
		endTime := time.Since(startTime)
		if endTime > time.Second {
			s.logger.Warn("Unexpected long import block time, over 1s", zap.Duration("TimeConsumed", endTime))
		}
		return true, nil
	}

	dbBlock, err := s.dbClient.BlockByHeight(ctx, blockHeight)
	if err != nil {
		s.logger.Warn("Cannot get block by height from database", zap.Uint64("height", blockHeight))
		return false, err
	}
	_, total, err := s.dbClient.TxsByBlockHeight(ctx, blockHeight, nil)
	if err != nil {
		s.logger.Warn("Cannot get total transactions in block by height from database", zap.Uint64("height", blockHeight))
		return false, err
	}
	dbBlock.NumTxs = total

	if !s.blockVerifier(dbBlock, networkBlock) {
		s.logger.Warn("Block in database is corrupted, upserting...", zap.Uint64("db numTxs", dbBlock.NumTxs), zap.Uint64("network numTxs", networkBlock.NumTxs), zap.Error(err))
		// Force dbBlock with new information from network block
		startTime := time.Now()
		if err := s.UpsertBlock(ctx, networkBlock); err != nil {
			s.logger.Warn("Cannot upsert block", zap.Uint64("height", blockHeight), zap.Error(err))
			return false, err
		}
		endTime := time.Since(startTime)
		s.metrics.RecordUpsertBlockTime(endTime)
		s.logger.Debug("Upsert block time", zap.Duration("TimeConsumed", endTime), zap.String("Avg", s.metrics.GetUpsertBlockTime()))
		return true, nil
	}
	return false, nil
}

// calculateTPS return TPS per each [10, 20, 50] blocks
func (s *infoServer) calculateTPS(startTime uint64) (uint64, error) {
	return 0, nil
}

// getAddressByHash return *types.Address from mgo.Collection("Address")
func (s *infoServer) getAddressByHash(address string) (*types.Address, error) {
	return nil, nil
}

func (s *infoServer) getTxsByBlockNumber(blockNumber int64, filter *types.Pagination) ([]*types.Transaction, error) {
	return nil, nil
}

func filterAddrSet(txs []*types.Transaction) map[string]*types.Address {
	addrs := make(map[string]*types.Address)
	for _, tx := range txs {
		addrs[tx.From] = &types.Address{
			Address:    tx.From,
			IsContract: false,
		}
		addrs[tx.To] = &types.Address{
			Address:    tx.From,
			IsContract: false,
		}
		addrs[tx.ContractAddress] = &types.Address{
			Address:    tx.From,
			IsContract: true,
		}
	}
	delete(addrs, "")
	delete(addrs, "0x")
	return addrs
}

// TODO(trinhdn): need to use workers instead
func (s *infoServer) getAddressBalances(ctx context.Context, addrs map[string]*types.Address) map[string]*types.Address {
	if addrs == nil || len(addrs) == 0 {
		return map[string]*types.Address{}
	}
	vals, err := s.cacheClient.Validators(ctx)
	if err != nil {
		vals = &types.Validators{
			Validators: []*types.Validator{},
		}
	}
	addressesName := map[string]string{}
	for _, v := range vals.Validators {
		addressesName[v.SmcAddress.String()] = v.Name
	}
	addressesName[cfg.StakingContractAddr] = cfg.StakingContractName

	var (
		code   common.Bytes
		result = map[string]*types.Address{}
	)
	for addr := range addrs {
		addressInfo := &types.Address{
			Address: addr,
			Name:    "",
		}
		addressInfo.BalanceString, err = s.kaiClient.GetBalance(ctx, addr)
		if err != nil {
			addressInfo.BalanceString = "0"
		}
		balance, _ := new(big.Int).SetString(addressInfo.BalanceString, 10)
		addressInfo.BalanceFloat, _ = new(big.Float).SetPrec(100).Quo(new(big.Float).SetInt(balance), new(big.Float).SetInt(cfg.Hydro)).Float64() //converting to KAI from HYDRO
		if !addrs[addr].IsContract {
			code, _ = s.kaiClient.GetCode(ctx, addr)
			if len(code) > 0 { // is contract
				addressInfo.IsContract = true
			}
		}
		if addressesName[addr] != "" {
			addressInfo.Name = addressesName[addr]
		}
		result[addr] = addressInfo
	}
	return result
}

func (s *infoServer) mergeAdditionalInfoToTxs(txs []*types.Transaction, receipts []*types.Receipt) []*types.Transaction {
	if receipts == nil || len(receipts) == 0 {
		return txs
	}
	receiptIndex := 0
	var (
		gasPrice   *big.Int
		gasUsed    *big.Int
		txFeeInOxy *big.Int
	)
	for _, tx := range txs {
		if decoded, err := s.kaiClient.DecodeInputData(tx.To, tx.InputData); err == nil {
			tx.DecodedInputData = decoded
		}
		if (receiptIndex > len(receipts)-1) || !(receipts[receiptIndex].TransactionHash == tx.Hash) {
			tx.Status = 0
			continue
		}

		tx.Logs = receipts[receiptIndex].Logs
		tx.Root = receipts[receiptIndex].Root
		tx.Status = receipts[receiptIndex].Status
		tx.GasUsed = receipts[receiptIndex].GasUsed
		tx.ContractAddress = receipts[receiptIndex].ContractAddress
		// update txFee
		gasPrice = new(big.Int).SetUint64(tx.GasPrice)
		gasUsed = new(big.Int).SetUint64(tx.GasUsed)
		txFeeInOxy = new(big.Int).Mul(gasPrice, gasUsed)
		tx.TxFee = new(big.Int).Mul(txFeeInOxy, big.NewInt(int64(math.Pow10(9)))).String()

		receiptIndex++
	}
	return txs
}
