// Package api
package api

import (
	"context"
	"fmt"

	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/labstack/echo"
)

type IKrc721 interface {
	KRC721Holders(c echo.Context) error
}

func (s *Server) KRC721Holders(c echo.Context) error {
	ctx := context.Background()
	holders, total, err := s.dbClient.KRC721Holders(ctx, types.KRC721HolderFilter{})
	if err != nil {
		return err
	}
	fmt.Println("total", total)
	return OK.SetData(holders).Build(c)
}
