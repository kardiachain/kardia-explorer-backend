// Package api
package api

import (
	"github.com/labstack/echo"
)

type IContract interface {
	Contracts(c echo.Context) error
	Contract(c echo.Context) error
	InsertContract(c echo.Context) error
	UpdateContract(c echo.Context) error
	UpdateSMCABIByType(c echo.Context) error
	ContractEvents(c echo.Context) error
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
			method: echo.GET,
			// Query params
			// [?status=(Verified, Unverified)]
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
			// Query params: ?page=0&limit=10&contractAddress=0x&methodName=0x&txHash=0x
			path:        "/contracts/events",
			fn:          srv.ContractEvents,
			middlewares: nil,
		},
	}
	for _, api := range apis {
		gr.Add(api.method, api.path, api.fn, api.middlewares...)
	}
}
