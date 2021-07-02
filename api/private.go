// Package api
package api

import (
	"github.com/labstack/echo"
)

type IPrivate interface {
	// Admin sector
	ReloadAddressesBalance(c echo.Context) error
	ReloadValidators(c echo.Context) error
	UpdateAddressName(c echo.Context) error
	UpsertNetworkNodes(c echo.Context) error
	RemoveNetworkNodes(c echo.Context) error
	UpdateSupplyAmounts(c echo.Context) error
	RemoveDuplicateEvents(c echo.Context) error
	SyncContractInfo(c echo.Context) error
	RefreshKRC20Info(c echo.Context) error
	RefreshKRC721Info(c echo.Context) error
	RefreshContractsInfo(c echo.Context) error
}

func bindPrivateAPIs(gr *echo.Group, srv EchoServer) {
	apis := []restDefinition{
		{
			method:      echo.PUT,
			path:        "/contracts/sync",
			fn:          srv.SyncContractInfo,
			middlewares: nil,
		},
		{
			method:      echo.PUT,
			path:        "/contracts/kcr20/refresh",
			fn:          srv.RefreshKRC20Info,
			middlewares: nil,
		},
		{
			method:      echo.PUT,
			path:        "/contracts/kcr721/refresh",
			fn:          srv.RefreshKRC721Info,
			middlewares: nil,
		},
		{
			method:      echo.PUT,
			path:        "/contracts/refresh",
			fn:          srv.RefreshContractsInfo,
			middlewares: nil,
		},
	}
	for _, api := range apis {
		gr.Add(api.method, api.path, api.fn, api.middlewares...)
	}
}
