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

	"github.com/kardiachain/explorer-backend/server"
)

// EchoServer define all API expose
type EchoServer interface {
	// General
	Ping(c echo.Context) error
	Info(c echo.Context) error
	Stats(c echo.Context) error

	// Info
	TokenInfo(c echo.Context) error

	// Chart
	TPS(c echo.Context) error

	// Validators
	ValidatorStats(c echo.Context) error
	Validators(c echo.Context) error

	// Blocks
	Blocks(c echo.Context) error
	Block(c echo.Context) error
	BlockExist(c echo.Context) error
	BlockTxs(c echo.Context) error

	// Addresses
	Addresses(c echo.Context) error
	AddressTxs(c echo.Context) error
	AddressHolders(c echo.Context) error
	AddressOwnedTokens(c echo.Context) error
	AddressInternalTxs(c echo.Context) error
	AddressContract(c echo.Context) error
	AddressTxByNonce(c echo.Context) error
	AddressTxHashByNonce(c echo.Context) error

	// Tx
	TxByHash(c echo.Context) error
	TxExist(c echo.Context) error

	// Contracts
	Contracts(c echo.Context) error
}

var echoSrv EchoServer

type restDefinition struct {
	method      string
	path        string
	fn          func(c echo.Context) error
	middlewares []echo.MiddlewareFunc
}

var apis = []restDefinition{
	{
		method:      echo.GET,
		path:        "/ping",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/stats",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	}, {
		method:      echo.GET,
		path:        "/tokens/info",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
	{
		method:      echo.GET,
		path:        "/ping",
		fn:          echoSrv.Ping,
		middlewares: nil,
	},
}

func Start(srv *server.Server) {
	e := echo.New()
	v1Gr := e.Group("api/1")

	echoSrv = srv
	for _, api := range apis {
		v1Gr.Add(api.method, api.path, api.fn, api.middlewares...)
	}

	if err := e.Start(":3000"); err != nil {
		panic("cannot start echo server")
	}
}

type EchoResponse struct {
}
