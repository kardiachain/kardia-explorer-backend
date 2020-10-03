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
