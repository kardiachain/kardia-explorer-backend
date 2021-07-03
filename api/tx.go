// Package api
package api

import (
	"github.com/labstack/echo"
)

type ITx interface {
	Txs(c echo.Context) error
	TxByHash(c echo.Context) error
	SearchAddressByName(c echo.Context) error
	GetHoldersListByToken(c echo.Context) error
	GetInternalTxs(c echo.Context) error
	UpdateInternalTxs(c echo.Context) error
}
