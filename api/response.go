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
// Package api
package api

import (
	"net/http"

	"github.com/labstack/echo"
)

var (
	OK             = EchoResponse{StatusCode: http.StatusOK, Code: 1000, Msg: "Success"}
	InternalServer = EchoResponse{StatusCode: http.StatusInternalServerError, Code: 1100, Msg: "Server busy..."}
	Invalid        = EchoResponse{StatusCode: http.StatusBadRequest, Code: 1101, Msg: "Bad request"}
	Unauthorized   = EchoResponse{StatusCode: http.StatusUnauthorized, Code: 401, Msg: "Unauthorized"}
)

type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type EchoResponse struct {
	StatusCode int         `json:"-"`
	Code       int         `json:"code"`
	Msg        string      `json:"msg"`
	Data       interface{} `json:"data,omitempty"`
}

func (r *EchoResponse) SetData(data interface{}) *EchoResponse {
	r.Data = data
	return r
}

func (r *EchoResponse) Build(c echo.Context) error {
	return c.JSON(r.StatusCode, r)
}
