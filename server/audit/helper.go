// Package audit
package audit

import (
	"context"
	"math/big"
	"time"

	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"go.uber.org/zap"
)

func (s *Server) UpsertBlock(ctx context.Context, block *types.Block) error {
	s.logger.Info("Upsert block:", zap.Uint64("Height", block.Height), zap.Int("Txs length", len(block.Txs)), zap.Int("Receipts length", len(block.Receipts)))
	if err := s.dbClient.DeleteBlockByHeight(ctx, block.Height); err != nil {
		return err
	}
	return s.ImportBlock(ctx, block, false)
}

func (s *Server) filterContracts(ctx context.Context, txs []*types.Transaction) {
	lgr := s.logger
	for _, tx := range txs {
		if tx.ContractAddress != "" {

			c := &types.Contract{
				Address:      tx.ContractAddress,
				Bytecode:     tx.InputData,
				OwnerAddress: tx.From,
				TxHash:       tx.Hash,
				CreatedAt:    tx.Time.Unix(),
				Type:         cfg.SMCTypeNormal,              // Set normal by default
				Status:       types.ContractStatusUnverified, // Unverified by default
				IsVerified:   false,
			}
			lgr.Info("Detect new contract", zap.String("ContractAddress", c.Address), zap.String("TxHash", c.TxHash))
			if err := s.dbClient.InsertContract(ctx, c, nil); err != nil {
				lgr.Error("cannot insert new contract", zap.Error(err))
			}
		}
	}
}

func filterAddrSet(txs []*types.Transaction) map[string]*types.Address {
	addrs := make(map[string]*types.Address)
	for _, tx := range txs {
		addrs[tx.From] = &types.Address{
			Address:    tx.From,
			IsContract: false,
		}
		addrs[tx.To] = &types.Address{
			Address:    tx.To,
			IsContract: false,
		}
		addrs[tx.ContractAddress] = &types.Address{
			Address:    tx.ContractAddress,
			IsContract: true,
		}
	}
	delete(addrs, "")
	delete(addrs, "0x")
	return addrs
}

func (s *Server) getAddressBalances(ctx context.Context, addrs map[string]*types.Address) []*types.Address {
	lgr := s.logger.With(zap.String("method", "getAddressBalance"))
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
		addressesName[v.SmcAddress] = v.Name
	}
	addressesName[cfg.StakingContractAddr] = cfg.StakingContractName

	var (
		code     common.Bytes
		addrsMap = map[string]*types.Address{}
	)
	lgr.Debug("Start process address", zap.Int("TotalAddress", len(addrs)))

	// Spawn go routine

	for addr := range addrs {
		timePerAddress := time.Now()
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
		lgr.Debug("FinishedRefreshAddress", zap.Duration("TotalTime", time.Since(timePerAddress)))
	}
	var result []*types.Address
	for _, info := range addrsMap {
		result = append(result, info)
	}
	return result
}

func (s *Server) mergeAdditionalInfoToTxs(ctx context.Context, txs []*types.Transaction, receipts []*types.Receipt) []*types.Transaction {
	if receipts == nil || len(receipts) == 0 {
		return txs
	}
	receiptIndex := 0
	var (
		gasPrice     *big.Int
		gasUsed      *big.Int
		txFeeInHydro *big.Int
	)

	for id := range txs {
		// todo: Temp remove since we going to decode when get tx details
		//smcABI, err := s.getSMCAbi(ctx, &types.Log{Address: tx.To})
		//if err == nil {
		//	decoded, err := s.kaiClient.DecodeInputWithABI(tx.To, tx.InputData, smcABI)
		//	if err == nil {
		//		tx.DecodedInputData = decoded
		//	}
		//}
		if (receiptIndex > len(receipts)-1) || !(receipts[receiptIndex].TransactionHash == txs[id].Hash) {
			txs[id].Status = 0
			continue
		}

		txs[id].Logs = receipts[receiptIndex].Logs
		txs[id].Root = receipts[receiptIndex].Root
		txs[id].Status = receipts[receiptIndex].Status
		txs[id].GasUsed = receipts[receiptIndex].GasUsed
		txs[id].ContractAddress = receipts[receiptIndex].ContractAddress
		// update txFee
		gasPrice = new(big.Int).SetUint64(txs[id].GasPrice)
		gasUsed = new(big.Int).SetUint64(txs[id].GasUsed)
		txFeeInHydro = new(big.Int).Mul(gasPrice, gasUsed)
		txs[id].TxFee = txFeeInHydro.String()
		receiptIndex++
	}
	return txs
}
