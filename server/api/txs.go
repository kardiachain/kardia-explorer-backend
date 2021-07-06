// Package api
package api

import (
	"context"
	"fmt"
	"math/big"
	"time"

	kClient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/go-kardia/lib/abi"
	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

type ITx interface {
	Txs(c echo.Context) error
	TxByHash(c echo.Context) error
	SearchAddressByName(c echo.Context) error
	GetInternalTxs(c echo.Context) error
	UpdateInternalTxs(c echo.Context) error
}

func (s *Server) Txs(c echo.Context) error {
	ctx := context.Background()
	pagination, page, limit := getPagingOption(c)
	var (
		err error
		txs []*types.Transaction
	)

	txs, err = s.cacheClient.LatestTransactions(ctx, pagination)
	if err != nil || txs == nil || len(txs) < limit {
		txs, err = s.dbClient.LatestTxs(ctx, pagination)
		if err != nil {
			return Invalid.Build(c)
		}
	}

	smcAddress := s.getValidatorsAddressAndRole(ctx)
	var result Transactions
	for _, tx := range txs {
		t := SimpleTransaction{
			Hash:             tx.Hash,
			BlockNumber:      tx.BlockNumber,
			Time:             tx.Time,
			From:             tx.From,
			To:               tx.To,
			ContractAddress:  tx.ContractAddress,
			Value:            tx.Value,
			TxFee:            tx.TxFee,
			Status:           tx.Status,
			DecodedInputData: tx.DecodedInputData,
			InputData:        tx.InputData,
		}

		if smcAddress[tx.To] != nil {
			t.Role = smcAddress[tx.To].Role
			t.IsInValidatorsList = true
		}
		addrInfo, _ := s.getAddressInfo(ctx, tx.From)
		if addrInfo != nil {
			t.FromName = addrInfo.Name
		}
		addrInfo, _ = s.getAddressInfo(ctx, tx.To)
		if addrInfo != nil {
			t.ToName = addrInfo.Name
		}
		result = append(result, t)
	}

	return OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: s.cacheClient.TotalTxs(ctx),
		Data:  result,
	}).Build(c)
}

func (s *Server) TxByHash(c echo.Context) error {
	lgr := s.logger
	ctx := context.Background()
	txHash := c.Param("txHash")
	if txHash == "" {
		return Invalid.Build(c)
	}

	// Direct decode logs
	tx, err := s.dbClient.TxByHash(ctx, txHash)
	if err != nil {
		// todo: @longnd Review if we can return here
		lgr.Error("cannot get tx from db", zap.Error(err))
		lgr.Info("Try to get transaction from network", zap.String("hash", txHash))
		tx, err = s.kaiClient.GetTransaction(ctx, txHash)
		if err != nil {
			lgr.Error("cannot get transaction from network", zap.Error(err))
			return Invalid.Build(c)
		}
		receipt, err := s.kaiClient.GetTransactionReceipt(ctx, txHash)
		if err != nil {
			lgr.Error("cannot get receipt by hash from network", zap.Error(err))
			return Invalid.Build(c)
		}
		tx.Logs = receipt.Logs
		tx.Root = receipt.Root
		tx.Status = receipt.Status
		tx.GasUsed = receipt.GasUsed
		tx.ContractAddress = receipt.ContractAddress
		tx.TxFee = new(big.Int).Mul(new(big.Int).SetUint64(tx.GasPrice), new(big.Int).SetUint64(receipt.GasUsed)).String()
	}
	/* else { // will be improved later when core blockchain support pending txs API
		if tx.Time.Sub(time.Now()) < 20*time.Second {
			tx.Status = 2 // marked as pending transaction if the duration between now and tx.Time is less than 20 seconds
		} else {
			tx.Status = 0 // marked as failed tx if this tx is submitted for too long
		}
	}*/
	//
	//var toContract *types.Contract
	//if tx.InputData != "0x" {
	//	toContract, _, err = s.dbClient.Contract(ctx, tx.To)
	//	if err != nil {
	//		lgr.Error("cannot get contract from db", zap.Error(err))
	//	}
	//}
	//
	//filter := &types.InternalTxsFilter{
	//	TransactionHash: txHash,
	//}
	var internalTxs []*InternalTransaction
	for id, l := range tx.Logs {
		if l.Topics[0] == cfg.KRCTransferTopic {
			// Get contract details
			start := time.Now()
			iTx := s.buildInternalTransaction(ctx, &tx.Logs[id])
			if iTx != nil {
				fmt.Printf("InternalTxLog:-------%+v\n", iTx.Log)
				fmt.Printf("InternalTxToken:-------%+v\n", iTx.KRCTokenInfo)
				internalTxs = append(internalTxs, iTx)
			}
			fmt.Println("TotalTime", time.Since(start))
		}
	}
	//iTxs, _, err := s.dbClient.GetListInternalTxs(ctx, filter)
	//if err != nil {
	//	s.logger.Warn("Cannot get internal transactions from db", zap.Error(err))
	//}
	//internalTxs := make([]*InternalTransaction, len(iTxs))
	//for i := range iTxs {
	//	logIndex, _ := iTxs[i].LogIndex.(uint)
	//	internalTxs[i] = &InternalTransaction{
	//		Log: &types.Log{
	//			Address:    iTxs[i].Contract,
	//			MethodName: cfg.KRCTransferMethodName,
	//			Arguments: map[string]interface{}{
	//				"from":  iTxs[i].From,
	//				"to":    iTxs[i].To,
	//				"value": iTxs[i].Value,
	//			},
	//			BlockHeight: iTxs[i].BlockHeight,
	//			Time:        iTxs[i].Time,
	//			TxHash:      iTxs[i].TransactionHash,
	//			Index:       logIndex,
	//		},
	//		From:  iTxs[i].From,
	//		To:    iTxs[i].To,
	//		Value: iTxs[i].Value,
	//	}
	//	fromInfo, _ := s.getAddressInfo(ctx, iTxs[i].From)
	//	if fromInfo != nil {
	//		internalTxs[i].FromName = fromInfo.Name
	//	}
	//	toInfo, _ := s.getAddressInfo(ctx, iTxs[i].To)
	//	if toInfo != nil {
	//		internalTxs[i].ToName = toInfo.Name
	//	}
	//	krcTokenInfo, err = s.getKRCTokenInfo(ctx, iTxs[i].Contract)
	//	if err != nil {
	//		s.logger.Info("Cannot get KRC Token Info", zap.String("smcAddress", iTxs[i].Contract), zap.Error(err))
	//		continue
	//	}
	//	internalTxs[i].KRCTokenInfo = krcTokenInfo
	//}

	result := &Transaction{
		BlockHash:        tx.BlockHash,
		BlockNumber:      tx.BlockNumber,
		Hash:             tx.Hash,
		From:             tx.From,
		To:               tx.To,
		Status:           tx.Status,
		ContractAddress:  tx.ContractAddress,
		Value:            tx.Value,
		GasPrice:         tx.GasPrice,
		GasLimit:         tx.GasLimit,
		GasUsed:          tx.GasUsed,
		TxFee:            tx.TxFee,
		Nonce:            tx.Nonce,
		Time:             tx.Time,
		InputData:        tx.InputData,
		DecodedInputData: s.buildFunctionCall(ctx, tx),
		Logs:             internalTxs,
		TransactionIndex: tx.TransactionIndex,
		LogsBloom:        tx.LogsBloom,
		Root:             tx.Root,
	}
	addrInfo, _ := s.getAddressDetail(ctx, tx.From)
	if addrInfo != nil {
		result.FromName = addrInfo.Name
	}
	addrInfo, _ = s.getAddressDetail(ctx, tx.To)
	if addrInfo != nil {
		result.ToName = addrInfo.Name
	}

	//smcAddress := s.getValidatorsAddressAndRole(ctx)
	//if smcAddress[result.To] != nil {
	//	result.Role = smcAddress[result.To].Role
	//	result.IsInValidatorsList = true
	//	return OK.SetData(result).Build(c)
	//}
	if result.Status == 0 {
		txTraceResult, err := s.kaiClient.TraceTransaction(ctx, result.Hash)
		if err != nil {
			s.logger.Warn("Cannot trace tx hash", zap.Error(err), zap.String("txHash", result.Hash))
			return OK.SetData(result).Build(c)
		}
		result.RevertReason = txTraceResult.RevertReason
	}
	return OK.SetData(result).Build(c)
}

func (s *Server) buildFunctionCall(ctx context.Context, tx *types.Transaction) *types.FunctionCall {

	var (
		functionCall *types.FunctionCall
	)

	contractInfo, _, err := s.dbClient.Contract(ctx, tx.To)
	if err != nil {
		return nil
	}

	var contractABI *abi.ABI
	if contractInfo.ABI != "" {
		contractABI, err = s.decodeSMCABIFromBase64(ctx, contractInfo.ABI, contractInfo.Address)
		if err != nil {
			return nil
		}
	} else {
		if contractInfo.Type == cfg.SMCTypeKRC20 {
			contractABI, err = kClient.KRC20ABI()
			if err != nil {
				return nil
			}
		}

		if contractInfo.Type == cfg.SMCTypeKRC721 {
			contractABI, err = kClient.KRC721ABI()
			if err != nil {
				return nil
			}
		}
	}

	if contractABI == nil {
		decoded, err := s.kaiClient.DecodeInputData(tx.To, tx.InputData)
		if err == nil {
			functionCall = decoded
		}
	} else {
		decoded, err := s.kaiClient.DecodeInputWithABI(tx.To, tx.InputData, contractABI)
		if err == nil {
			functionCall = decoded
		}
	}

	return functionCall
}

func (s *Server) buildInternalTransaction(ctx context.Context, l *types.Log) *InternalTransaction {
	lgr := s.logger
	contractInfo, _, err := s.dbClient.Contract(ctx, l.Address)
	if err != nil {
		lgr.Error("cannot get contract from db, skip process log", zap.Error(err))
		return nil
	}

	var contractABI *abi.ABI
	if contractInfo.ABI != "" {
		contractABI, err = s.decodeSMCABIFromBase64(ctx, contractInfo.ABI, contractInfo.Address)
		if err != nil {
			return nil
		}
	} else {
		if contractInfo.Type == cfg.SMCTypeKRC20 {
			contractABI, err = kClient.KRC20ABI()
			if err != nil {
				return nil
			}
		}

		if contractInfo.Type == cfg.SMCTypeKRC721 {
			contractABI, err = kClient.KRC721ABI()
			if err != nil {
				return nil
			}
		}
	}
	var unpackedLog *types.Log
	unpackedLog, err = s.kaiClient.UnpackLog(l, contractABI)
	if err != nil {
		return nil
	}

	fmt.Printf("UnpackedLog: %+v \n", unpackedLog)

	var from, to string
	internalTx := &InternalTransaction{
		Log: unpackedLog,
	}
	fromInfo, _ := s.getAddressDetail(ctx, from)
	if fromInfo != nil {
		internalTx.FromName = fromInfo.Name
	}
	toInfo, _ := s.getAddressDetail(ctx, to)
	if toInfo != nil {
		internalTx.ToName = toInfo.Name
	}
	krcInfo := &types.KRCTokenInfo{
		Address:     contractInfo.Address,
		TokenName:   contractInfo.Name,
		TokenType:   contractInfo.Type,
		TokenSymbol: contractInfo.Symbol,
		TotalSupply: contractInfo.TotalSupply,
		Decimals:    int64(contractInfo.Decimals),
		Logo:        contractInfo.Logo,
		IsVerified:  contractInfo.IsVerified,
	}
	internalTx.KRCTokenInfo = krcInfo

	return internalTx
}

func (s *Server) getAddressDetail(ctx context.Context, address string) (*types.Address, error) {
	lgr := s.logger
	addressDetail := &types.Address{Address: address}
	addrInfo, err := s.cacheClient.AddressInfo(ctx, address)
	if err == nil {
		return addrInfo, nil
	}
	addrInfo, err = s.dbClient.AddressByHash(ctx, address)
	if err != nil {
		lgr.Error("Missing address", zap.Error(err), zap.String("address", address))
		return addressDetail, nil
	}
	err = s.cacheClient.UpdateAddressInfo(ctx, addrInfo)
	if err != nil {
		s.logger.Error("Cannot store address info to cache", zap.Error(err))
	}
	return addrInfo, nil
}
