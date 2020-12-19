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
	"fmt"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/kardiachain/explorer-backend/cfg"
)

type PrivateAPI interface {
	UpdateCirculatingSupply(c echo.Context) error
}

// EchoServer define all API expose
type EchoServer interface {
	// General
	Stats(c echo.Context) error
	TotalHolders(c echo.Context) error

	// Info
	TokenInfo(c echo.Context) error
	Nodes(c echo.Context) error

	// Validators
	ValidatorStats(c echo.Context) error
	Validators(c echo.Context) error

	// Blocks
	Blocks(c echo.Context) error
	Block(c echo.Context) error
	BlockExist(c echo.Context) error
	BlockTxs(c echo.Context) error
	PersistentErrorBlocks(c echo.Context) error

	// Addresses
	Addresses(c echo.Context) error
	Balance(c echo.Context) error
	AddressTxs(c echo.Context) error
	AddressHolders(c echo.Context) error
	// Tx
	Txs(c echo.Context) error
	TxByHash(c echo.Context) error

	PrivateAPI
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
			path:   "/dashboard/token/circulating",
			fn:     srv.UpdateCirculatingSupply,
		},
		// Blocks
		{
			method: echo.GET,
			// Query params: ?page=1&limit=10
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
			// Params: block's hash
			// Query params: ?page=1&limit=10
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
			// Query params: ?page=1&limit=10
			path:        "/txs",
			fn:          srv.Txs,
			middlewares: []echo.MiddlewareFunc{checkPagination()},
		},
		// Address
		{
			method: echo.GET,
			path:   "/addresses",
			fn:     srv.Addresses,
		},
		{
			method: echo.GET,
			path:   "/addresses/:address/balance",
			fn:     srv.Balance,
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

	fmt.Println("API server", cfg.Port)

	bind(v1Gr, srv)
	if err := e.Start(cfg.Port); err != nil {
		panic("cannot start echo server")
	}
}
