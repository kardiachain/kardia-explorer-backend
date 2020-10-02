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

package server

import (
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

// Server instance kind of a router, which receive request from client (explorer)
// and control how we react those request
type Server struct {
	Logger  *zap.Logger
	infoSrv *infoServer
	apiSrv  *apiServer
}

func (s *Server) Ping(c echo.Context) error {
	panic("implement me")
}

func (s *Server) Info(c echo.Context) error {
	panic("implement me")
}

func (s *Server) Stats(c echo.Context) error {
	panic("implement me")
}

func (s *Server) TokenInfo(c echo.Context) error {
	panic("implement me")
}

func (s *Server) TPS(c echo.Context) error {
	panic("implement me")
}

func (s *Server) ValidatorStats(c echo.Context) error {
	panic("implement me")
}

func (s *Server) Validators(c echo.Context) error {
	panic("implement me")
}

func (s *Server) Blocks(c echo.Context) error {
	panic("implement me")
}

func (s *Server) Block(c echo.Context) error {
	panic("implement me")
}

func (s *Server) BlockExist(c echo.Context) error {
	panic("implement me")
}

func (s *Server) BlockTxs(c echo.Context) error {
	panic("implement me")
}

func (s *Server) Addresses(c echo.Context) error {
	panic("implement me")
}

func (s *Server) AddressTxs(c echo.Context) error {
	panic("implement me")
}

func (s *Server) AddressHolders(c echo.Context) error {
	panic("implement me")
}

func (s *Server) AddressOwnedTokens(c echo.Context) error {
	panic("implement me")
}

func (s *Server) AddressInternalTxs(c echo.Context) error {
	panic("implement me")
}

func (s *Server) AddressContract(c echo.Context) error {
	panic("implement me")
}

func (s *Server) AddressTxByNonce(c echo.Context) error {
	panic("implement me")
}

func (s *Server) AddressTxHashByNonce(c echo.Context) error {
	panic("implement me")
}

func (s *Server) TxByHash(c echo.Context) error {
	panic("implement me")
}

func (s *Server) TxExist(c echo.Context) error {
	panic("implement me")
}

func (s *Server) Contracts(c echo.Context) error {
	panic("implement me")
}

func NewServer() (*Server, error) {
	return &Server{}, nil
}
