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

	"github.com/kardiachain/kardia-explorer-backend/cfg"
)

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
		{
			method:      echo.POST,
			path:        "/validators/reload",
			fn:          srv.ReloadValidators,
			middlewares: nil,
		},
	}
	bindContractAPIs(gr, srv)
	for _, api := range apis {
		gr.Add(api.method, api.path, api.fn, api.middlewares...)
	}
}

func bindContractAPIs(gr *echo.Group, srv EchoServer) {
	apis := []restDefinition{
		{
			method:      echo.POST,
			path:        "/contracts",
			fn:          srv.InsertContract,
			middlewares: nil,
		},
		{
			method:      echo.PUT,
			path:        "/contracts",
			fn:          srv.UpdateContract,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/contracts",
			fn:          srv.Contracts,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/contracts/:contractAddress",
			fn:          srv.Contract,
			middlewares: nil,
		},
		{
			method:      echo.PUT,
			path:        "/contracts/abi",
			fn:          srv.UpdateSMCABIByType,
			middlewares: nil,
		},
		{
			method: echo.GET,
			// Query params: ?page=0&limit=10&contractAddress=0x&methodName=0x
			path:        "/contracts/events",
			fn:          srv.ContractEvents,
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
