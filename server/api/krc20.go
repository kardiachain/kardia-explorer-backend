// Package api
package api

import (
	"context"

	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

type IKrc20 interface {
	KRC20Holders(c echo.Context) error
}

func bindKRC20APIs(gr *echo.Group, srv RestServer) {
	apis := []restDefinition{
		{
			method: echo.GET,
			// Query params
			// [?status=(Verified, Unverified)]
			path:        "/krc20/:contractAddress/holders",
			fn:          srv.KRC20Holders,
			middlewares: nil,
		},
	}
	for _, api := range apis {
		gr.Add(api.method, api.path, api.fn, api.middlewares...)
	}
}

func (s *Server) KRC20Holders(c echo.Context) error {
	ctx := context.Background()
	var (
		page, limit int
		err         error
	)
	pagination, page, limit := getPagingOption(c)
	filterCrit := &types.HolderFilter{
		Pagination:      pagination,
		ContractAddress: c.Param("contractAddress"),
	}
	holders, total, err := s.dbClient.GetListHolders(ctx, filterCrit)
	if err != nil {
		s.logger.Warn("Cannot get events from db", zap.Error(err))
	}
	tokenInfo, _ := s.getTokenInfo(ctx, c.Param("contractAddress"))
	if tokenInfo != nil {
		for i := range holders {
			holders[i].Logo = tokenInfo.Logo
			holders[i].TokenDecimals = tokenInfo.Decimals
			// remove redundant field
			holders[i].TokenName = ""
			holders[i].TokenSymbol = ""
			holders[i].Logo = ""
			holders[i].ContractAddress = ""
			// add address names
			holderInfo, _ := s.getAddressDetail(ctx, holders[i].HolderAddress)
			if holderInfo != nil {
				holders[i].HolderName = holderInfo.Name
			}
		}
	}
	return OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  holders,
	}).Build(c)
}
