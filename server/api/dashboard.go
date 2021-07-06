// Package api
package api

import (
	"context"

	"github.com/labstack/echo"
)

func (s *Server) Stats(c echo.Context) error {
	ctx := context.Background()
	totalContracts, err := s.cacheClient.TotalContracts(ctx)
	if err != nil {
		return err
	}
	totalAddresses, err := s.cacheClient.TotalAddresses(ctx)
	if err != nil {
		return err
	}

	return OK.SetData(struct {
		TotalHolders   int64 `json:"totalHolders"`
		TotalContracts int64 `json:"totalContracts"`
	}{
		TotalHolders:   totalAddresses,
		TotalContracts: totalContracts,
	}).Build(c)
}

func (s *Server) TotalHolders(c echo.Context) error {
	ctx := context.Background()
	totalHolders, totalContracts := s.cacheClient.TotalHolders(ctx)
	return OK.SetData(struct {
		TotalHolders   uint64 `json:"totalHolders"`
		TotalContracts uint64 `json:"totalContracts"`
	}{
		TotalHolders:   totalHolders,
		TotalContracts: totalContracts,
	}).Build(c)
}
