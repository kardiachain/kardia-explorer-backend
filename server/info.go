// Package server
package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/kardiachain/go-kardia/lib/abi"
	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/db"
	"github.com/kardiachain/kardia-explorer-backend/kardia"
	"github.com/kardiachain/kardia-explorer-backend/metrics"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/kardiachain/kardia-explorer-backend/utils"
)

type InfoServer interface {
	// API
	LatestBlockHeight(ctx context.Context) (uint64, error)
	GetCurrentStats(ctx context.Context) error
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
		Name:                     cmData.Name,
		Symbol:                   cmData.Symbol,
		Decimal:                  18,
		TotalSupply:              cmData.TotalSupply,
		ERC20CirculatingSupply:   cmData.CirculatingSupply,
		MainnetCirculatingSupply: 0,
		Price:                    cmQuote.Price,
		Volume24h:                cmQuote.Volume24h,
		Change1h:                 cmQuote.PercentChange1h,
		Change24h:                cmQuote.PercentChange24h,
		Change7d:                 cmQuote.PercentChange7d,
		MarketCap:                cmQuote.MarketCap,
	}
	if err := s.cacheClient.UpdateTokenInfo(ctx, tokenInfo); err != nil {
		return nil, err
	}

	return tokenInfo, nil
}

func (s *infoServer) GetCurrentStats(ctx context.Context) uint64 {
	stats := s.dbClient.Stats(ctx)
	s.logger.Info("Current stats of network", zap.Uint64("UpdatedAtBlock", stats.UpdatedAtBlock),
		zap.Uint64("TotalTransactions", stats.TotalTransactions), zap.Uint64("TotalAddresses", stats.TotalAddresses),
		zap.Uint64("TotalContracts", stats.TotalContracts))
	_ = s.cacheClient.SetTotalTxs(ctx, stats.TotalTransactions)
	_ = s.cacheClient.UpdateTotalHolders(ctx, stats.TotalAddresses, stats.TotalContracts)

	cfg.GenesisAddresses = append(cfg.GenesisAddresses, &types.Address{
		Address: cfg.TreasuryContractAddr,
		Name:    cfg.TreasuryContractName,
	})
	cfg.GenesisAddresses = append(cfg.GenesisAddresses, &types.Address{
		Address: cfg.StakingContractAddr,
		Name:    cfg.StakingContractName,
	})
	cfg.GenesisAddresses = append(cfg.GenesisAddresses, &types.Address{
		Address: cfg.KardiaDeployerAddr,
		Name:    cfg.KardiaDeployerName,
	})
	cfg.GenesisAddresses = append(cfg.GenesisAddresses, &types.Address{
		Address: cfg.ParamsContractAddr,
		Name:    cfg.ParamsContractName,
	})
	vals, _ := s.kaiClient.Validators(ctx)
	//todo: longnd - Temp remove
	//_ = s.cacheClient.UpdateValidators(ctx, vals)
	_ = s.dbClient.ClearValidators(ctx)
	_ = s.dbClient.UpsertValidators(ctx, vals)
	for _, val := range vals {
		cfg.GenesisAddresses = append(cfg.GenesisAddresses, &types.Address{
			Address: val.SmcAddress.Hex(),
			Name:    val.Name,
		})
	}
	for i, addr := range cfg.GenesisAddresses {
		balance, _ := s.kaiClient.GetBalance(ctx, addr.Address)
		balanceInBigInt, _ := new(big.Int).SetString(balance, 10)
		balanceFloat, _ := new(big.Float).SetPrec(100).Quo(new(big.Float).SetInt(balanceInBigInt), new(big.Float).SetInt(cfg.Hydro)).Float64() //converting to KAI from HYDRO

		cfg.GenesisAddresses[i].BalanceFloat = balanceFloat
		cfg.GenesisAddresses[i].BalanceString = balance
		code, _ := s.kaiClient.GetCode(ctx, addr.Address)
		if len(code) > 0 {
			cfg.GenesisAddresses[i].IsContract = true
		}

		// write this address to db
		_ = s.dbClient.InsertAddress(ctx, cfg.GenesisAddresses[i])
	}
	return stats.UpdatedAtBlock
}

func (s *infoServer) UpdateCurrentStats(ctx context.Context) error {
	totalAddrs, totalContracts := s.cacheClient.TotalHolders(ctx)
	stats := &types.Stats{
		UpdatedAt:         time.Now(),
		UpdatedAtBlock:    s.cacheClient.LatestBlockHeight(ctx),
		TotalTransactions: s.cacheClient.TotalTxs(ctx),
		TotalAddresses:    totalAddrs,
		TotalContracts:    totalContracts,
	}
	return s.dbClient.UpdateStats(ctx, stats)
}

// BlockByHash return block by its hash
func (s *infoServer) BlockByHash(ctx context.Context, hash string) (*types.Block, error) {
	lgr := s.logger.With(zap.String("Hash", hash))
	cacheBlock, err := s.cacheClient.BlockByHash(ctx, hash)
	if err == nil {
		return cacheBlock, nil
	}

	dbBlock, err := s.dbClient.BlockByHash(ctx, hash)
	if err == nil {
		return dbBlock, nil
	}
	// Something wrong or we stay behind the network
	lgr.Warn("cannot find block in db", zap.Error(err))
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

	dbBlock, err := s.dbClient.BlockByHeight(ctx, blockHeight)
	if err == nil {
		return dbBlock, nil
	}
	// Something wrong or we stay behind the network
	lgr.Warn("cannot find block by height in db", zap.Uint64("Height", blockHeight))

	return s.kaiClient.BlockByHeight(ctx, blockHeight)
}

// BlockByHeightFromRPC get block from RPC based on block number
func (s *infoServer) BlockByHeightFromRPC(ctx context.Context, blockHeight uint64) (*types.Block, error) {
	return s.kaiClient.BlockByHeight(ctx, blockHeight)
}

// ImportBlock handle workflow of import block into system
func (s *infoServer) ImportBlock(ctx context.Context, block *types.Block, writeToCache bool) error {
	lgr := s.logger.With(zap.String("method", "ImportBlock"))
	lgr.Info("Importing block:", zap.Uint64("Height", block.Height),
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
	block.Txs = s.mergeAdditionalInfoToTxs(ctx, block.Txs, block.Receipts)

	if err := s.filterStakingEvent(ctx, block.Txs); err != nil {
		s.logger.Warn("Filter staking event failed", zap.Error(err))
	}

	if err := s.filterProposalEvent(ctx, block.Txs); err != nil {
		s.logger.Warn("Filter proposal event failed", zap.Error(err))
	}

	// Start import block
	startTime := time.Now()
	if err := s.dbClient.InsertBlock(ctx, block); err != nil {
		return err
	}
	endTime := time.Since(startTime)
	s.metrics.RecordInsertBlockTime(endTime)
	s.logger.Info("Total time for import block", zap.Duration("TimeConsumed", endTime), zap.String("Avg", s.metrics.GetInsertBlockTime()))

	if writeToCache {
		if err := s.cacheClient.InsertTxsOfBlock(ctx, block); err != nil {
			return err
		}
	}

	startTime = time.Now()
	if err := s.dbClient.InsertTxs(ctx, block.Txs); err != nil {
		return err
	}
	endTime = time.Since(startTime)
	s.metrics.RecordInsertTxsTime(endTime)
	s.logger.Info("Total time for import tx", zap.Duration("TimeConsumed", endTime), zap.String("Avg", s.metrics.GetInsertTxsTime()))

	// update active addresses
	startTime = time.Now()
	addrsMap := filterAddrSet(block.Txs)
	addrsList := s.getAddressBalances(ctx, addrsMap)
	if err := s.dbClient.UpdateAddresses(ctx, addrsList); err != nil {
		return err
	}
	endTime = time.Since(startTime)
	s.metrics.RecordInsertActiveAddressTime(endTime)
	s.logger.Info("Total time for update addresses", zap.Duration("TimeConsumed", endTime), zap.String("Avg", s.metrics.GetInsertActiveAddressTime()))
	startTime = time.Now()
	totalAddr, totalContractAddr, err := s.dbClient.GetTotalAddresses(ctx)
	if err != nil {
		return err
	}
	err = s.cacheClient.UpdateTotalHolders(ctx, totalAddr, totalContractAddr)
	if err != nil {
		return err
	}
	s.logger.Info("Total time for getting active addresses", zap.Duration("TimeConsumed", time.Since(startTime)))

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
	var addresses []*types.Address

	for w := 0; w <= 10; w++ {
		go func(jobs <-chan types.Transaction, results chan<- response) {
			for tx := range jobs {
				receipt, err := s.kaiClient.GetTransactionReceipt(ctx, tx.Hash)
				if err != nil {
					s.logger.Warn("get receipt err", zap.String("tx hash", tx.Hash), zap.Error(err))
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
					s.logger.Warn("cannot get address by hash")
					results <- response{
						err: err,
					}
					continue
				}

				if address == nil || address.IsContract {
					if err := s.dbClient.UpdateAddresses(ctx, addresses); err != nil {
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

	if len(listTxByToAddress) > 0 {
		if err := s.dbClient.InsertListTxByAddress(ctx, listTxByFromAddress); err != nil {
			return err
		}
	}

	if len(listTxByToAddress) > 0 {
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
		// Minus network block reward and total txs before re-importing this block
		totalTxs := s.cacheClient.TotalTxs(ctx)
		totalTxs -= networkBlock.NumTxs
		_ = s.cacheClient.SetTotalTxs(ctx, totalTxs)
		// Force replace dbBlock with new information from network block
		startTime := time.Now()
		if err := s.UpsertBlock(ctx, networkBlock); err != nil {
			s.logger.Warn("Cannot upsert block", zap.Uint64("height", blockHeight), zap.Error(err))
			return false, err
		}
		endTime := time.Since(startTime)
		s.metrics.RecordUpsertBlockTime(endTime)
		return true, nil
	}
	return false, nil
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

func (s *infoServer) getAddressBalances(ctx context.Context, addrs map[string]*types.Address) []*types.Address {
	if addrs == nil || len(addrs) == 0 {
		return nil
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
		code     common.Bytes
		addrsMap = map[string]*types.Address{}
	)
	for addr := range addrs {
		addressInfo := &types.Address{
			Address: addr,
			Name:    "",
		}
		// Override when addr existed
		dbAddrInfo, err := s.dbClient.AddressByHash(ctx, addr)
		if err == nil && dbAddrInfo != nil {
			addressInfo = dbAddrInfo
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
		addrsMap[addr] = addressInfo
	}
	var result []*types.Address
	for _, info := range addrsMap {
		result = append(result, info)
	}
	return result
}

func (s *infoServer) mergeAdditionalInfoToTxs(ctx context.Context, txs []*types.Transaction, receipts []*types.Receipt) []*types.Transaction {
	if receipts == nil || len(receipts) == 0 {
		return txs
	}
	receiptIndex := 0
	var (
		gasPrice     *big.Int
		gasUsed      *big.Int
		txFeeInHydro *big.Int
	)
	for _, tx := range txs {
		smcABI, err := s.getSMCAbi(ctx, &types.Log{
			Address: tx.To,
		})
		if err != nil {
			continue
		}
		decoded, err := s.kaiClient.DecodeInputWithABI(tx.To, tx.InputData, smcABI)
		if err == nil {
			tx.DecodedInputData = decoded
		}
		if (receiptIndex > len(receipts)-1) || !(receipts[receiptIndex].TransactionHash == tx.Hash) {
			tx.Status = 0
			continue
		}

		tx.Logs = receipts[receiptIndex].Logs
		if len(tx.Logs) > 0 {
			err := s.storeEvents(ctx, tx.Logs, txs[0].Time)
			if err != nil {
				s.logger.Warn("Cannot store events to db", zap.Error(err))
			}
		}
		tx.Root = receipts[receiptIndex].Root
		tx.Status = receipts[receiptIndex].Status
		tx.GasUsed = receipts[receiptIndex].GasUsed
		tx.ContractAddress = receipts[receiptIndex].ContractAddress
		// update txFee
		gasPrice = new(big.Int).SetUint64(tx.GasPrice)
		gasUsed = new(big.Int).SetUint64(tx.GasUsed)
		txFeeInHydro = new(big.Int).Mul(gasPrice, gasUsed)
		tx.TxFee = txFeeInHydro.String()

		receiptIndex++
	}
	return txs
}

func (s *infoServer) LatestBlockHeight(ctx context.Context) (uint64, error) {
	return s.kaiClient.LatestBlockNumber(ctx)
}

func (s *infoServer) BlockCacheSize(ctx context.Context) (int64, error) {
	return s.cacheClient.ListSize(ctx, cache.KeyBlocks)
}

func (s *infoServer) storeEvents(ctx context.Context, logs []types.Log, blockTime time.Time) error {
	var (
		holdersList     []*types.TokenHolder
		internalTxsList []*types.TokenTransfer
	)
	for i := range logs {
		smcABI, err := s.getSMCAbi(ctx, &logs[i])
		if err != nil {
			continue
		}
		decodedLog, err := s.kaiClient.UnpackLog(&logs[i], smcABI)
		if err != nil {
			decodedLog = &logs[i]
		}
		decodedLog.Time = blockTime
		logs[i] = *decodedLog
		if logs[i].Topics[0] == cfg.KRCTransferTopic {
			holders, err := s.getKRCHolder(ctx, decodedLog)
			if err != nil {
				continue
			}
			holdersList = append(holdersList, holders...)
			iTx := s.getInternalTxs(ctx, decodedLog)
			internalTxsList = append(internalTxsList, iTx)
		}
	}
	// insert holders and internal txs to db
	err := s.dbClient.UpdateHolders(ctx, holdersList)
	if err != nil {
		s.logger.Warn("Cannot update holder info to db", zap.Error(err), zap.Any("holdersList", holdersList))
	}
	err = s.dbClient.UpdateInternalTxs(ctx, internalTxsList)
	if err != nil {
		s.logger.Warn("Cannot update internal txs to db", zap.Error(err), zap.Any("holdersList", holdersList))
	}

	return s.dbClient.InsertEvents(logs)
}

func (s *infoServer) getSMCAbi(ctx context.Context, log *types.Log) (*abi.ABI, error) {
	smcABIStr, err := s.cacheClient.SMCAbi(ctx, log.Address)
	if err != nil {
		smc, _, err := s.dbClient.Contract(ctx, log.Address)
		if err != nil {
			s.logger.Warn("Cannot get smc info from db", zap.Error(err), zap.String("smcAddr", log.Address))
			return nil, err
		}
		if smc.Type != "" {
			err = s.cacheClient.UpdateSMCAbi(ctx, log.Address, cfg.SMCTypePrefix+smc.Type)
			if err != nil {
				s.logger.Warn("Cannot store smc abi to cache", zap.Error(err))
				return nil, err
			}
			smcABIStr, err = s.cacheClient.SMCAbi(ctx, cfg.SMCTypePrefix+smc.Type)
			if err != nil {
				// query then reinsert abi of this SMC type to cache
				smcABIBase64, err := s.dbClient.SMCABIByType(ctx, smc.Type)
				if err != nil {
					s.logger.Warn("Cannot get smc abi by type from DB", zap.Error(err))
					return nil, err
				}
				err = s.cacheClient.UpdateSMCAbi(ctx, cfg.SMCTypePrefix+smc.Type, smcABIBase64)
				if err != nil {
					s.logger.Warn("Cannot store smc abi by type to cache", zap.Error(err))
					return nil, err
				}
				smcABIStr, err = s.cacheClient.SMCAbi(ctx, cfg.SMCTypePrefix+smc.Type)
				if err != nil {
					s.logger.Warn("Cannot get smc abi from cache", zap.Error(err))
					return nil, err
				}
			}
		} else if smc.ABI != "" {
			smcABIStr = smc.ABI
		}
	}
	abiData, err := base64.StdEncoding.DecodeString(smcABIStr)
	if err != nil {
		s.logger.Warn("Cannot decode smc abi", zap.Error(err))
		return nil, err
	}
	jsonABI, err := abi.JSON(bytes.NewReader(abiData))
	if err != nil {
		s.logger.Warn("Cannot convert decoded smc abi to JSON abi", zap.Error(err))
		return nil, err
	}
	// store this abi to cache
	err = s.cacheClient.UpdateSMCAbi(ctx, log.Address, smcABIStr)
	if err != nil {
		s.logger.Warn("Cannot store smc abi to cache", zap.Error(err))
		return nil, err
	}
	return &jsonABI, nil
}

func (s *infoServer) getKRCTokenInfo(ctx context.Context, krcTokenAddr string) (*types.KRCTokenInfo, error) {
	krcTokenInfo, err := s.cacheClient.KRCTokenInfo(ctx, krcTokenAddr)
	if err == nil {
		return krcTokenInfo, nil
	}
	s.logger.Warn("Cannot get KRC token info from cache, getting from database instead")
	addrInfo, err := s.dbClient.AddressByHash(ctx, krcTokenAddr)
	if err != nil {
		s.logger.Warn("Cannot get KRC token info from db", zap.Error(err))
		return nil, err
	}
	result := &types.KRCTokenInfo{
		Address:     addrInfo.Address,
		TokenName:   addrInfo.TokenName,
		TokenType:   addrInfo.ErcTypes,
		TokenSymbol: addrInfo.TokenSymbol,
		TotalSupply: addrInfo.TotalSupply,
		Decimals:    addrInfo.Decimals,
		Logo:        addrInfo.Logo,
	}
	err = s.cacheClient.UpdateKRCTokenInfo(ctx, result)
	if err != nil {
		s.logger.Warn("Cannot store KRC token info to cache", zap.Error(err))
		return nil, err
	}
	return result, nil
}

func (s *infoServer) getKRCHolder(ctx context.Context, log *types.Log) ([]*types.TokenHolder, error) {
	holdersList := make([]*types.TokenHolder, 2)
	krcTokenInfo, err := s.getKRCTokenInfo(ctx, log.Address)
	if err != nil {
		return nil, err
	}
	if krcTokenInfo.TokenType != "KRC20" {
		return nil, fmt.Errorf("not a KRC20 token")
	}
	if log.Arguments["from"] == "" || log.Arguments["to"] == "" {
		return nil, fmt.Errorf("sender and receiver is not found")
	}
	krcABI, err := s.getSMCAbi(ctx, log)
	if err != nil {
		return nil, err
	}
	fromBalance, err := s.kaiClient.GetKRCBalanceByAddress(ctx, krcABI, common.HexToAddress(log.Address), common.HexToAddress(log.Arguments["from"].(string)))
	if err != nil {
		return nil, err
	}
	toBalance, err := s.kaiClient.GetKRCBalanceByAddress(ctx, krcABI, common.HexToAddress(log.Address), common.HexToAddress(log.Arguments["to"].(string)))
	if err != nil {
		return nil, err
	}
	holdersList[0] = &types.TokenHolder{
		TokenName:       krcTokenInfo.TokenName,
		TokenSymbol:     krcTokenInfo.TokenSymbol,
		TokenDecimals:   krcTokenInfo.Decimals,
		ContractAddress: log.Address,
		HolderAddress:   log.Arguments["from"].(string),
		BalanceString:   fromBalance.String(),
		BalanceFloat:    new(big.Int).Div(fromBalance, new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)).Int64(),
		UpdatedAt:       time.Now().Unix(),
	}
	holdersList[1] = &types.TokenHolder{
		TokenName:       krcTokenInfo.TokenName,
		TokenSymbol:     krcTokenInfo.TokenSymbol,
		TokenDecimals:   krcTokenInfo.Decimals,
		ContractAddress: log.Address,
		HolderAddress:   log.Arguments["to"].(string),
		BalanceString:   toBalance.String(),
		BalanceFloat:    new(big.Int).Div(toBalance, new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)).Int64(),
		UpdatedAt:       time.Now().Unix(),
	}
	return holdersList, nil
}

func (s *infoServer) getInternalTxs(ctx context.Context, log *types.Log) *types.TokenTransfer {
	return &types.TokenTransfer{
		TransactionHash: log.TxHash,
		Contract:        log.Address,
		From:            log.Arguments["from"].(string),
		To:              log.Arguments["to"].(string),
		Value:           log.Arguments["value"].(string),
		Time:            log.Time,
	}
}

func (s *infoServer) insertHistoryTransferKRC(ctx context.Context, smcAddr string) error {
	txs, _, err := s.dbClient.TxsByAddress(ctx, smcAddr, nil)
	fmt.Printf("@@@@@@@@@@@@@@@@@@@@@@ txs: %+v err: %v\n", txs, err)
	if err != nil {
		return err
	}
	for _, tx := range txs {
		if len(tx.Logs) > 0 {
			err = s.storeEvents(ctx, tx.Logs, txs[0].Time)
			if err != nil {
				s.logger.Warn("Cannot store events to db", zap.Error(err))
			}
		}
	}
	return nil
}
