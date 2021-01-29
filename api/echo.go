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

package api

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/kardiachain/explorer-backend/cfg"
)

// EchoServer define all API expose
type EchoServer interface {
	// General
	Ping(c echo.Context) error
	Stats(c echo.Context) error
	TotalHolders(c echo.Context) error

	// Info
	TokenInfo(c echo.Context) error
	UpdateSupplyAmounts(c echo.Context) error
	Nodes(c echo.Context) error
	UpsertNetworkNodes(c echo.Context) error
	RemoveNetworkNodes(c echo.Context) error

	// Staking-related
	ValidatorStats(c echo.Context) error
	Validators(c echo.Context) error
	GetValidatorsByDelegator(c echo.Context) error
	GetCandidatesList(c echo.Context) error
	GetSlashEvents(c echo.Context) error
	GetSlashedTokens(c echo.Context) error

	// Proposal
	GetProposalsList(c echo.Context) error
	GetProposalDetails(c echo.Context) error
	GetParams(c echo.Context) error

	// Blocks
	Blocks(c echo.Context) error
	Block(c echo.Context) error
	BlockTxs(c echo.Context) error
	BlocksByProposer(c echo.Context) error
	PersistentErrorBlocks(c echo.Context) error

	// Addresses
	Addresses(c echo.Context) error
	AddressInfo(c echo.Context) error
	AddressTxs(c echo.Context) error
	AddressHolders(c echo.Context) error
	ReloadAddressesBalance(c echo.Context) error
	UpdateAddressName(c echo.Context) error

	// Tx
	Txs(c echo.Context) error
	TxByHash(c echo.Context) error
}

type restDefinition struct {
	method      string
	path        string
	fn          func(c echo.Context) error
	middlewares []echo.MiddlewareFunc
}

func bind(gr *echo.Group, srv EchoServer) {
	apis := []restDefinition{
		{
			method:      echo.GET,
			path:        "/ping",
			fn:          srv.Ping,
			middlewares: nil,
		},
		{
			method: echo.GET,
			path:   "/dashboard/stats",
			fn:     srv.Stats,
		},
		{
			method: echo.GET,
			path:   "/dashboard/holders/total",
			fn:     srv.TotalHolders,
		},
		{
			method: echo.GET,
			path:   "/dashboard/token",
			fn:     srv.TokenInfo,
		},
		{
			method: echo.PUT,
			path:   "/dashboard/token/supplies",
			fn:     srv.UpdateSupplyAmounts,
		},
		{
			method: echo.PUT,
			path:   "/nodes",
			fn:     srv.UpsertNetworkNodes,
		},
		{
			method: echo.DELETE,
			path:   "/nodes/:nodeID",
			fn:     srv.RemoveNetworkNodes,
		},
		// Blocks
		{
			method: echo.GET,
			// Query params: ?page=0&limit=10
			path: "/blocks",
			fn:   srv.Blocks,
		},
		{
			method: echo.GET,
			path:   "/blocks/:block",
			fn:     srv.Block,
		},
		{
			method: echo.GET,
			path:   "/blocks/error",
			fn:     srv.PersistentErrorBlocks,
		},
		{
			method: echo.GET,
			// Params: proposer address
			// Query params: ?page=0&limit=10
			path:        "/blocks/proposer/:address",
			fn:          srv.BlocksByProposer,
			middlewares: nil,
		},
		{
			method: echo.GET,
			// Params: block's hash
			// Query params: ?page=0&limit=10
			path:        "/block/:block/txs",
			fn:          srv.BlockTxs,
			middlewares: []echo.MiddlewareFunc{checkPagination()},
		},
		{
			method: echo.GET,
			path:   "/txs/:txHash",
			fn:     srv.TxByHash,
		},
		{
			method: echo.GET,
			// Query params: ?page=0&limit=10
			path:        "/txs",
			fn:          srv.Txs,
			middlewares: []echo.MiddlewareFunc{checkPagination()},
		},
		// Address
		{
			method: echo.GET,
			// Query params: ?page=0&limit=10&sort=1
			path: "/addresses",
			fn:   srv.Addresses,
		},
		{
			method: echo.GET,
			path:   "/addresses/:address",
			fn:     srv.AddressInfo,
		},
		{
			method: echo.POST,
			path:   "/addresses/reload",
			fn:     srv.ReloadAddressesBalance,
		},
		// Tokens
		{
			method:      echo.GET,
			path:        "/addresses/:address/txs",
			fn:          srv.AddressTxs,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/nodes",
			fn:          srv.Nodes,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/validators",
			fn:          srv.Validators,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/validators/:address",
			fn:          srv.ValidatorStats,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/delegators/:address/validators",
			fn:          srv.GetValidatorsByDelegator,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/validators/candidates",
			fn:          srv.GetCandidatesList,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/validators/:address/slash",
			fn:          srv.GetSlashEvents,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/validators/slashed/tokens",
			fn:          srv.GetSlashedTokens,
			middlewares: nil,
		},
		// Proposal
		{
			method: echo.GET,
			// Query params: ?page=0&limit=10
			path:        "/proposal",
			fn:          srv.GetProposalsList,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/proposal/:id",
			fn:          srv.GetProposalDetails,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/proposal/params",
			fn:          srv.GetParams,
			middlewares: nil,
		},
		{
			method:      echo.PUT,
			path:        "/addresses",
			fn:          srv.UpdateAddressName,
			middlewares: nil,
		},
	}

	for _, api := range apis {
		gr.Add(api.method, api.path, api.fn, api.middlewares...)
	}

}

func Start(srv EchoServer, cfg cfg.ExplorerConfig) {
	e := echo.New()

	e.Use(middleware.CORS())
	e.Use(middleware.Logger())
	e.Use(middleware.Gzip())

	v1Gr := e.Group("/api/v1")
	bind(v1Gr, srv)
	if err := e.Start(cfg.Port); err != nil {
		panic("cannot start echo server")
	}
}
