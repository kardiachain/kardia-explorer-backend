package server

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/api"
	"github.com/kardiachain/explorer-backend/types"
)

type echoServer struct {
	logger *zap.Logger
	info   InfoServer
}

func (s *echoServer) Register(gr *echo.Group) {

	s.registerDashboardService(gr)
	s.registerBlockService(gr)
	s.registerStakingService(gr)
	s.registerAddressService(gr)
	s.registerTransactionService(gr)

	s.registerInternalService(gr)

}

func (s *echoServer) registerDashboardService(gr *echo.Group) {
	dashboardGr := gr.Group("/dashboard")
	dashboardGr.GET("/stats", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		stats, err := s.info.Stats(ctx)
		if err != nil {
			return api.Err(err, c)
		}

		return api.Success(stats, c)
	})

	dashboardGr.GET("/holders", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		holders, err := s.info.TokenHolders(ctx)
		if err != nil {
			return api.Err(err, c)
		}

		return api.Success(holders, c)
	})

	dashboardGr.GET("/token", func(c echo.Context) error {
		// Request may delay, so lets timeout = 10s
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		tokenInfo, err := s.info.TokenInfo(ctx)
		if err != nil {
			return api.Err(err, c)
		}

		return api.Success(tokenInfo, c)
	})
	dashboardGr.GET("/nodes", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		nodes, err := s.info.Nodes(ctx)
		if err != nil {
			return api.Err(err, c)
		}
		return api.Success(nodes, c)
	})
}

func (s *echoServer) registerBlockService(gr *echo.Group) {
	blockGr := gr.Group("/blocks")
	blockGr.GET("", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		pagination := getPagination(c)

		blocks, err := s.info.Blocks(ctx, pagination)
		if err != nil {
			return api.Err(err, c)
		}

		type rBlock struct {
			Height          uint64    `json:"height,omitempty" bson:"height"`
			Time            time.Time `json:"time,omitempty" bson:"time"`
			ProposerAddress string    `json:"proposerAddress,omitempty" bson:"proposerAddress"`
			NumTxs          uint64    `json:"numTxs" bson:"numTxs"`
			GasLimit        uint64    `json:"gasLimit,omitempty" bson:"gasLimit"`
			GasUsed         uint64    `json:"gasUsed" bson:"gasUsed"`
			Rewards         string    `json:"rewards" bson:"rewards"`
		}

		var rBlocks []rBlock
		for _, block := range blocks {
			b := rBlock{
				Height:          block.Height,
				Time:            block.Time,
				ProposerAddress: block.ProposerAddress,
				NumTxs:          block.NumTxs,
				GasLimit:        block.GasLimit,
				GasUsed:         block.GasUsed,
				Rewards:         block.Rewards,
			}
			rBlocks = append(rBlocks, b)
		}
		latestBlock, err := s.info.LatestBlockHeight(ctx)
		if err != nil {
			return api.Err(err, c)
		}
		pagination.Total = latestBlock

		return api.Pagination(pagination, blocks, c)
	})
	blockGr.GET("/:block", func(c echo.Context) error {
		blockInfo := c.Param("block")
		if strings.HasPrefix(blockInfo, "0x") {
			return s.blockByHash(c, blockInfo)
		}

		blockHeight, err := strconv.ParseUint(blockInfo, 10, 64)
		if err != nil || blockHeight <= 0 {
			return api.Invalid.Build(c)
		}

		return s.blockByHeight(c, blockHeight)
	})
	blockGr.GET("/:block/txs", func(c echo.Context) error {
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		blockInfo := c.Param("block")
		pagination := getPagination(c)

		var block *types.Block

		if strings.HasPrefix(blockInfo, "0x") {
			block, err = s.info.BlockByHash(ctx, blockInfo)
			if err != nil {
				return api.Err(err, c)
			}
		} else {
			blockHeight, err := strconv.ParseUint(blockInfo, 10, 64)
			if err != nil || blockHeight <= 0 {
				return api.Invalid.Build(c)
			}

			block, err = s.info.BlockByHeight(ctx, blockHeight)
			if err != nil {
				return api.Err(err, c)
			}
		}

		if block == nil {
			return api.Err(errors.New("invalid block"), c)
		}

		txs, err := s.info.BlockTxs(ctx, block, pagination)
		if err != nil {
			return api.Err(err, c)
		}
		pagination.Total = block.NumTxs

		type rTx struct {
			Hash        string    `json:"hash" bson:"hash"`
			BlockNumber uint64    `json:"blockNumber" bson:"blockNumber"`
			Time        time.Time `json:"time" bson:"time"`
			From        string    `json:"from" bson:"from"`
			To          string    `json:"to" bson:"to"`
			Value       string    `json:"value" bson:"value"`
			TxFee       string    `json:"txFee"`
			Status      uint      `json:"status" bson:"status"`
		}

		var rTxs []*rTx
		for _, tx := range txs {
			t := &rTx{
				Hash:        tx.Hash,
				BlockNumber: tx.BlockNumber,
				Time:        tx.Time,
				From:        tx.From,
				To:          tx.To,
				Value:       tx.Value,
				TxFee:       tx.TxFee,
				Status:      tx.Status,
			}
			rTxs = append(rTxs, t)
		}

		return api.Pagination(pagination, rTxs, c)
	})
	blockGr.GET("/proposer/:address", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		address := c.Param("address")
		pagination := getPagination(c)

		blocks, total, err := s.info.BlocksByProposer(ctx, address, pagination)
		if err != nil {
			return api.Err(err, c)
		}

		pagination.Total = total
		return api.Pagination(pagination, blocks, c)
	})
}

func (s *echoServer) registerStakingService(gr *echo.Group) {
	delegatorGr := gr.Group("/delegators")
	delegatorGr.GET("/:address/validators", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		address := c.Param("address")
		validators, err := s.info.ValidatorsOfDelegator(ctx, address)
		if err != nil {
			return api.Err(err, c)
		}

		return api.Success(validators, c)
	})

	validatorGr := gr.Group("/validators")
	validatorGr.GET("", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		validators, err := s.info.Validators(ctx)
		if err != nil {
			return api.Err(err, c)
		}

		return api.Success(validators, c)

	})
	validatorGr.GET("/candidates", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		valsList, err := s.info.CandidatesList(ctx)
		if err != nil {
			return api.Err(err, c)
		}
		return api.Success(valsList, c)

	})
	validatorGr.GET("/:address", func(c echo.Context) error {
		ctx := context.Background()
		pagination := getPagination(c)
		address := c.Param("address")

		validators, err := s.info.ValidatorStats(ctx, address, pagination)
		if err != nil {
			return api.Err(err, c)
		}

		return api.Success(validators, c)
	})
	validatorGr.GET("/:address/slash", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		address := c.Param("address")

		events, err := s.info.SlashEvents(ctx, address)
		if err != nil {
			return api.Err(err, c)
		}

		return api.Success(events, c)
	})

}

func (s *echoServer) registerAddressService(gr *echo.Group) {
	addressGr := gr.Group("/addresses")
	addressGr.GET("/:address/txs", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		address := c.Param("address")

		pagination := getPagination(c)

		txs, err := s.info.AddressTxs(ctx, address, pagination)

		type rTx struct {
			Hash        string    `json:"hash" bson:"hash"`
			BlockNumber uint64    `json:"blockNumber" bson:"blockNumber"`
			Time        time.Time `json:"time" bson:"time"`
			From        string    `json:"from" bson:"from"`
			To          string    `json:"to" bson:"to"`
			Value       string    `json:"value" bson:"value"`
			TxFee       string    `json:"txFee"`
			Status      uint      `json:"status" bson:"status"`
		}

		totalTxs, err := s.info.TotalTxsOfAddress(ctx, address)
		if err != nil {
			return api.Err(err, c)
		}

		pagination.Total = totalTxs

		var result []rTx
		for _, tx := range txs {
			t := rTx{
				Hash:        tx.Hash,
				BlockNumber: tx.BlockNumber,
				Time:        tx.Time,
				From:        tx.From,
				To:          tx.To,
				Value:       tx.Value,
				TxFee:       tx.TxFee,
				Status:      tx.Status,
			}
			result = append(result, t)
		}

		return api.Pagination(pagination, result, c)
	})
	addressGr.GET("/:address/balance", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		address := c.Param("address")

		balance, err := s.info.Balance(ctx, address)
		if err != nil {
			return api.Err(err, c)
		}

		return api.Success(balance, c)
	})

}

func (s *echoServer) registerTransactionService(gr *echo.Group) {
	txGr := gr.Group("/txs")
	txGr.GET("", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var err error
		pagination := getPagination(c)

		txs, err := s.info.Txs(ctx, pagination)
		if err != nil {
			return api.Err(err, c)
		}

		type rTx struct {
			Hash        string    `json:"hash" bson:"hash"`
			BlockNumber uint64    `json:"blockNumber" bson:"blockNumber"`
			Time        time.Time `json:"time" bson:"time"`
			From        string    `json:"from" bson:"from"`
			To          string    `json:"to" bson:"to"`
			Value       string    `json:"value" bson:"value"`
			TxFee       string    `json:"txFee"`
			Status      uint      `json:"status" bson:"status"`
		}

		var result []rTx
		for _, tx := range txs {
			t := rTx{
				Hash:        tx.Hash,
				BlockNumber: tx.BlockNumber,
				Time:        tx.Time,
				From:        tx.From,
				To:          tx.To,
				Value:       tx.Value,
				TxFee:       tx.TxFee,
				Status:      tx.Status,
			}
			result = append(result, t)
		}

		totalTxs, err := s.info.TotalTxs(ctx)
		if err != nil {
			return api.Err(err, c)
		}

		pagination.Total = totalTxs

		return api.Pagination(pagination, result, c)
	})
	txGr.GET("/:txHash", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		txHash := c.Param("txHash")
		if txHash == "" {
			return api.Invalid.Build(c)
		}

		tx, err := s.info.TxByHash(ctx, txHash)
		if err != nil {
			return api.Err(err, c)
		}

		return api.Success(tx, c)
	})
}

func (s *echoServer) registerInternalService(gr *echo.Group) {
	internalGr := gr.Group("/internal")
	internalGr.POST("/", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		m := struct {
			CirculatingSupply int64 `json:"circulatingSupply"`
		}{}
		if err := c.Bind(&m); err != nil {
			return api.Invalid.Build(c)
		}

		if err := s.info.UpdateCirculatingSupply(ctx, m.CirculatingSupply); err != nil {
			return api.Err(err, c)
		}

		return api.Success(nil, c)
	})
}

func NewEchoServer(cfg Config) (api.EchoServer, error) {
	infoServer, err := NewInfoServer(cfg)
	if err != nil {
		return nil, err

	}
	return &echoServer{
		info: infoServer,
	}, nil
}

func (s *echoServer) blockByHash(c echo.Context, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	block, err := s.info.BlockByHash(ctx, hash)
	if err != nil {
		return api.Err(err, c)
	}

	return api.Success(block, c)
}

func (s *echoServer) blockByHeight(c echo.Context, height uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	block, err := s.info.BlockByHeight(ctx, height)
	if err != nil {
		return api.Err(err, c)
	}

	return api.Success(block, c)
}

func getPagination(c echo.Context) *types.Pagination {
	var page, limit int
	var err error
	pageParams := c.QueryParam("page")
	limitParams := c.QueryParam("limit")
	page, err = strconv.Atoi(pageParams)
	if err != nil {
		page = 0
	}
	limit, err = strconv.Atoi(limitParams)
	if err != nil {
		limit = 20
	}
	pagination := &types.Pagination{
		Skip:  page * limit,
		Limit: limit,
	}
	pagination.Sanitize()
	return pagination
}
