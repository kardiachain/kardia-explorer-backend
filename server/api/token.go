// Package api
package api

import (
	"context"

	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

func (s *Server) GetHoldersListByToken(c echo.Context) error {
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
