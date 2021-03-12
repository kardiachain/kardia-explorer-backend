// Package api
package api

import (
	"github.com/labstack/echo"
)

func bindStakingAPIs(gr *echo.Group, srv EchoServer) {
	apis := []restDefinition{
		//Validator
		{
			method:      echo.GET,
			path:        "/staking/stats",
			fn:          srv.StakingStats,
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
			fn:          srv.Validator,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/delegators/:address/validators",
			fn:          srv.ValidatorsByDelegator,
			middlewares: nil,
		},
		{
			method:      echo.GET,
			path:        "/validators/candidates",
			fn:          srv.Candidates,
			middlewares: nil,
		},
	}
	for _, api := range apis {
		gr.Add(api.method, api.path, api.fn, api.middlewares...)
	}
}
