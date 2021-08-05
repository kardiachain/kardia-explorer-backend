// Package api
package api

import (
	"context"

	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/labstack/echo"
)

type IKrc721 interface {
	KRC721Holders(c echo.Context) error
}

func bindKRC721APIs(gr *echo.Group, srv RestServer) {
	apis := []restDefinition{
		{
			method: echo.GET,
			// Query params
			// [?status=(Verified, Unverified)]
			path:        "/krc721/:contractAddress/holders",
			fn:          srv.KRC721Holders,
			middlewares: nil,
		},
	}
	for _, api := range apis {
		gr.Add(api.method, api.path, api.fn, api.middlewares...)
	}
}

func (s *Server) KRC721Holders(c echo.Context) error {
	ctx := context.Background()
	var (
		page, limit int
		err         error
	)
	contractAddress := c.Param("contractAddress")
	pagination, page, limit := getPagingOption(c)
	holders, total, err := s.dbClient.KRC721Holders(ctx, types.KRC721HolderFilter{
		Pagination:      pagination,
		ContractAddress: contractAddress,
	})
	if err != nil {
		return Invalid.Build(c)
	}
	return OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  holders,
	}).Build(c)
}

func (s *Server) KRC721Inventory(c echo.Context) error {
	ctx := context.Background()
	var (
		page, limit int
		err         error
	)
	contractAddress := c.Param("contractAddress")
	pagination, page, limit := getPagingOption(c)
	holders, total, err := s.dbClient.KRC721Holders(ctx, types.KRC721HolderFilter{
		Pagination:      pagination,
		ContractAddress: contractAddress,
	})
	if err != nil {
		return Invalid.Build(c)
	}
	return OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  holders,
	}).Build(c)
}
