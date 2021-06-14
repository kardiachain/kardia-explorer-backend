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

//todo: Improve response structure
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

type EchoR struct {
	c          echo.Context
	StatusCode int         `json:"-"`
	Code       int64       `json:"code"`
	Msg        string      `json:"msg"`
	Data       interface{} `json:"data,omitempty"`
}

func BuildResponse(c echo.Context) EchoR {
	return EchoR{c: c}
}

func (r EchoR) Unauthorized() error {
	r.Code = 401
	r.Msg = "Unauthorized"
	return r.c.JSON(http.StatusUnauthorized, r)
}

func (r EchoR) BadRequest() error {
	r.Code = 400
	r.Msg = "Bad Request"
	return r.c.JSON(http.StatusBadRequest, r)
}

func (r EchoR) NotFound() error {
	r.Code = 401
	r.Msg = "Not Found"
	return r.c.JSON(http.StatusNotFound, r)
}

func (r EchoR) Err(err error) error {
	r.Code = 400
	r.Msg = err.Error()
	return r.c.JSON(http.StatusBadRequest, r)
}

func (r EchoR) OK(data interface{}) error {
	r.Code = 200
	r.Msg = "Success"
	r.Data = data
	return r.c.JSON(http.StatusOK, r)
}
